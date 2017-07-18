package attributes

import (
	"errors"
	"fmt"
	"github.com/bor3ham/reja/schema"
)

type BaseFilter struct {
	QArgKey string
	QArgValues []string
}

func (bf BaseFilter) GetQArgKey() string {
	return bf.QArgKey
}
func (bf BaseFilter) GetQArgValues() []string {
	return bf.QArgValues
}

func filterException(text string, args ...interface{}) ([]schema.Filter, error) {
	return []schema.Filter{}, errors.New(fmt.Sprintf(text, args...))
}
