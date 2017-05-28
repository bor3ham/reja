package database

import (
  "database/sql"
  "fmt"
)

var Database *sql.DB

func InitialiseDatabase(database *sql.DB) {
  Database = database
}

func logQuery(query string) {
  fmt.Println(query)
}

func QueryRow(query string, args ...interface{}) *sql.Row {
  logQuery(query)
  return Database.QueryRow(query, args...)
}

func Query(query string, args ...interface{}) (*sql.Rows, error) {
  logQuery(query)
  return Database.Query(query, args...)
}
