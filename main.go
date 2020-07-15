package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	_ "github.com/heroku/x/hmetrics/onload"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type Address struct {
	Name string `json:"Name"`
	AddressLine1 string `json:"AddressLine1"`
	AddressLine2 string `json:"AddressLine2"`
	City string `json:"City"`
	State string `json:"State"`
	Zip string `json:"Zip"`
}

type PreviewPost struct {
	Description  string    `json:"Description"`
	Size         string    `json:"Size"`
	DryRun       bool      `json:"DryRun"`
	Front        string    `json:"Front"`
	Back         string    `json:"Back"`
	To           Address   `json:"To"`
	From         Address   `json:"From"`
}

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		log.Fatal("$PORT must be set")
	}

	router := gin.New()
	router.Use(gin.Logger())
	router.LoadHTMLGlob("templates/*.tmpl.html")
	router.Static("/static", "static")

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl.html", nil)
	})

	router.POST("/v1/postcard/preview", func(c *gin.Context) {
		BaseUrl := "https://print.directmailers.com/api/v1/postcard/"
		DirectmailApiKey := os.Getenv("DIRECT_MAIL_KEY")

		var previewPost = PreviewPost{
			Description: "test",
			Size:        "4.25x6",
			DryRun:      true,
			Front:       "<html><body>Front</body></html>",
			Back:       "<html><body>Back</body></html>",
			To: Address{
				Name: "Adam Derewecki",
				AddressLine1: "960 Wisconsin St",
				AddressLine2: "",
				City: "San Francisco",
				State: "CA",
				Zip: "94107",
			},
			From: Address{
				Name: "Adam Derewecki",
				AddressLine1: "960 Wisconsin St",
				AddressLine2: "",
				City: "San Francisco",
				State: "CA",
				Zip: "94107",
			},
		}
		fmt.Printf("%+v", previewPost)
		jsonValue, _ := json.Marshal(previewPost)
		fmt.Printf("%+v", jsonValue)

		client := &http.Client{}
		req, err := http.NewRequest("POST", BaseUrl, bytes.NewReader(jsonValue))
		if err != nil {
		}

		req.Header.Set("Content-Type", `application/json`)
		req.Header.Set("Accept", `application/json`)
		req.Header.Set("Authorization", "Basic " + DirectmailApiKey)
		fmt.Printf("%+v", req)

		resp, err := client.Do(req)

		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)

		if err != nil {
			c.JSON(500, gin.H{
				"status": "ERR",
				"err": err,
			})
		} else {
			c.JSON(200, gin.H{
				"status": "OK",
				"body":   body,
			})
		}
	})

	router.Run(":" + port)
}
