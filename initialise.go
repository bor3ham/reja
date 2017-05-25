package reja

import (
	"database/sql"
)

func Initialise(database *sql.DB) {
  InitialiseDatabase(database)
}
