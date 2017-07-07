package servers

import (
	"fmt"
	"net/http"
)

func BadRequest(w http.ResponseWriter, title string, detail string) {
	errorBlob := struct {
		Exceptions []interface{} `json:"errors"`
	}{}
	errorBlob.Exceptions = append(errorBlob.Exceptions, struct {
		Title  string
		Detail string
	}{
		Title:  title,
		Detail: detail,
	})
	errorBytes := MustJSONMarshal(errorBlob)
	fmt.Fprintf(w, string(errorBytes))
}
