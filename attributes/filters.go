package attributes

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

func filterException(text string, args ...interface{}) ([]schema.Filter, error) {
	return []schema.Filter{}, errors.New(fmt.Sprintf(text, args...))
}
