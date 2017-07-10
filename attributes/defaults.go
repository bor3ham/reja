package attributes

import (
// "github.com/bor3ham/reja/schema"
)

func DefaultBlankText(instance interface{}) TextValue {
	blank := ""
	return TextValue{
		Provided: true,
		Value:    &blank,
	}
}
