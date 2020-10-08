package routes

import (
	"database/sql"
	"fmt"
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

		fmt.Println(id, email)
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
	fmt.Println("in DevisePostHandler")
	rows, err := hnd.DB.Query("SELECT id, encrypted_password FROM users WHERE email = $1", email)
	fmt.Println("in DevisePostHandler: performed query")
	if err != nil {
		fmt.Println("in DevisePostHandler: performed query: err", err)
	}

	defer rows.Close()
	fmt.Println("in DevisePostHandler: performed Close", rows)
	if rows == nil {
		fmt.Println("in DevisePostHandler: performed Close, rows is nil")
		c.JSON(http.StatusOK, gin.H{})
	} else {
		rows.Next()
		fmt.Println("in DevisePostHandler: in first Next()")

		err = rows.Scan(&id, &encrypted_password)

		fmt.Println("id", id, "password", encrypted_password)
	}

	pepper := ""

	/*
		hashedPassword, err := devisecrypto.Digest(password, stretches, pepper)
		if err != nil {
			panic(err)
		}
		fmt.Println("hashedPassword: ", hashedPassword)

		// and to compare with a previously hashed password
	*/

	newPassword := "password1"
	val := devisecrypto.Compare(newPassword, pepper, encrypted_password)
	fmt.Println(`Passwords are the same?`, val)

	newPassword = "password"
	val = devisecrypto.Compare(newPassword, pepper, encrypted_password)
	fmt.Println(`Passwords are the same?`, val)
}
