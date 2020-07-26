package routes

import (
	"database/sql"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	helpers "github.com/derwiki/postcardmailerapi/app"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
	"strconv"
)

type AddressesListGetSchema struct {
	UserId int `form:"user_id"`
}

func AddressesListGetHandler(c *gin.Context) {
	fmt.Println("in GET /v1/addresses")
	var addressesListGetSchema AddressesListGetSchema

	err := c.Bind(&addressesListGetSchema)
	if err != nil {
		log.Fatal(err)
	}

	UserId := addressesListGetSchema.UserId

	helpers.SetCorsHeaders(c)

	connStr := os.Getenv("DATABASE_URL")
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Println("opening db")
		log.Fatal(err)
	}

	// StmtCache caches Prepared Stmts for you
	// dbCache := sq.NewStmtCacher(db)

	// StatementBuilder keeps your syntax neat
	// mydb := sq.StatementBuilder.RunWith(dbCache)
	// select_users := mydb.Select("*").From("users")

	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	sql, args, err := psql.Select("id", "name", "address1", "address2", "city", "state", "postal_code").From("addresses").Where(sq.Eq{"user_id": UserId}).ToSql()
	if err != nil {
		log.Println("constructing query")
		log.Fatal(err)
	}
	log.Println("sql", sql)
	log.Println("args", args)

	rows, err := db.Query(sql, args...)
	if err != nil {
		log.Println("executing query")
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

func AddressesListOptionsHandler(c *gin.Context) {
	fmt.Println("in OPTIONS /v1/addresses")
	helpers.SetCorsHeaders(c)
	c.JSON(http.StatusOK, gin.H{})
}
