package relationships

import (
	"errors"
	"github.com/bor3ham/reja/format"
	"github.com/bor3ham/reja/instances"
)

type Pointer struct {
	Data *instances.InstancePointer `json:"data"`
}

type PointerSet struct {
	Data []instances.InstancePointer `json:"data"`
}

// temporary function to flatten list as part of refactor
func flattenMaps(relationMap map[string]map[string][]string) map[string][]string {
	flatMap := map[string][]string{}
	for _, relations := range relationMap {
		for modelType, ids := range relations {
			_, exists := flatMap[modelType]
			if !exists {
				flatMap[modelType] = []string{}
			}
			flatMap[modelType] = append(flatMap[modelType], ids...)
		}
	}
	// unique the list of ids
	distinctMap := map[string][]string{}
	for model, ids := range flatMap {
		distinctMap[model] = []string{}
		distincts := map[string]bool{}
		for _, id := range ids {
			distincts[id] = true
		}
		for id, _ := range distincts {
			distinctMap[model] = append(distinctMap[model], id)
		}
	}
	return distinctMap
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
