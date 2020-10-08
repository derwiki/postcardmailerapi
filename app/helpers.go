package helpers

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func SetCorsHeaders(c *gin.Context) {
	var origin = c.GetHeader("Origin")
	fmt.Println("origin", origin)
	c.Header("Access-Control-Allow-Origin", origin)
	c.Header("Access-Control-Allow-Headers", "access-control-allow-origin, access-control-allow-headers")
	c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	c.Header("Access-Control-Allow-Credentials", "true")
	c.Header("Content-Type", `application/json`)
	c.Header("Accept", `application/json`)
}

func GetLoggedInUserID(c *gin.Context, DB *sql.DB) int {
	SessionID, err := c.Cookie("SessionId")
	if err != nil {
		if err == http.ErrNoCookie {
			c.JSON(http.StatusUnauthorized, gin.H{"status": "unauthorized"})
			return 0
		}
		log.Fatal(err)
	}

	var UserID int
	var issued_at time.Time
	rows, err := DB.Query("SELECT user_id, issued_at FROM sessions WHERE session_id = $1", SessionID)
	if err != nil {
		fmt.Println("AddressesListGetHandler: performed query: err", err)
	}

	defer rows.Close()
	rows.Next()
	err = rows.Scan(&UserID, &issued_at)
	if err != nil {
		fmt.Println("AddressesListGetHandler: no user found for SessionID", SessionID)
		c.JSON(http.StatusForbidden, gin.H{})
		return 0
	}
	err = rows.Err()
	if err != nil {
		fmt.Println("AddressesListGetHandler: row.Err", rows.Err())
	}
	fmt.Println("UserID", UserID, "issued_at", issued_at)
	return UserID
}
