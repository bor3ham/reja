package reja

import (
  "database/sql"
)

var Database *sql.DB

func Initialise(database *sql.DB) {
  Database = database
}
