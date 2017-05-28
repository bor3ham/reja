package relationships

type Relationship interface {
  GetColumns() []string
  GetKeyedValues(string) map[int]interface{}
  GetEmptyKeyedValue() interface{}
}

type PointerData struct {
  Type string `json:"type"`
  ID interface{} `json:"id"`
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
