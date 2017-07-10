package attributes

import (
// "github.com/bor3ham/reja/schema"
)

func DefaultText(textValue string) (func(interface{}) TextValue) {
	return func(instance interface{}) TextValue {
		return TextValue{
			Provided: true,
			Value: &textValue,
		}
	}
}

func DefaultBool(boolValue bool) (func(interface{}) BoolValue) {
	return func(instance interface{}) BoolValue {
		return BoolValue{
			Provided: true,
			Value: &boolValue,
		}
	}
}
