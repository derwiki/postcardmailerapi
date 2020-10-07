package routes

import (
	"fmt"
	"log"
	"net/http"
	"time"

	helpers "github.com/derwiki/postcardmailerapi/app"
	"github.com/gin-gonic/gin"
)

type SigninPostSchema struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func SigninPostHandler(c *gin.Context) {
	var signinPost SigninPostSchema
	err := c.BindJSON(&signinPost)
	if err != nil {
		log.Fatal(err)
	}
	var SessionId string
	SessionId, err = c.Cookie("SessionId")
	fmt.Println(signinPost.Email, signinPost.Password, SessionId)
	c.SetCookie("SessionId", "26d64bcb-b428-4508-b629-522cb4b01b0e", 3600, "/", "", true, true)
	c.SetSameSite(http.SameSiteNoneMode)
	helpers.SetCorsHeaders(c)
	c.JSON(http.StatusOK, gin.H{"status": "ok", "email": signinPost.Email})
}

func SigninOptionsHandler(c *gin.Context) {
	fmt.Println("in OPTIONS /v1/signin")
	helpers.SetCorsHeaders(c)
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// addCookie will apply a new cookie to the response of a http request
// with the key/value specified.
func addCookie(w http.ResponseWriter, name, value string, ttl time.Duration) {
	expire := time.Now().Add(ttl)
	cookie := http.Cookie{
		Name:    name,
		Value:   value,
		Expires: expire,
	}
	http.SetCookie(w, &cookie)
}
