package routes

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/derwiki/postcardmailerapi/app/schemas"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"
)

type PostcardHandler struct {
	DB               *sql.DB
	DirectmailApiKey string
}

func (hnd PostcardHandler) AddRoutes(router gin.IRouter) {
	router.POST("/v1/postcard/preview", hnd.PostcardPreviewPostHandler)
	router.OPTIONS("/v1/postcard/preview", hnd.PostcardPreviewOptionsHandler)
}

type PostcardPreviewRequestSchema struct {
	Front  string
	Back   string
	To     []schemas.Address
	From   schemas.Address
	UserId int
}

type PreviewPostcardApiRequestSchema struct {
	Description string
	Size        string
	DryRun      bool
	Front       string
	Back        string
	To          schemas.Address
	From        schemas.Address
}

type UnprocessableEntityErrorResponseSchema struct {
	Message    string
	StatusCode int
}

type UnprocessableEntityResponseSchema struct {
	Error UnprocessableEntityErrorResponseSchema
}

type MyResponse struct {
	StatusCode int
	Body       []byte
	Err        error
}

func (hnd PostcardHandler) PostcardPreviewPostHandler(c *gin.Context) {
	cookie, err := c.Cookie("SessionId")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("cookie", cookie)
	var responses []MyResponse

	var postcardPreviewRequest PostcardPreviewRequestSchema
	err = c.BindJSON(&postcardPreviewRequest)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(postcardPreviewRequest)

	ch := make(chan MyResponse)
	concurrencyLevel := len(postcardPreviewRequest.To)
	// 10 sec timeout.
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// This works, but if goroutine terminates before sending a message
	// to the channel you'll never receive this many messages. Main
	// goroutine will hang.
	// for i := 0; i < concurrencyLevel; i++ {
	//   go	 hnd.PreviewPostcardApiRequest(ch, postcardPreviewRequest, postcardPreviewRequest.To[i])
	// }

	var respJson = gin.H{}
	var Successes []DirectMailPostcardPostResponseSchema
	var Failures []UnprocessableEntityResponseSchema

	done := make(chan struct{})
	go func() {
		defer close(done)
		for my_resp := range ch {
			fmt.Println("received")
			//respJson[strconv.Itoa(i)] = resp
			responses = append(responses, my_resp)

			if my_resp.Err != nil {
				fmt.Printf("err: ReadAll: %s", my_resp.Err)
			} else {

				if my_resp.StatusCode == 422 {
					fmt.Println("status code 422")
					var unprocessableEntity UnprocessableEntityResponseSchema
					json.Unmarshal(my_resp.Body, &unprocessableEntity)
					fmt.Printf("%+v", unprocessableEntity)
					Failures = append(Failures, unprocessableEntity)
					// TODO(derwiki) make sure to handle this case with one user, it happens a lot
				}
				if my_resp.StatusCode == 200 {
					var directMailPostcardPost DirectMailPostcardPostResponseSchema
					json.Unmarshal(my_resp.Body, &directMailPostcardPost)
					fmt.Printf("%+v", directMailPostcardPost)
					Successes = append(Successes, directMailPostcardPost)
					// TODO(derwiki) save to DB
				}
			}
		}
	}()

	wg := sync.WaitGroup{}
	for i := 0; i < concurrencyLevel; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			hnd.PreviewPostcardApiRequest(ctx, ch, postcardPreviewRequest, postcardPreviewRequest.To[index])
		}(i)
	}

	wg.Wait()
	// Stop the processor goroutine
	close(ch)
	// Wait for it to end
	<-done
	fmt.Println("len(responses)", len(responses))
	respJson[`Successes`] = Successes
	respJson[`Failures`] = Failures

	c.JSON(200, respJson)
}

func (hnd PostcardHandler) PreviewPostcardApiRequest(ctx context.Context, ch chan<- MyResponse, postcardPreviewRequest PostcardPreviewRequestSchema, to schemas.Address) {
	fmt.Println("PreviewPostcardApiRequest enter")
	BaseUrl := "https://print.directmailers.com/api/v1/postcard/"

	var previewPostcardApiRequest = PreviewPostcardApiRequestSchema{
		Description: "test",
		Size:        "4.25x6",
		DryRun:      true,
		Front:       postcardPreviewRequest.Front,
		Back:        postcardPreviewRequest.Back,
		To:          to,
		From:        postcardPreviewRequest.From,
	}
	fmt.Printf("%+v", previewPostcardApiRequest)
	jsonValue, _ := json.Marshal(previewPostcardApiRequest)
	fmt.Printf("%+v", jsonValue)

	client := &http.Client{}
	req, err := http.NewRequest("POST", BaseUrl, bytes.NewReader(jsonValue))
	if err != nil {
		fmt.Printf("err: NewRequest: %s", err)
	}
	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", `application/json`)
	req.Header.Set("Accept", `application/json`)
	req.Header.Set("Authorization", "Basic "+hnd.DirectmailApiKey)
	fmt.Printf("%+v", req)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("err: client.Do: %s", err)
		// You have to return here, there is no body
		ch <- MyResponse{Err: err}
		return
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("err: ReadAll: %s", err)
		ch <- MyResponse{Err: err}
		return
	}

	ch <- MyResponse{resp.StatusCode, body, nil}
}

func (hnd PostcardHandler) PostcardPreviewOptionsHandler(c *gin.Context) {
	//fmt.Println("in OPTIONS /v1/postcards/preview")
	//helpers.SetCorsHeaders(c)
	//c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

type ThumbnailsResponseSchema struct {
	Large  string
	Medium string
	Small  string
}
type AddressResponseSchema struct {
	AddressLine1 string
	AddressLine2 string
	City         string
	Name         string
	State        string
	Zip          string
}
type DirectMailPostcardPostResponseSchema struct {
	EstimatedDeliveryDate string
	ActualDeliveryDate    string
	MailingDate           string
	Front                 string
	Back                  string
	BackThumbnails        ThumbnailsResponseSchema
	FrontThumbnails       ThumbnailsResponseSchema
	From                  AddressResponseSchema
	To                    AddressResponseSchema
	Canceled              bool
	Cost                  int
	Created               string
	Description           string
	DryRun                bool
	Medium                string
	PostalCarrier         string
	PostalClass           string
	PrintRecord           string
	RenderedPdf           string
	Size                  string
	Status                string
}
