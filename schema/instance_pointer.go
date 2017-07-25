package schema

type InstancePointer struct {
	ID   *string `json:"id"`
	Type string  `json:"type"`
}

func (ip InstancePointer) Equal(oip InstancePointer) bool {
	if ip.ID != nil && oip.ID != nil {
		return *ip.ID == *oip.ID && ip.Type == oip.Type
	}
	if ip.ID == nil && oip.ID == nil {
		return ip.Type == oip.Type
	}
	return false
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
