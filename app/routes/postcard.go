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
	sharedschemas "github.com/derwiki/postcardmailerapi/app/schemas"
	"github.com/gin-gonic/gin"
)

type PostcardHandler struct {
	DB               *sql.DB
	DirectMailAPIKey string
}

func (hnd PostcardHandler) AddRoutes(router gin.IRouter) {
	router.POST("/v1/postcard/preview", hnd.postcardPreviewPostHandler)
	router.OPTIONS("/v1/postcard/preview", hnd.postcardOptionsHandler)
	router.POST("/v1/postcard/send", hnd.postcardSendPostHandler)
	router.OPTIONS("/v1/postcard/send", hnd.postcardOptionsHandler)
}

type postcardPreviewRequestSchema struct {
	Front  string
	Back   string
	To     []sharedschemas.Address
	From   sharedschemas.Address
	UserId int
}

type previewPostcardAPIRequestSchema struct {
	Description string
	Size        string
	DryRun      bool
	Front       string
	Back        string
	To          sharedschemas.Address
	From        sharedschemas.Address
}

type unprocessableEntityErrorResponseSchema struct {
	Message    string
	StatusCode int
}

type unprocessableEntityResponseSchema struct {
	Error unprocessableEntityErrorResponseSchema
}

type directMailResponse struct {
	StatusCode int
	Body       []byte
	Err        error
}

func (hnd PostcardHandler) postcardPostHandler(c *gin.Context, dryrun bool) {
	var responses []directMailResponse
	UserID := helpers.GetLoggedInUserID(c, hnd.DB)
	log.Println("PostcardPreviewPostHandler: UserID", UserID, "dryrun", dryrun)

	var postcardPreviewRequest postcardPreviewRequestSchema
	err := c.BindJSON(&postcardPreviewRequest)
	if err != nil {
		log.Fatal(err)
	}

	ch := make(chan directMailResponse)
	concurrencyLevel := len(postcardPreviewRequest.To)
	timeOut := 10 * time.Second
	ctx, cancel := context.WithTimeout(c.Request.Context(), timeOut)
	defer cancel()

	// This works, but if goroutine terminates before sending a message
	// to the channel you'll never receive this many messages. Main
	// goroutine will hang.
	// for i := 0; i < concurrencyLevel; i++ {
	//   go	 hnd.previewPostcardAPIRequest(ch, postcardPreviewRequest, postcardPreviewRequest.To[i])
	// }

	var Response = gin.H{}
	var Successes []directMailPostcardPostResponseSchema
	var Failures []unprocessableEntityResponseSchema

	done := make(chan struct{})
	go func() {
		defer close(done)
		for DirectMailAPIResponse := range ch {
			responses = append(responses, DirectMailAPIResponse)

			if DirectMailAPIResponse.Err != nil {
				log.Printf("err: ReadAll: %s", DirectMailAPIResponse.Err)
			} else {

				if DirectMailAPIResponse.StatusCode == 422 {
					log.Println("status code 422")
					var unprocessableEntity unprocessableEntityResponseSchema
					json.Unmarshal(DirectMailAPIResponse.Body, &unprocessableEntity)
					log.Printf("%+v", unprocessableEntity)
					Failures = append(Failures, unprocessableEntity)
					// TODO(derwiki) make sure to handle this case with one user, it happens a lot
				}
				if DirectMailAPIResponse.StatusCode == 200 {
					var directMailPostcardPost directMailPostcardPostResponseSchema
					json.Unmarshal(DirectMailAPIResponse.Body, &directMailPostcardPost)
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
			hnd.previewPostcardAPIRequest(ctx, ch, postcardPreviewRequest, postcardPreviewRequest.To[index], dryrun, UserID)
		}(i)
	}

	wg.Wait()
	// Stop the processor goroutine
	close(ch)
	// Wait for it to end
	<-done
	Response[`Successes`] = Successes
	Response[`Failures`] = Failures

	c.JSON(200, Response)
}

func (hnd PostcardHandler) previewPostcardAPIRequest(ctx context.Context, ch chan<- directMailResponse, postcardPreviewRequest postcardPreviewRequestSchema, to sharedschemas.Address, dryrun bool, UserID int) {
	log.Println("previewPostcardAPIRequest enter")
	BaseURL := "https://print.directmailers.com/api/v1/postcard/"

	jsonValue, err := json.Marshal(
		previewPostcardAPIRequestSchema{
			Description: "Testing from Golang",
			Size:        "4.25x6",
			DryRun:      dryrun,
			Front:       postcardPreviewRequest.Front,
			Back:        postcardPreviewRequest.Back,
			To:          to,
			From:        postcardPreviewRequest.From,
		})

	if err != nil {
		log.Fatal(err)
		return
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", BaseURL, bytes.NewReader(jsonValue))
	if err != nil {
		log.Printf("err: NewRequest: %s", err)
	}
	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", `application/json`)
	req.Header.Set("Accept", `application/json`)
	req.Header.Set("Authorization", "Basic "+hnd.DirectMailAPIKey)

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("err: client.Do: %s", err)
		ch <- directMailResponse{Err: err}
		return
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("err: ReadAll: %s", err)
		ch <- directMailResponse{Err: err}
		return
	}

	log.Println("About to save, UserID", UserID)
	// TODO(derwiki) Save to database here

	ch <- directMailResponse{resp.StatusCode, body, nil}
}

func (hnd PostcardHandler) postcardPreviewPostHandler(c *gin.Context) {
	PostcardHandler.postcardPostHandler(hnd, c, true)
}
func (hnd PostcardHandler) postcardSendPostHandler(c *gin.Context) {
	PostcardHandler.postcardPostHandler(hnd, c, false)
}

func (hnd PostcardHandler) postcardOptionsHandler(c *gin.Context) {}

type thumbnailsResponseSchema struct {
	Large  string
	Medium string
	Small  string
}

type addressResponseSchema struct {
	AddressLine1 string
	AddressLine2 string
	City         string
	Name         string
	State        string
	Zip          string
}
type directMailPostcardPostResponseSchema struct {
	EstimatedDeliveryDate string
	ActualDeliveryDate    string
	MailingDate           string
	Front                 string
	Back                  string
	BackThumbnails        thumbnailsResponseSchema
	FrontThumbnails       thumbnailsResponseSchema
	From                  addressResponseSchema
	To                    addressResponseSchema
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
