package routes

import (
	"database/sql"
	"log"
	"net/http"

	devisecrypto "github.com/consyse/go-devise-encryptor"
	"github.com/gin-gonic/gin"
)

type DBTestHandler struct {
	DB *sql.DB
}

func (hnd DBTestHandler) AddRoutes(router gin.IRouter) {
	router.POST("/v1/playground/dbtest", hnd.DbTestPostHandler)
	router.OPTIONS("/v1/playground/dbtest", hnd.DbTestOptionsHandler)
	router.POST("/v1/playground/devise", hnd.DevisePostHandler)
}

func (hnd DBTestHandler) DbTestPostHandler(c *gin.Context) {
	//helpers.SetCorsHeaders(c)
	//connStr := os.Getenv("DATABASE_URL")
	//db, err := sql.Open("postgres", connStr)
	//if err != nil {
	//	log.Fatal(err)
	//}

	rows, err := hnd.DB.Query("SELECT id, email FROM users ORDER BY id ASC")

	defer rows.Close()
	for rows.Next() {
		var email string
		var id int

		err = rows.Scan(&id, &email)

		log.Println(id, email)
	}
	// err was not used
	if err != nil {
		log.Fatal(err)
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (hnd DBTestHandler) DbTestOptionsHandler(c *gin.Context) {
	//fmt.Println("in OPTIONS /v1/playground/dbtest")
	//helpers.SetCorsHeaders(c)
	//c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (hnd DBTestHandler) DevisePostHandler(c *gin.Context) {
	var encrypted_password string
	var id int
	email := "pc@derwiki.net"
	log.Println("in DevisePostHandler")
	rows, err := hnd.DB.Query("SELECT id, encrypted_password FROM users WHERE email = $1", email)
	log.Println("in DevisePostHandler: performed query")
	if err != nil {
		log.Println("in DevisePostHandler: performed query: err", err)
	}

	defer rows.Close()
	log.Println("in DevisePostHandler: performed Close", rows)
	if rows == nil {
		log.Println("in DevisePostHandler: performed Close, rows is nil")
		c.JSON(http.StatusOK, gin.H{})
	} else {
		rows.Next()
		log.Println("in DevisePostHandler: in first Next()")

		err = rows.Scan(&id, &encrypted_password)

		log.Println("id", id, "password", encrypted_password)
	}

	pepper := ""

	newPassword := "password1"
	val := devisecrypto.Compare(newPassword, pepper, encrypted_password)
	log.Println(`Passwords are the same?`, val)

	newPassword = "password"
	val = devisecrypto.Compare(newPassword, pepper, encrypted_password)
	log.Println(`Passwords are the same?`, val)
}
