package database

import (
	"database/sql"
	// "log"
)

var config struct {
	Database *sql.DB
}

type QueryBlob struct {
	Query string
	Args  []interface{}
}

func InitialiseDatabase(database *sql.DB) {
	config.Database = database
}

func LogQuery(query string) {
	// log.Println(query)
}

func QueryRow(query string, args ...interface{}) *sql.Row {
	LogQuery(query)
	return config.Database.QueryRow(query, args...)
}

func Query(query string, args ...interface{}) (*sql.Rows, error) {
	LogQuery(query)
	return config.Database.Query(query, args...)
}

func Exec(query string, args ...interface{}) (sql.Result, error) {
	LogQuery(query)
	return config.Database.Exec(query, args...)
}

func Begin() (*sql.Tx, error) {
	return config.Database.Begin()
}
