package instances

type Instance interface {
	GetID() string
	SetID(string)
	GetType() string
	SetValues(values []interface{})
	GetValues() []interface{}
}

type InstancePointer struct {
	ID   *string `json:"id"`
	Type string  `json:"type"`
}

func (ip *InstancePointer) GetID() string {
	if ip.ID == nil {
		return ""
	}
	return *ip.ID
}
func (ip *InstancePointer) SetID(id string) {
	ip.ID = &id
}
func (ip *InstancePointer) GetType() string {
	return ip.Type
}
