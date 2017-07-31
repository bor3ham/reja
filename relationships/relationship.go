package relationships

import (
	"errors"
	"github.com/bor3ham/reja/schema"
)

type Pointer struct {
	Provided bool                    `json:"-"`
	Data     *schema.InstancePointer `json:"data"`
}

func (p Pointer) Equal(op Pointer) bool {
	if p.Data == nil {
		return (op.Data == nil)
	} else if op.Data == nil {
		return false
	}
	return p.Data.Equal(*op.Data)
}

func pointerFromResult(val interface{}) Pointer {
	asResult, ok := val.(schema.Result)
	if !ok {
		panic("Bad result value")
	}
	if asResult.Data == nil {
		return Pointer{}
	}
	asPointer, ok := asResult.Data.(schema.InstancePointer)
	if !ok {
		panic("Bad pointer value in result value")
	}
	return Pointer{Data: &asPointer}
}

type PointerSet struct {
	Provided bool                     `json:"-"`
	Data     []schema.InstancePointer `json:"data"`
}

func (ps PointerSet) Counts() map[string]int {
	counts := map[string]int{}
	if ps.Data == nil {
		return counts
	}
	for _, pointer := range ps.Data {
		key := pointer.Type + ":"
		if pointer.ID == nil {
			key += "nil"
		} else {
			key += *pointer.ID
		}
		_, exists := counts[key]
		if exists {
			counts[key] += 1
		} else {
			counts[key] = 1
		}
	}
	return counts
}
func (ps PointerSet) Equal(ops PointerSet) bool {
	if ps.Data == nil {
		return (ops.Data == nil)
	} else if ops.Data == nil {
		return false
	}
	psCounts := ps.Counts()
	opsCounts := ops.Counts()
	for key, count := range psCounts {
		oCount, exists := opsCounts[key]
		if !exists || oCount != count {
			return false
		}
	}
	for key, count := range opsCounts {
		oCount, exists := psCounts[key]
		if !exists || oCount != count {
			return false
		}
	}
	return true
}

func pointerSetFromPage(val interface{}) PointerSet {
	asPage, ok := val.(schema.Page)
	if !ok {
		panic("Bad page value")
	}
	dataSet := []schema.InstancePointer{}
	if asPage.Data != nil {
		for _, item := range asPage.Data {
			asInstance, ok := item.(schema.InstancePointer)
			if !ok {
				panic("Bad instance value in page value")
			}
			dataSet = append(dataSet, asInstance)
		}
	}
	return PointerSet{Data: dataSet}
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
func (stub RelationshipStub) AvailableFilters() []interface{} {
	return []interface{}{}
}
func (stub RelationshipStub) ValidateFilters(map[string][]string) ([]schema.Filter, error) {
	return []schema.Filter{}, nil
}
func (stub RelationshipStub) GetFilterWhere(int, map[string][]string) ([]string, []interface{}) {
	return []string{}, []interface{}{}
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
func (stub RelationshipStub) ValidateUpdate(
	c schema.Context,
	newVal interface{},
	oldVal interface{},
) (
	interface{},
	error,
) {
	return nil, nil
}
func (stub RelationshipStub) GetInsert(val interface{}) ([]string, []interface{}) {
	return []string{}, []interface{}{}
}
func (stub RelationshipStub) GetInsertQueries(newId string, val interface{}) []schema.Query {
	return []schema.Query{}
}
func (stub RelationshipStub) GetUpdateQueries(id string, oldVal interface{}, newVal interface{}) []schema.Query {
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
		ID:   &parsedId,
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
				Data:     nil,
			}, nil
		} else {
			pointer, err := ParseStringInstancePointer(resultVal.Data)
			if err != nil {
				return Pointer{}, err
			}
			return Pointer{
				Provided: true,
				Data:     &pointer,
			}, nil
		}
	}

	return Pointer{
		Provided: false,
		Data:     nil,
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
