package routes

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"

	devisecrypto "github.com/consyse/go-devise-encryptor"
	"github.com/gin-gonic/gin"
)

type signinPostSchema struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// SigninHandler gives this route access to DB
type SigninHandler struct {
	DB *sql.DB
}

// AddRoutes hooks POST and OPTIONS into the main router
func (sh SigninHandler) AddRoutes(router gin.IRouter) {
	router.POST("/v1/signin", sh.signinPostHandler)
	router.OPTIONS("/v1/signin", sh.signinOptionsHandler)
}

func (sh SigninHandler) signinPostHandler(c *gin.Context) {
	secure := true
	if os.Getenv("APPLICATION_ENV") == "development" {
		secure = false
	}
	c.SetSameSite(http.SameSiteNoneMode)
	httpOnly := true

	var signinPost signinPostSchema
	err := c.BindJSON(&signinPost)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(signinPost.Email)

	var encryptedPassword string
	var id int
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	sql, args, err := psql.Select("id", "encrypted_password").From("users").Where(sq.Eq{"email": signinPost.Email}).ToSql()
	if err != nil {
		log.Println("signinPostHandler constructing query")
		c.JSON(http.StatusNotFound, gin.H{})
		return
	}
	rows, err := sh.DB.Query(sql, args...)
	if err != nil {
		log.Println("signinPostHandler: performed query: err", err)
	}

	defer rows.Close()
	if rows == nil {
		log.Println("signinPostHandler: no user found for email", signinPost.Email)
		// TODO(derwiki) change this to a better response
		c.JSON(http.StatusNotFound, gin.H{})
		return
	}
	rows.Next()
	err = rows.Scan(&id, &encryptedPassword)
	if err != nil {
		log.Println("signinPostHandler: rows scan: err", err)
	}
	log.Println("id", id, "password", encryptedPassword)

	val := devisecrypto.Compare(signinPost.Password, "", encryptedPassword)
	if val == false {
		log.Println("signinPostHandler: passwords don't match, clearing cookie")
		c.SetCookie("SessionId", "", -3600, "/v1/", "", secure, httpOnly)
		c.JSON(http.StatusForbidden, gin.H{})
		return
	}

	// Create a new random session token
	sessionToken := uuid.New().String()
	now := sq.Expr("NOW()")
	log.Println("signinPostHandler: sessionToken", sessionToken, "now", now)
	upsertSessionQuery := psql.Insert("sessions").Columns("user_id", "session_id", "issued_at", "created_at", "updated_at").
		Values(id, sessionToken, now, now, now).
		Suffix("ON CONFLICT (user_id) DO UPDATE SET session_id = $1, issued_at = NOW(), created_at = NOW(), updated_at = NOW() WHERE sessions.user_id = $2", sessionToken, id)
	rows, err = upsertSessionQuery.RunWith(sh.DB).Query()

	if err != nil {
		log.Println("signinPostHandler: performed query: err", err)
		c.JSON(http.StatusUnprocessableEntity, gin.H{})
	}
	log.Println("rows", rows)
	c.SetCookie("SessionId", sessionToken, 3600, "/v1/", "", secure, httpOnly)
	c.JSON(http.StatusOK, gin.H{"status": "ok", "email": signinPost.Email})
}

func (sh SigninHandler) signinOptionsHandler(c *gin.Context) {
	log.Println("in OPTIONS /v1/signin")
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
