package reja

type Instance interface {
  GetID() int
	GetFields() []interface{}
  SetValues(values []interface{})
	Clean()
}
