package routes

import (
	"bytes"
	"encoding/json"
	"fmt"
	helpers "github.com/derwiki/postcardmailerapi/app"
	"github.com/derwiki/postcardmailerapi/app/schemas"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
)

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

func PostcardPreviewPostHandler(c *gin.Context) {
	helpers.SetCorsHeaders(c)
	var responses []string

	var postcardPreviewRequest PostcardPreviewRequestSchema
	err := c.BindJSON(&postcardPreviewRequest)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(postcardPreviewRequest)

	ch := make(chan string)
	concurrencyLevel := len(postcardPreviewRequest.To)

	for i := 0; i < concurrencyLevel; i++ {
		go PreviewPostcardApiRequest(ch, postcardPreviewRequest, postcardPreviewRequest.To[i])
	}

	var respJson = gin.H{}
	for i := 0; i < concurrencyLevel; i++ {
		buffer := <-ch
		fmt.Println("received", i)
		respJson[strconv.Itoa(i)] = buffer
		responses = append(responses, buffer)
	}
	fmt.Println("len(responses)", len(responses))

	c.JSON(200, respJson)
}

func PreviewPostcardApiRequest(ch chan<- string, postcardPreviewRequest PostcardPreviewRequestSchema, to schemas.Address) {
	fmt.Println("PreviewPostcardApiRequest enter")
	BaseUrl := "https://print.directmailers.com/api/v1/postcard/"
	DirectmailApiKey := os.Getenv("DIRECT_MAIL_KEY")

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

	req.Header.Set("Content-Type", `application/json`)
	req.Header.Set("Accept", `application/json`)
	req.Header.Set("Authorization", "Basic "+DirectmailApiKey)
	fmt.Printf("%+v", req)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("err: client.Do: %s", err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("err: ReadAll: %s", err)
	}

	if resp.StatusCode == 422 {
		fmt.Println("status code 422", resp.Status)
		var unprocessableEntity UnprocessableEntityResponseSchema
		json.Unmarshal(body, &unprocessableEntity)
		fmt.Printf("%+v", unprocessableEntity)
		// TODO(derwiki) make sure to handle this case with one user, it happens a lot
	}
	if resp.StatusCode == 200 {
		var directMailPostcardPost DirectMailPostcardPostResponseSchema
		json.Unmarshal(body, &directMailPostcardPost)
		fmt.Printf("%+v", directMailPostcardPost)
		// TODO(derwiki) save to DB
	}
	ch <- string(body)
}

func PostcardPreviewOptionsHandler(c *gin.Context) {
	fmt.Println("in OPTIONS /v1/postcards/preview")
	helpers.SetCorsHeaders(c)
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
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
