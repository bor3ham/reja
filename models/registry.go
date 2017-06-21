package models

var Models map[string]Model = make(map[string]Model)

func RegisterModel(m Model) {
	Models[m.Type] = m
}

func GetModel(stringType string) *Model {
	model, exists := Models[stringType]
	if exists {
		return &model
	}
	return nil
}
