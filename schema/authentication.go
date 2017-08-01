package schema

import (
	"net/http"
)

type User interface {
}

type Authenticator interface {
	GetUser(*http.Request) (User, error)
}