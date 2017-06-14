package database

import (
	"database/sql"
	"fmt"
	"github.com/gorilla/context"
	"net/http"
)

var config struct {
	Database *sql.DB
}

func InitialiseDatabase(database *sql.DB) {
	config.Database = database
}

func logQuery(query string) {
	fmt.Println(query)
}

func QueryRow(query string, args ...interface{}) *sql.Row {
	logQuery(query)
	return config.Database.QueryRow(query, args...)
}

func Query(query string, args ...interface{}) (*sql.Rows, error) {
	logQuery(query)
	return config.Database.Query(query, args...)
}

func GetRequestQueryCount(r *http.Request) int {
	current := context.Get(r, "queries")
	if current != nil {
		currentInt, ok := current.(int)
		if !ok {
			panic("Unable to convert query count to integer")
		}
		return currentInt
	}
	return 0
}

func incrementRequestQueryCount(r *http.Request) {
	queries := GetRequestQueryCount(r)
	queries += 1
	context.Set(r, "queries", queries)
}

func RequestQueryRow(r *http.Request, query string, args ...interface{}) *sql.Row {
	incrementRequestQueryCount(r)
	return QueryRow(query, args...)
}

func RequestQuery(r *http.Request, query string, args ...interface{}) (*sql.Rows, error) {
	incrementRequestQueryCount(r)
	return Query(query, args...)
}
