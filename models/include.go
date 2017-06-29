package models

import (
	"errors"
	"fmt"
	"github.com/bor3ham/reja/http"
	"strings"
)

type Include struct {
	Children map[string]*Include
}

func validateInclude(model *Model, include *Include) error {
	// its valid without children
	if len(include.Children) == 0 {
		return nil
	}
	// pass unknown models, can't effectively validate GFK's
	if model == nil {
		return nil
	}
	// go through children
	for key, child := range include.Children {
		var relation Relationship
		for _, modelRelation := range model.Relationships {
			if modelRelation.GetKey() == key {
				relation = modelRelation
			}
		}
		if relation == nil {
			return errors.New(fmt.Sprintf("Relation %s not found on model %s", key, model.Type))
		}
		// recurse on its children
		childModel := GetModel(relation.GetType())
		err := validateInclude(childModel, child)
		if err != nil {
			return err
		}
	}
	return nil
}

func parseInclude(model *Model, params map[string][]string) (*Include, error) {
	// extract from querystring
	includeString, err := http.GetStringParam(
		params,
		"include",
		"Included Relations",
		"",
	)
	if err != nil {
		return nil, err
	}

	// split out of querystring into tree
	includeMap := Include{
		Children: map[string]*Include{},
	}
	arguments := strings.Split(includeString, ",")
	for _, argument := range arguments {
		components := strings.Split(argument, ".")
		baseLevel := includeMap.Children
		for _, component := range components {
			if len(component) == 0 {
				continue
			}
			_, exists := baseLevel[component]
			if !exists {
				newInclude := Include{
					Children: map[string]*Include{},
				}
				baseLevel[component] = &newInclude
			}
			baseLevel = baseLevel[component].Children
		}
	}

	// validate the tree
	err = validateInclude(model, &includeMap)
	if err != nil {
		return nil, err
	}

	return &includeMap, nil
}
