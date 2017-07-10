package relationships

import (
	"errors"
	"github.com/bor3ham/reja/schema"
)

type Pointer struct {
	Provided bool `json:"-"`
	Data *schema.InstancePointer `json:"data"`
}

type PointerSet struct {
	Provided bool                     `json:"-"`
	Data     []schema.InstancePointer `json:"data"`
}

type RelationshipStub struct{}

func (stub RelationshipStub) GetSelectDirectColumns() []string {
	return []string{}
}
func (stub RelationshipStub) GetSelectDirectVariables() []interface{} {
	return []interface{}{}
}
func (stub RelationshipStub) GetSelectExtraColumns() []string {
	return []string{}
}
func (stub RelationshipStub) GetSelectExtraVariables() []interface{} {
	return []interface{}{}
}
func (stub RelationshipStub) DefaultFallback(
	c schema.Context,
	val interface{},
	instance interface{},
) (
	interface{},
	error,
) {
	return val, nil
}
func (stub RelationshipStub) Validate(c schema.Context, val interface{}) (interface{}, error) {
	return val, nil
}
func (stub RelationshipStub) GetInsertColumns(val interface{}) []string {
	return []string{}
}
func (stub RelationshipStub) GetInsertValues(val interface{}) []interface{} {
	return []interface{}{}
}
func (stub RelationshipStub) GetInsertQueries(newId string, val interface{}) []schema.Query {
	return []schema.Query{}
}

func AssertPointer(val interface{}) Pointer {
	pointerVal, ok := val.(Pointer)
	if !ok {
		panic("Bad pointer value")
	}
	return pointerVal
}

func AssertPointerSet(val interface{}) PointerSet {
	pageVal, ok := val.(PointerSet)
	if !ok {
		panic("Bad pointer set value")
	}
	return pageVal
}

func ParseStringInstancePointer(stringPointer interface{}) (schema.InstancePointer, error) {
	pointer, ok := stringPointer.(map[string]interface{})
	if !ok {
		return schema.InstancePointer{}, errors.New("Invalid pointer.")
	}

	// parse the id
	pointerId, exists := pointer["id"]
	if !exists {
		return schema.InstancePointer{}, errors.New("Invalid pointer (missing ID).")
	}
	parsedId, ok := pointerId.(string)
	if !ok {
		return schema.InstancePointer{}, errors.New("Invalid pointer (bad ID).")
	}

	// parse the type
	pointerType, exists := pointer["type"]
	if !exists {
		return schema.InstancePointer{}, errors.New("Invalid pointer (missing Type).")
	}
	parsedType, ok := pointerType.(string)
	if !ok {
		return schema.InstancePointer{}, errors.New("Invalid pointer (bad Type).")
	}

	return schema.InstancePointer{
		Type: parsedType,
		ID: &parsedId,
	}, nil
}

func ParseResultPointer(val interface{}) (Pointer, error) {
	resultVal, ok := val.(schema.Result)
	if !ok {
		panic("Invalid result")
	}

	if resultVal.Provided {
		if resultVal.Data == nil {
			return Pointer{
				Provided: true,
				Data: nil,
			}, nil
		} else {
			pointer, err := ParseStringInstancePointer(resultVal.Data)
			if err != nil {
				return Pointer{}, err
			}
			return Pointer{
				Provided: true,
				Data: &pointer,
			}, nil
		}
	}

	return Pointer{
		Provided: false,
		Data: nil,
	}, nil
}

func ParsePagePointerSet(val interface{}) (PointerSet, error) {
	pageVal, ok := val.(schema.Page)
	if !ok {
		panic("Invalid page")
	}
	pointersVal := PointerSet{
		Provided: pageVal.Provided,
	}
	for _, stringPointer := range pageVal.Data {
		pointer, err := ParseStringInstancePointer(stringPointer)
		if err != nil {
			return PointerSet{}, err
		}
		pointersVal.Data = append(pointersVal.Data, pointer)
	}
	return pointersVal, nil
}
