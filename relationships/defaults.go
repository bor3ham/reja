package relationships

import (
	"github.com/bor3ham/reja/schema"
)

func DefaultNullPointer(c schema.Context, instance interface{}) Pointer {
	return Pointer{
		Provided: true,
		Data:     nil,
	}
}
