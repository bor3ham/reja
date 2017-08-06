package schema

import (
	"net/http"
)

type User interface {
}

type Authenticator interface {
	GetUser(http.ResponseWriter, *http.Request, Context) (User, error)
}
