package relationships

import (
	"github.com/bor3ham/reja/context"
	"github.com/bor3ham/reja/instances"
	"github.com/bor3ham/reja/format"
)

type Relationship interface {
	GetKey() string
	GetType() string

	GetInstanceColumnNames() []string
	GetInstanceColumnVariables() []interface{}
	GetExtraColumnNames() []string
	GetExtraColumnVariables() []interface{}

	GetDefaultValue() interface{}
	GetValues(
		context.Context,
		[]string,
		[][]interface{},
	) (
		map[string]interface{},
		map[string]map[string][]string,
	)

	ValidateNew(interface{}) (interface{}, error)
}

type Pointer struct {
	Data *instances.InstancePointer `json:"data"`
}

type PointerSet struct {
	Data     []instances.InstancePointer `json:"data"`
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
		instanceVals, ok := val.(format.Page)
		if !ok {
			panic("Bad pointer page value")
		}
		pageVal = PointerSet{}
		for _, genericInstance := range instanceVals.Data {
			instance, ok := genericInstance.(instances.Instance)
			if !ok {
				panic("Bad pointer page value")
			}
			instanceId := instance.GetID()
			pageVal.Data = append(pageVal.Data, instances.InstancePointer{
				ID: &instanceId,
				Type: instance.GetType(),
			})
		}

	}
	return pageVal
}
