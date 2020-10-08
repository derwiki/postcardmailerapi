package routes

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/gin-gonic/gin"
)

// AddressHandler contains the handler for address related endpoints
type AddressesHandler struct {
	DB *sql.DB
}

func (a AddressesHandler) AddRoutes(router gin.IRouter) {
	router.GET("/v1/addresses", a.AddressesListGetHandler)
	router.OPTIONS("/v1/addresses", a.AddressesListOptionsHandler)
}

type AddressesListGetSchema struct {
	UserId int `form:"user_id"`
}

func (a AddressesHandler) AddressesListGetHandler(c *gin.Context) {
	fmt.Println("in GET /v1/addresses")

	SessionID, err := c.Cookie("SessionId")
	if err != nil {
		if err == http.ErrNoCookie {
			c.JSON(http.StatusUnauthorized, gin.H{"status": "unauthorized"})
			return
		}
		log.Fatal(err)
	}

	var user_id int
	var issued_at time.Time
	rows, err := a.DB.Query("SELECT user_id, issued_at FROM sessions WHERE session_id = $1", SessionID)
	if err != nil {
		fmt.Println("DevisePostHandler: performed query: err", err)
	}

	defer rows.Close()
	if rows == nil {
		fmt.Println("DevisePostHandler: no user found for SessionID", SessionID)
		// TODO(derwiki) change this to a better response
		c.JSON(http.StatusNotFound, gin.H{})
		return
	} else {
		rows.Next()
		err = rows.Scan(&user_id, &issued_at)
		if err != nil {
			fmt.Println("DevisePostHandler: rows scan: err", err)
			c.JSON(http.StatusForbidden, gin.H{})
			return
		}
		fmt.Println("user_id", user_id, "issued_at", issued_at)
	}

	var addressesListGetSchema AddressesListGetSchema

	err = c.Bind(&addressesListGetSchema)
	if err != nil {
		log.Fatal(err)
	}

	UserId := user_id

	// Done in middleware
	// helpers.SetCorsHeaders(c)

	// StmtCache caches Prepared Stmts for you
	// dbCache := sq.NewStmtCacher(db)

	// StatementBuilder keeps your syntax neat
	// mydb := sq.StatementBuilder.RunWith(dbCache)
	// select_users := mydb.Select("*").From("users")

	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	sql, args, err := psql.Select("id", "name", "address1", "address2", "city", "state", "postal_code").From("addresses").Where(sq.Eq{"user_id": UserId}).ToSql()
	if err != nil {
		log.Println("constructing query")
		// You should return error here
		log.Fatal(err)
	}
	log.Println("sql", sql)
	log.Println("args", args)

	rows, err = a.DB.Query(sql, args...)
	if err != nil {
		log.Println("executing query")
		// You should return error here
		log.Fatal(err)
	}

	defer rows.Close()
	respJson := gin.H{}
	for rows.Next() {
		var id int
		var name string
		var address1 string
		var address2 string
		var city string
		var state string
		var postalCode string

		err = rows.Scan(&id, &name, &address1, &address2, &city, &state, &postalCode)
		if err != nil {
			log.Fatal(err)
		}

		idString := strconv.Itoa(id)
		respJson[idString] = name
		fmt.Println(idString, name, address1, address2, city, state, postalCode)
	}

	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	c.JSON(http.StatusOK, respJson)
}

func (a AddressesHandler) AddressesListOptionsHandler(c *gin.Context) {
	// Done in middleware
	//	fmt.Println("in OPTIONS /v1/addresses")
	//	helpers.SetCorsHeaders(c)
	//c.JSON(http.StatusOK, gin.H{})
}
