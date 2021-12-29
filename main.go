package main

import (
	"database/sql"

	"log"
	"os"

	helpers "github.com/derwiki/postcardmailerapi/app"
	"github.com/derwiki/postcardmailerapi/app/routes"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	_ "github.com/heroku/x/hmetrics/onload"
	_ "github.com/lib/pq"
)

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		log.Fatal("$PORT must be set")
	}

	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.New()
	router.Use(gin.Logger())

	router.Use(func(c *gin.Context) {
		helpers.SetCorsHeaders(c)
		c.Next()
	})
	store := cookie.NewStore([]byte("secret"))
	router.Use(sessions.Sessions("mysession", store))

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
	postcardHandler := routes.PostcardHandler{DB: db, DirectMailAPIKey: os.Getenv("DIRECT_MAIL_KEY")}
	postcardHandler.AddRoutes(router)
	signinHandler := routes.SigninHandler{DB: db}
	signinHandler.AddRoutes(router)
	signoutHandler := routes.SignoutHandler{DB: db}
	signoutHandler.AddRoutes(router)
	signedinHandler := routes.SignedinHandler{DB: db}
	signedinHandler.AddRoutes(router)
	profileHandler := routes.ProfileHandler{DB: db}
	profileHandler.AddRoutes(router)

	// Similar handler for SignupPost
	router.POST("/v1/signup", routes.SignupPostHandler)
	router.OPTIONS("/v1/signup", routes.SignupOptionsHandler)

	router.Run(":" + port)
}
