package routes

import (
	"database/sql"
	"log"
	"net/http"

	helpers "github.com/derwiki/postcardmailerapi/app"
	"github.com/gin-gonic/gin"
)

// SignedinHandler gives this route access to DB
type SignedinHandler struct {
	DB *sql.DB
}

// AddRoutes hooks POST and OPTIONS into the main router
func (sh SignedinHandler) AddRoutes(router gin.IRouter) {
	router.GET("/v1/signedin", sh.signedinGetHandler)
	router.OPTIONS("/v1/signedin", sh.signedinOptionsHandler)
}

func (sh SignedinHandler) signedinGetHandler(c *gin.Context) {
	UserID, ok := helpers.GetLoggedInUserID(c, sh.DB)
	if ok {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "user_id": UserID})
	}
}

func (sh SignedinHandler) signedinOptionsHandler(c *gin.Context) {
	log.Println("in OPTIONS /v1/signedin")
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
