package routes

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

type DBTestHandler struct {
	DB *sql.DB
}

func (hnd DBTestHandler) AddRoutes(router gin.IRouter) {
	router.POST("/v1/playground/dbtest", hnd.DbTestPostHandler)
	router.OPTIONS("/v1/playground/dbtest", hnd.DbTestOptionsHandler)
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
