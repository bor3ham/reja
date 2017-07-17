package servers

import (
	"net/http"
	"encoding/json"
	"github.com/bor3ham/reja/schema"
)

func BadRequest(c schema.Context, w http.ResponseWriter, title string, detail string) {
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
	encoder := json.NewEncoder(w)
	if c.GetServer().Whitespace() {
		encoder.SetIndent("", "    ")
	}
	encoder.Encode(errorBlob)
}
