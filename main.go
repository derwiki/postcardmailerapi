package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	_ "github.com/heroku/x/hmetrics/onload"
	_ "github.com/lib/pq"
	"io/ioutil"
	"log"
	"net/http"
	"os"
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

func dbTest() {
	connStr := os.Getenv("DATABASE_URL")
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	rows, err := db.Query("SELECT id, email FROM users ORDER BY id ASC")

	defer rows.Close()
	for rows.Next() {
		var email string
		var id int

		err = rows.Scan(&id, &email)

		fmt.Println(id, email)
	}
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
	router.POST("/v1/dbtest", func(c *gin.Context) {
		dbTest()
		c.JSON(200, gin.H{"status": "ok"})
	})

	router.POST("/v1/postcard/preview", func(c *gin.Context) {
		var responses []string

		ch := make(chan string)
		concurrencyLevel := 30

		for i := 0; i < concurrencyLevel; i++ {
			go PreviewPostcardApiRequest(ch)
		}

		for i := 0; i < concurrencyLevel; i++ {
			buffer := <-ch
			fmt.Println("received", i)
			responses = append(responses, buffer)
		}
		fmt.Println("len(responses)", len(responses))

		c.JSON(200, gin.H{
			"0": responses[0],
			"1": responses[1],
			"2": responses[2],
			"3": responses[3],
			"4": responses[4],
		})
	})

	router.Run(":" + port)
}
