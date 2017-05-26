package relationships

type Relationship interface {
  GetColumns() []string
}

type PointerData struct {
  Type string `json:"type"`
  ID interface{} `json:"id"`
}

type Pointer struct {
  Data *PointerData `json:"data"`
}

func (p *Pointer) Clean() {
  if p.Data != nil {
    if p.Data.ID == nil {
      p.Data = nil
    }
  }
}
