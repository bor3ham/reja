package schema

import (
	"database/sql"
)

type Server interface {
	GetDatabase() *sql.DB

	GetDefaultDirectPageSize() int
	GetMaximumDirectPageSize() int
	GetIndirectPageSize() int

	GetModel(string) Model
}
