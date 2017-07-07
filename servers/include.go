package servers

import (
	"errors"
	"fmt"
	"github.com/bor3ham/reja/schema"
	"strings"
)

func validateInclude(c schema.Context, model *schema.Model, include *schema.Include) error {
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
		var relation schema.Relationship
		for _, modelRelation := range model.Relationships {
			if modelRelation.GetKey() == key {
				relation = modelRelation
			}
		}
		if relation == nil {
			return errors.New(fmt.Sprintf("Relation %s not found on model %s", key, model.Type))
		}
		// recurse on its children
		childModel := c.GetServer().GetModel(relation.GetType())
		err := validateInclude(c, childModel, child)
		if err != nil {
			return err
		}
	}
	return nil
}

func parseInclude(
	c schema.Context,
	model *schema.Model,
	params map[string][]string,
) (
	*schema.Include,
	error,
) {
	// extract from querystring
	includeString, err := GetStringParam(
		params,
		"include",
		"Included Relations",
		"",
	)
	if err != nil {
		return nil, err
	}

	// split out of querystring into tree
	includeMap := schema.Include{
		Children: map[string]*schema.Include{},
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
				newInclude := schema.Include{
					Children: map[string]*schema.Include{},
				}
				baseLevel[component] = &newInclude
			}
			baseLevel = baseLevel[component].Children
		}
	}

	// validate the tree
	err = validateInclude(c, model, &includeMap)
	if err != nil {
		return nil, err
	}

	return &includeMap, nil
}
