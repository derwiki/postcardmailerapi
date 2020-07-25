package routes

import (
	"bytes"
	"encoding/json"
	"fmt"
	helpers "github.com/derwiki/postcardmailerapi/app"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
)

type Address struct {
	Name         string
	AddressLine1 string
	AddressLine2 string
	City         string
	State        string
	Zip          string
}

type PreviewPost struct {
	Description string
	Size        string
	DryRun      bool
	Front       string
	Back        string
	To          Address
	From        Address
}

func PreviewPostHandler(c *gin.Context) {
	helpers.SetCorsHeaders(c)
	var responses []string

	ch := make(chan string)
	concurrencyLevel := 3

	for i := 0; i < concurrencyLevel; i++ {
		go PreviewPostcardApiRequest(ch)
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

func PreviewPostcardApiRequest(ch chan<- string) {
	fmt.Println("PreviewPostcardApiRequest enter")
	BaseUrl := "https://print.directmailers.com/api/v1/postcard/"
	DirectmailApiKey := os.Getenv("DIRECT_MAIL_KEY")

	var previewPost = PreviewPost{
		Description: "test",
		Size:        "4.25x6",
		DryRun:      true,
		Front:       "<html><body>Front</body></html>",
		Back:        "<html><body>Back</body></html>",
		To: Address{
			Name:         "Adam Derewecki",
			AddressLine1: "960 Wisconsin St",
			AddressLine2: "",
			City:         "San Francisco",
			State:        "CA",
			Zip:          "94107",
		},
		From: Address{
			Name:         "Adam Derewecki",
			AddressLine1: "960 Wisconsin St",
			AddressLine2: "",
			City:         "San Francisco",
			State:        "CA",
			Zip:          "94107",
		},
	}
	fmt.Printf("%+v", previewPost)
	jsonValue, _ := json.Marshal(previewPost)
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
	ch <- string(body)
}

func PreviewPostOptionsHandler(c *gin.Context) {
	fmt.Println("in OPTIONS /v1/postcards/preview")
	helpers.SetCorsHeaders(c)
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
