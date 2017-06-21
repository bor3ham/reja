package relationships

import (
	"github.com/bor3ham/reja/context"
)

type Relationship interface {
	GetKey() string
	GetType() string

	GetInstanceColumnNames() []string
	GetInstanceColumnVariables() []interface{}
	GetExtraColumnNames() []string
	GetExtraColumnVariables() []interface{}

	GetDefaultValue() interface{}
	GetValues(
		context.Context,
		[]string,
		[][]interface{},
	) (
		map[string]interface{},
		map[string][]string,
	)
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
