package reja

type Instance interface {
  GetID() string
  SetID(string)
  SetValues(values []interface{})
}
