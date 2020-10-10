package routes

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"

	helpers "github.com/derwiki/postcardmailerapi/app"
	"github.com/derwiki/postcardmailerapi/app/schemas"
	"github.com/gin-gonic/gin"
)

type PostcardHandler struct {
	DB               *sql.DB
	DirectmailApiKey string
}

func (hnd PostcardHandler) AddRoutes(router gin.IRouter) {
	router.POST("/v1/postcard/preview", hnd.PostcardPreviewPostHandler)
	router.OPTIONS("/v1/postcard/preview", hnd.PostcardPreviewOptionsHandler)
	router.POST("/v1/postcard/send", hnd.PostcardSendPostHandler)
	router.OPTIONS("/v1/postcard/send", hnd.PostcardPreviewOptionsHandler)
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

func (hnd PostcardHandler) PostcardPostHandler(c *gin.Context, dryrun bool) {
	var responses []MyResponse
	UserID := helpers.GetLoggedInUserID(c, hnd.DB)
	log.Println("PostcardPreviewPostHandler: UserID", UserID, "dryrun", dryrun)

	var postcardPreviewRequest PostcardPreviewRequestSchema
	err := c.BindJSON(&postcardPreviewRequest)
	if err != nil {
		log.Fatal(err)
	}
	// fmt.Println(postcardPreviewRequest)

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

	var Response = gin.H{}
	var Successes []DirectMailPostcardPostResponseSchema
	var Failures []UnprocessableEntityResponseSchema

	done := make(chan struct{})
	go func() {
		defer close(done)
		for DirectMailAPIResponse := range ch {
			// fmt.Println("received")
			//respJson[strconv.Itoa(i)] = resp
			responses = append(responses, DirectMailAPIResponse)

			if DirectMailAPIResponse.Err != nil {
				log.Printf("err: ReadAll: %s", DirectMailAPIResponse.Err)
			} else {

				if DirectMailAPIResponse.StatusCode == 422 {
					log.Println("status code 422")
					var unprocessableEntity UnprocessableEntityResponseSchema
					json.Unmarshal(DirectMailAPIResponse.Body, &unprocessableEntity)
					log.Printf("%+v", unprocessableEntity)
					Failures = append(Failures, unprocessableEntity)
					// TODO(derwiki) make sure to handle this case with one user, it happens a lot
				}
				if DirectMailAPIResponse.StatusCode == 200 {
					var directMailPostcardPost DirectMailPostcardPostResponseSchema
					json.Unmarshal(DirectMailAPIResponse.Body, &directMailPostcardPost)
					// fmt.Printf("%+v", directMailPostcardPost)
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
			hnd.PreviewPostcardApiRequest(ctx, ch, postcardPreviewRequest, postcardPreviewRequest.To[index], dryrun, UserID)
		}(i)
	}

	wg.Wait()
	// Stop the processor goroutine
	close(ch)
	// Wait for it to end
	<-done
	// fmt.Println("len(responses)", len(responses))
	Response[`Successes`] = Successes
	Response[`Failures`] = Failures

	c.JSON(200, Response)
}

func (hnd PostcardHandler) PreviewPostcardApiRequest(ctx context.Context, ch chan<- MyResponse, postcardPreviewRequest PostcardPreviewRequestSchema, to schemas.Address, dryrun bool, UserID int) {
	log.Println("PreviewPostcardApiRequest enter")
	BaseUrl := "https://print.directmailers.com/api/v1/postcard/"

	var previewPostcardApiRequest = PreviewPostcardApiRequestSchema{
		Description: "Testing from Golang",
		Size:        "4.25x6",
		DryRun:      dryrun,
		Front:       postcardPreviewRequest.Front,
		Back:        postcardPreviewRequest.Back,
		To:          to,
		From:        postcardPreviewRequest.From,
	}
	// fmt.Printf("%+v", previewPostcardApiRequest)
	jsonValue, _ := json.Marshal(previewPostcardApiRequest)
	// fmt.Printf("%+v", jsonValue)

	client := &http.Client{}
	req, err := http.NewRequest("POST", BaseUrl, bytes.NewReader(jsonValue))
	if err != nil {
		log.Printf("err: NewRequest: %s", err)
	}
	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", `application/json`)
	req.Header.Set("Accept", `application/json`)
	req.Header.Set("Authorization", "Basic "+hnd.DirectmailApiKey)
	// fmt.Printf("%+v", req)

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("err: client.Do: %s", err)
		ch <- MyResponse{Err: err}
		return
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("err: ReadAll: %s", err)
		ch <- MyResponse{Err: err}
		return
	}

	log.Println("About to save, UserID", UserID)
	// TODO(derwiki) Save to database here

	ch <- MyResponse{resp.StatusCode, body, nil}
}

func (hnd PostcardHandler) PostcardPreviewPostHandler(c *gin.Context) {
	PostcardHandler.PostcardPostHandler(hnd, c, true)
}
func (hnd PostcardHandler) PostcardSendPostHandler(c *gin.Context) {
	PostcardHandler.PostcardPostHandler(hnd, c, false)
}

// middleware takes care of this
func (hnd PostcardHandler) PostcardPreviewOptionsHandler(c *gin.Context) {}

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
