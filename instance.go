package reja

type Instance interface {
	GetFields() []interface{}
  SetValues(values []interface{})
	Clean()
}
