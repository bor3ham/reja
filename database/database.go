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

func LogQuery(query string) {
	// log.Println(query)
}
