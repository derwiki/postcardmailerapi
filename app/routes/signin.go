package routes

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"

	devisecrypto "github.com/consyse/go-devise-encryptor"
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
	secure := true
	if os.Getenv("APPLICATION_ENV") == "development" {
		secure = false
	}
	c.SetSameSite(http.SameSiteNoneMode)
	httpOnly := true
	var encryptedPassword string
	var id int
	var signinPost SigninPostSchema
	err := c.BindJSON(&signinPost)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(signinPost.Email, signinPost.Password)

	email := signinPost.Email
	// rows, err := sh.DB.Query("SELECT id, encrypted_password FROM users WHERE email = $1", email)

	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	sql, args, err := psql.Select("id", "encrypted_password").From("users").Where(sq.Eq{"email": email}).ToSql()
	if err != nil {
		log.Println("SigninPostHandler constructing query")
		c.JSON(http.StatusNotFound, gin.H{})
		return
	}
	rows, err := sh.DB.Query(sql, args...)
	if err != nil {
		fmt.Println("SigninPostHandler: performed query: err", err)
	}

	defer rows.Close()
	if rows == nil {
		fmt.Println("SigninPostHandler: no user found for email", email)
		// TODO(derwiki) change this to a better response
		c.JSON(http.StatusNotFound, gin.H{})
		return
	}
	rows.Next()
	err = rows.Scan(&id, &encryptedPassword)
	if err != nil {
		fmt.Println("SigninPostHandler: rows scan: err", err)
	}
	fmt.Println("id", id, "password", encryptedPassword)

	val := devisecrypto.Compare(signinPost.Password, "", encryptedPassword)
	if val == false {
		fmt.Println("SigninPostHandler: passwords don't match, clearing cookie")
		c.SetCookie("SessionId", "", -3600, "/v1/", "", secure, httpOnly)
		c.JSON(http.StatusForbidden, gin.H{})
		return
	}

	// Create a new random session token
	sessionToken := uuid.New().String()
	fmt.Println("SigninPostHandler: sessionToken", sessionToken)
	now := time.Now()
	rows, err = sh.DB.Query(`
		INSERT INTO sessions (user_id, session_id, issued_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id)
		DO UPDATE SET session_id = $2, issued_at = $3, updated_at = $5
		WHERE sessions.user_id = $1
		`, id, sessionToken, now, now, now)
	if err != nil {
		fmt.Println("SigninPostHandler: performed query: err", err)
		c.JSON(http.StatusUnprocessableEntity, gin.H{})
	}
	fmt.Println("rows", rows)
	c.SetCookie("SessionId", sessionToken, 3600, "/v1/", "", secure, httpOnly)
	c.JSON(http.StatusOK, gin.H{"status": "ok", "email": signinPost.Email})
}

func (a SigninHandler) SigninOptionsHandler(c *gin.Context) {
	fmt.Println("in OPTIONS /v1/signin")
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
