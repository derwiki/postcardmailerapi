package main

import (
	"database/sql"
	"fmt"
	helpers "github.com/derwiki/postcardmailerapi/app"
	routes "github.com/derwiki/postcardmailerapi/app/routes"
	"github.com/gin-gonic/gin"
	_ "github.com/heroku/x/hmetrics/onload"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"os"
)

type SignupPost struct {
	Email    string `json:"email"`
	Password string `json:"password"`
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

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		log.Fatal("$PORT must be set")
	}

	router := gin.New()
	router.Use(gin.Logger())

	router.POST("/v1/signup", routes.SignupPostHandler)
	router.OPTIONS("/v1/signup", routes.SignupOptionsHandler)
	router.POST("/v1/postcard/preview", routes.PreviewPostHandler)

	router.POST("/v1/dbtest", func(c *gin.Context) {
		helpers.SetCorsHeaders(c)
		dbTest()
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	router.OPTIONS("/v1/dbtest", func(c *gin.Context) {
		fmt.Println("in OPTIONS /v1/dbtest")
		helpers.SetCorsHeaders(c)
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	router.Run(":" + port)
}
