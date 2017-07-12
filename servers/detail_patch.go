package servers

import (
	"github.com/bor3ham/reja/schema"
	"net/http"
)

func detailPATCH(
	w http.ResponseWriter,
	r *http.Request,
	c schema.Context,
	m *schema.Model,
	id string,
	include *schema.Include,
) {

}
