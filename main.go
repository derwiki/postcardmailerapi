package main

import (
	routes "github.com/derwiki/postcardmailerapi/app/routes"
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

	router.POST("/v1/signup", routes.SignupPostHandler)
	router.OPTIONS("/v1/signup", routes.SignupOptionsHandler)
	router.POST("/v1/postcard/preview", routes.PostcardPreviewHandler)
	router.OPTIONS("/v1/postcard/preview", routes.PreviewPostOptionsHandler)
	router.POST("/v1/playground/dbtest", routes.DbTestPostHandler)
	router.OPTIONS("/v1/playground/dbtest", routes.DbTestOptionsHandler)

	router.Run(":" + port)
}
