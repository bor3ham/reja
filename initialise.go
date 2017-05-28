package reja

import (
	"database/sql"
  "github.com/bor3ham/reja/database"
)

func Initialise(db *sql.DB) {
	database.InitialiseDatabase(db)
}
