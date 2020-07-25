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

type SignupPost struct {
	EMAIL    string `json:"email"`
	PASSWORD string `json:"password"`
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

func setCorsHeaders(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Headers", "access-control-allow-origin, access-control-allow-headers")
	c.Header("Access-Control-Allow-Methods", "POST, OPTIONS")
}

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		log.Fatal("$PORT must be set")
	}

	router := gin.New()
	router.Use(gin.Logger())

	router.POST("/v1/dbtest", func(c *gin.Context) {
		setCorsHeaders(c)
		dbTest()
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	router.OPTIONS("/v1/dbtest", func(c *gin.Context) {
		fmt.Println("in OPTIONS /v1/dbtest")
		setCorsHeaders(c)
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	router.POST("/v1/signup", func(c *gin.Context) {
		fmt.Println("in POST /v1/signup")
		var signupPost SignupPost
		err := c.BindJSON(&signupPost)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(signupPost.EMAIL, signupPost.PASSWORD)
		fmt.Println(c.Params)
		setCorsHeaders(c)
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	router.OPTIONS("/v1/signup", func(c *gin.Context) {
		fmt.Println("in OPTIONS /v1/signup")
		setCorsHeaders(c)
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	router.POST("/v1/postcard/preview", func(c *gin.Context) {
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
	})

	router.Run(":" + port)
}
