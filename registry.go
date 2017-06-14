package reja

import (
  "github.com/bor3ham/reja/models"
)

var Models map[string]models.Model = make(map[string]models.Model)

func RegisterModel(m models.Model) {
	Models[m.Type] = m
}
