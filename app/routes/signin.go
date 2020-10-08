package routes

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"

	devisecrypto "github.com/consyse/go-devise-encryptor"
	helpers "github.com/derwiki/postcardmailerapi/app"
	"github.com/gin-gonic/gin"
)

type SigninPostSchema struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type SigninHandler struct {
	DB *sql.DB
}

func (sh SigninHandler) AddRoutes(router gin.IRouter) {
	router.POST("/v1/signin", sh.SigninPostHandler)
	router.OPTIONS("/v1/signin", sh.SigninOptionsHandler)
}

func (sh SigninHandler) SigninPostHandler(c *gin.Context) {
	var encrypted_password string
	var id int
	var signinPost SigninPostSchema
	err := c.BindJSON(&signinPost)
	if err != nil {
		log.Fatal(err)
	}
	var SessionId string
	SessionId, err = c.Cookie("SessionId")
	fmt.Println(signinPost.Email, signinPost.Password, SessionId)
	//session := sessions.Default(c)

	email := signinPost.Email
	rows, err := sh.DB.Query("SELECT id, encrypted_password FROM users WHERE email = $1", email)
	if err != nil {
		fmt.Println("DevisePostHandler: performed query: err", err)
	}

	defer rows.Close()
	if rows == nil {
		fmt.Println("DevisePostHandler: no user found for email", email)
		// TODO(derwiki) change this to a better response
		c.JSON(http.StatusOK, gin.H{})
	} else {
		rows.Next()
		err = rows.Scan(&id, &encrypted_password)
		if err != nil {
			fmt.Println("DevisePostHandler: rows scan: err", err)
		}
		fmt.Println("id", id, "password", encrypted_password)
	}

	val := devisecrypto.Compare(signinPost.Password, "", encrypted_password)
	fmt.Println(`Passwords are the same?`, val)
	if val {
		fmt.Println(`Setting logged-in cookie`)
		// Create a new random session token
		sessionToken := uuid.New().String()
		fmt.Println("Created sessionToken", sessionToken)
		now := time.Now()
		rows, err := sh.DB.Query("INSERT INTO sessions (user_id, session_id, issued_at, created_at, updated_at) VALUES ($1, $2, $3, $4, $5)", id, sessionToken, now, now, now)
		if err != nil {
			fmt.Println("DevisePostHandler: performed query: err", err)
		}
		fmt.Println("rows", rows)

		c.SetSameSite(http.SameSiteNoneMode)
		c.SetCookie("SessionId", sessionToken, 3600, "/", "", true, true)
		// session.Set("SessionId", sessionToken)

	}
	helpers.SetCorsHeaders(c)
	c.JSON(http.StatusOK, gin.H{"status": "ok", "email": signinPost.Email})
}

func (a SigninHandler) SigninOptionsHandler(c *gin.Context) {
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
