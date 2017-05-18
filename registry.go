package reja

var Models map[string]Model = make(map[string]Model)

func RegisterModel(m Model) {
  Models[m.Type] = m
}
