package routes_signup

import (
	"fmt"
	helpers "github.com/derwiki/postcardmailerapi/app"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

type SignupPostSchema struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func SignupPostHandler(c *gin.Context) {
	var signupPost SignupPostSchema
	err := c.BindJSON(&signupPost)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(signupPost.Email, signupPost.Password)
	helpers.SetCorsHeaders(c)
	c.JSON(http.StatusOK, gin.H{"status": "ok", "email": signupPost.Email})
}

func SignupOptionsHandler(c *gin.Context) {
	fmt.Println("in OPTIONS /v1/signup")
	helpers.SetCorsHeaders(c)
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
