package routes

import (
	"database/sql"
	"log"
	"net/http"

	sq "github.com/Masterminds/squirrel"
	helpers "github.com/derwiki/postcardmailerapi/app"
	sharedschemas "github.com/derwiki/postcardmailerapi/app/schemas"
	"github.com/gin-gonic/gin"
)

// AddressHandler contains the handler for address related endpoints
type ProfileHandler struct {
	DB *sql.DB
}

func (a ProfileHandler) AddRoutes(router gin.IRouter) {
	router.GET("/v1/profile", a.ProfileGetHandler)
	router.OPTIONS("/v1/profile", a.ProfileOptionsHandler)
}

func (a ProfileHandler) ProfileGetHandler(c *gin.Context) {
	log.Println("in GET /v1/profile")

	UserID, ok := helpers.GetLoggedInUserID(c, a.DB)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{})
		return
	}

	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	sql, args, err := psql.Select("address_id").From("users").Where(sq.Eq{"id": UserID}).ToSql()
	if err != nil {
		log.Fatal("ProfileGetHandler:users constructing query", err)
		return
	}
	log.Println("ProfileGetHandler:users sql", sql, "args", args)
	rows, err := a.DB.Query(sql, args...)
	if err != nil {
		log.Fatal("ProfileGetHandler:users executing query", err)
		return
	}
	defer rows.Close()
	rows.Next()
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("ProfileGetHandler:users sql", sql, "args", args)
	var address_id int
	err = rows.Scan(&address_id)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("ProfileGetHandler:users address_id", address_id)

	sql, args, err = psql.Select("id", "name", "address1", "address2", "city", "state", "postal_code").From("addresses").Where(sq.Eq{"id": address_id}).Where(sq.Eq{"deactivated_at": nil}).ToSql()
	if err != nil {
		log.Fatal("ProfileGetHandler.addresses constructing query", err)
		return
	}
	log.Println("ProfileGetHandler.addresses sql", sql, "args", args)

	rows, err = a.DB.Query(sql, args...)
	if err != nil {
		log.Fatal("ProfileGetHandler.addresses executing query", err)
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

		RespJSON["address"] = sharedschemas.Address{Name: name, AddressLine1: address1, AddressLine2: address2, City: city, State: state, Zip: postalCode}
		log.Println("address", name, address1, address2, city, state, postalCode)
	}

	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	c.JSON(http.StatusOK, RespJSON)
}

func (a ProfileHandler) ProfileOptionsHandler(c *gin.Context) {}
