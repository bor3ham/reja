package utils

import (
	"net/http"
)

type AuthError struct {
	text   string
	Status int
}

func (e AuthError) Error() string {
	return e.text
}

func Unauthorised(reason string) AuthError {
	return AuthError{
		text:   reason,
		Status: http.StatusUnauthorized,
	}
}

func Forbidden() AuthError {
	return AuthError{
		text:   "You do not have access to this resource or action.",
		Status: http.StatusForbidden,
	}
}

func TooManyRequests() AuthError {
	return AuthError{
		text:   "Too many requests.",
		Status: http.StatusTooManyRequests,
	}
}
