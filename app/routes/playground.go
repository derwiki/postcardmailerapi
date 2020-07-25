package routes

import (
	"database/sql"
	"fmt"
	helpers "github.com/derwiki/postcardmailerapi/app"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
)

func DbTestPostHandler(c *gin.Context) {
	helpers.SetCorsHeaders(c)
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
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func DbTestOptionsHandler(c *gin.Context) {
	fmt.Println("in OPTIONS /v1/playground/dbtest")
	helpers.SetCorsHeaders(c)
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
