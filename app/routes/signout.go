package routes

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	sq "github.com/Masterminds/squirrel"
	helpers "github.com/derwiki/postcardmailerapi/app"

	"github.com/gin-gonic/gin"
)

type signoutPostSchema struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// SignoutHandler gives this route access to DB
type SignoutHandler struct {
	DB *sql.DB
}

// AddRoutes hooks POST and OPTIONS into the main router
func (sh SignoutHandler) AddRoutes(router gin.IRouter) {
	router.POST("/v1/signout", sh.signoutPostHandler)
	router.OPTIONS("/v1/signout", sh.signoutOptionsHandler)
}

func (sh SignoutHandler) signoutPostHandler(c *gin.Context) {
	secure := true
	if os.Getenv("APPLICATION_ENV") == "development" {
		secure = false
	}

	UserID := helpers.GetLoggedInUserID(c, sh.DB)

	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	sql, args, err := psql.Delete("").From("sessions").Where(sq.Eq{"user_id": UserID}).ToSql()
	if err != nil {
		log.Println("signoutPostHandler constructing query")
		c.JSON(http.StatusNotFound, gin.H{})
		return
	}
	rows, err := sh.DB.Query(sql, args...)
	if err != nil {
		log.Println("signoutPostHandler: performed query: err", err)
		c.JSON(http.StatusNotFound, gin.H{})
		return
	}

	defer rows.Close()
	if rows == nil {
		log.Println("signoutPostHandler: no session found for UserID", UserID)
		// TODO(derwiki) change this to a better response
		c.JSON(http.StatusNotFound, gin.H{})
		return
	}
	log.Println("rows", rows)
	c.SetSameSite(http.SameSiteNoneMode)
	httpOnly := true
	c.SetCookie("SessionId", "", -1, "/v1/", "", secure, httpOnly)
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (sh SignoutHandler) signoutOptionsHandler(c *gin.Context) {
	log.Println("in OPTIONS /v1/signout")
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
