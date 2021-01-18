package routes

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"

	sq "github.com/Masterminds/squirrel"
	helpers "github.com/derwiki/postcardmailerapi/app"
	sharedschemas "github.com/derwiki/postcardmailerapi/app/schemas"
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
	UserID int `form:"user_id"`
}

func (a AddressesHandler) AddressesListGetHandler(c *gin.Context) {
	log.Println("in GET /v1/addresses")

	UserID := helpers.GetLoggedInUserID(c, a.DB)
	if UserID == 0 {
		c.JSON(http.StatusForbidden, gin.H{})
		return
	}

	var addressesListGetSchema AddressesListGetSchema

	err := c.Bind(&addressesListGetSchema)
	if err != nil {
		log.Fatal(err)
	}

	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	sql, args, err := psql.Select("id", "name", "address1", "address2", "city", "state", "postal_code").From("addresses").Where(sq.Eq{"user_id": UserID}).Where(sq.Eq{"deactivated_at": nil}).ToSql()
	if err != nil {
		log.Fatal("AddressesListGetHandler constructing query", err)
		return
	}
	log.Println("sql", sql, "args", args)

	rows, err := a.DB.Query(sql, args...)
	if err != nil {
		log.Fatal("AddressesListGetHandler executing query", err)
		return
	}

	defer rows.Close()
	RespJSON := gin.H{}
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
		RespJSON[idString] = sharedschemas.Address{Name: name, AddressLine1: address1, AddressLine2: address2, City: city, State: state, Zip: postalCode}
		log.Println(idString, name, address1, address2, city, state, postalCode)
	}

	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	c.JSON(http.StatusOK, RespJSON)
}

func (a AddressesHandler) AddressesListOptionsHandler(c *gin.Context) {}
