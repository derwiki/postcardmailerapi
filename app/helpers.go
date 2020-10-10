package helpers

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	sq "github.com/Masterminds/squirrel"

	"github.com/gin-gonic/gin"
)

func SetCorsHeaders(c *gin.Context) {
	var origin = c.GetHeader("Origin")
	log.Println("SetCorsHeaders: Origin", origin)
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
	var IssuedAt time.Time
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	sql, args, err := psql.Select("user_id", "issued_at").From("sessions").Where(sq.Eq{"session_id": SessionID}).ToSql()
	if err != nil {
		log.Println("GetLoggedInUserID constructing query")
		log.Fatal(err)
		return 0
	}
	rows, err := DB.Query(sql, args...)

	if err != nil {
		log.Println("GetLoggedInUserID: performed query: err", err)
		return 0
	}

	defer rows.Close()
	rows.Next()
	err = rows.Scan(&UserID, &IssuedAt)
	if err != nil {
		log.Println("GetLoggedInUserID: no user found for SessionID", SessionID)
		c.JSON(http.StatusForbidden, gin.H{})
		return 0
	}
	err = rows.Err()
	if err != nil {
		log.Println("GetLoggedInUserID: row.Err", err)
		return 0
	}
	log.Println("UserID", UserID, "IssuedAt", IssuedAt)

	return UserID
}
