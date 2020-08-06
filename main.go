package main

import (
	"database/sql"

	helpers "github.com/derwiki/postcardmailerapi/app"
	"github.com/derwiki/postcardmailerapi/app/routes"
	"github.com/gin-gonic/gin"
	_ "github.com/heroku/x/hmetrics/onload"
	_ "github.com/lib/pq"
	"log"
	"os"
)

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		log.Fatal("$PORT must be set")
	}

	router := gin.New()
	router.Use(gin.Logger())
	// Consider doing this, and removing all calls to SetCorsHeaders
	router.Use(func(c *gin.Context) {
		helpers.SetCorsHeaders(c)
		c.Next()
	})

	connStr := os.Getenv("DATABASE_URL")
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Println("opening db")
		log.Fatal(err)
	}
	addressHandler := routes.AddressesHandler{DB: db}
	addressHandler.AddRoutes(router)
	dbTestHandler := routes.DBTestHandler{DB: db}
	dbTestHandler.AddRoutes(router)
	postcardHandler := routes.PostcardHandler{DB: db, DirectmailApiKey: os.Getenv("DIRECT_MAIL_KEY")}
	postcardHandler.AddRoutes(router)

	// Similar handler for SignupPost
	router.POST("/v1/signup", routes.SignupPostHandler)
	router.OPTIONS("/v1/signup", routes.SignupOptionsHandler)

	router.Run(":" + port)
}
