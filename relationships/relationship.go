package relationships

import (
	"github.com/bor3ham/reja/context"
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
}

type PointerData struct {
	Type string  `json:"type"`
	ID   *string `json:"id"`
}

type Pointer struct {
	Data *PointerData `json:"data"`
}

type Pointers struct {
	Data []*PointerData `json:"data"`
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
