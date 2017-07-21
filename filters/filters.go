package filters

import (
	"errors"
	"fmt"
	"github.com/bor3ham/reja/schema"
)

const ISNULL_SUFFIX = "__is_null"
const LENGTH_SUFFIX = "__length"
const GT_SUFFIX = "__gt"
const LT_SUFFIX = "__lt"
const CONTAINS_SUFFIX = "__contains"
const AFTER_SUFFIX = "__after"
const BEFORE_SUFFIX = "__before"
const TYPE_SUFFIX = "__type"
const ID_SUFFIX = "__id"

type FilterDescription struct {
	Key string `json:"key"`
	Description string `json:"description"`
	Examples []string `json:"examples"`
}

func Exception(text string, args ...interface{}) ([]schema.Filter, error) {
	return []schema.Filter{}, errors.New(fmt.Sprintf(text, args...))
}
