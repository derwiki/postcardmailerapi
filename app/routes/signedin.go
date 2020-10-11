package routes

import (
	"database/sql"
	helpers "github.com/derwiki/postcardmailerapi/app"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
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
	UserID := helpers.GetLoggedInUserID(c, sh.DB)
	if UserID > 0 {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "user_id": UserID})
	}
}

func (sh SignedinHandler) signedinOptionsHandler(c *gin.Context) {
	log.Println("in OPTIONS /v1/signedin")
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
