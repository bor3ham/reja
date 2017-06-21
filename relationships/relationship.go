package relationships

import (
	"github.com/bor3ham/reja/context"
)

type Relationship interface {
	GetKey() string
	GetType() string

	GetColumnNames() []string
	GetColumnVariables() []interface{}

	GetDefaultValue() interface{}
	GetValues(context.Context, []string) map[string]interface{}
}

type PointerData struct {
	Type string  `json:"type"`
	ID   *string `json:"id"`
}

type Pointer struct {
	Data *PointerData `json:"data"`
}

type Pointers struct {
	Data []*PointerData `json:"data"`
}

func (p *Pointer) Clean() {
	if p.Data != nil {
		if p.Data.ID == nil {
			p.Data = nil
		}
	}
}
