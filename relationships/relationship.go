package relationships

import (
	"errors"
	"github.com/bor3ham/reja/format"
	"github.com/bor3ham/reja/instances"
	"github.com/bor3ham/reja/context"
)

type Pointer struct {
	Data *instances.InstancePointer `json:"data"`
}

type PointerSet struct {
	Data []instances.InstancePointer `json:"data"`
}

type RelationshipStub struct {}
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
func (stub RelationshipStub) ValidateNew(c context.Context, val interface{}) (interface{}, error) {
	return val, nil
}

func AssertPointerSet(val interface{}) PointerSet {
	pageVal, ok := val.(PointerSet)
	if !ok {
		panic("Bad pointer page value")
	}
	return pageVal
}

func ParsePagePointerSet(val interface{}) (PointerSet, error) {
	pageVal, ok := val.(format.Page)
	if !ok {
		panic("Invalid pointer set")
	}
	pointersVal := PointerSet{}
	for _, stringPointer := range pageVal.Data {
		pointer, ok := stringPointer.(map[string]interface{})
		if !ok {
			return PointerSet{}, errors.New("Invalid pointer in pointer set.")
		}

		// parse the id
		pointerId, exists := pointer["id"]
		if !exists {
			return PointerSet{}, errors.New("Invalid pointer in pointer set (missing ID).")
		}
		parsedId, ok := pointerId.(string)
		if !ok {
			return PointerSet{}, errors.New("Invalid pointer in pointer set (bad ID).")
		}

		// parse the type
		pointerType, exists := pointer["type"]
		if !exists {
			return PointerSet{}, errors.New("Invalid pointer in pointer set (missing Type).")
		}
		parsedType, ok := pointerType.(string)
		if !ok {
			return PointerSet{}, errors.New("Invalid pointer in pointer set (bad Type).")
		}

		// valid pointer
		pointersVal.Data = append(pointersVal.Data, instances.InstancePointer{
			Type: parsedType,
			ID:   &parsedId,
		})
	}
	return pointersVal, nil
}
