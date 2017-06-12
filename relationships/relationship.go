package relationships

import (
    "net/http"
)

type Relationship interface {
  GetColumnNames() []string
  GetColumnVariables() []interface{}

  GetDefaultValue() interface{}
  GetValues(*http.Request, []string) map[string]interface{}
}

type PointerData struct {
  Type string `json:"type"`
  ID *string `json:"id"`
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
