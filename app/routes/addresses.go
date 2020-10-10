package routes

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"

	sq "github.com/Masterminds/squirrel"
	helpers "github.com/derwiki/postcardmailerapi/app"
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
	log.Println("in GET /v1/addresses")

	// BEGIN checking authentication cookie
	UserID := helpers.GetLoggedInUserID(c, a.DB)
	if UserID == 0 {
		log.Println("not logged in")
		return
	}
	log.Println("already logged in")
	// TODO(derwiki) make sure issued_at is recent 2 hours (or whatever)
	// END checking authentication cookie

	var addressesListGetSchema AddressesListGetSchema

	err := c.Bind(&addressesListGetSchema)
	if err != nil {
		log.Fatal(err)
	}

	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	sql, args, err := psql.Select("id", "name", "address1", "address2", "city", "state", "postal_code").From("addresses").Where(sq.Eq{"user_id": UserID}).ToSql()
	if err != nil {
		log.Println("AddressesListGetHandler constructing query")
		// You should return error here
		log.Fatal(err)
	}
	log.Println("sql", sql)
	log.Println("args", args)

	rows, err := a.DB.Query(sql, args...)
	if err != nil {
		log.Println("AddressesListGetHandler executing query")
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
		log.Println(idString, name, address1, address2, city, state, postalCode)
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
