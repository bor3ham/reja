package servers

import (
	"fmt"
	"github.com/bor3ham/reja/schema"
	"github.com/bor3ham/reja/utils"
	"github.com/davecgh/go-spew/spew"
	"strings"
)

func CanAccessAllInstances(c schema.Context, instances []schema.Instance) bool {
	typeMap := map[string][]string{}
	for _, instance := range instances {
		instanceType := instance.GetType()
		_, exists := typeMap[instanceType]
		if !exists {
			typeMap[instanceType] = []string{}
		}
		typeMap[instanceType] = append(typeMap[instanceType], instance.GetID())
	}
	spew.Dump(typeMap)

	server := c.GetServer()
	for modelType, ids := range typeMap {
		model := server.GetModel(modelType)
		nextArg, idSpots := utils.StringInAsArgs(1, ids)
		query := fmt.Sprintf(
			`
				select %s from %s where %s in (%s)
			`,
			model.IDColumn,
			model.Table,
			model.IDColumn,
			idSpots,
		)
		args := []interface{}{}
		for _, id := range ids {
			args = append(args, id)
		}

		authFilters, authArgs := model.Manager.GetFilterForUser(c.GetUser(), nextArg)
		if len(authFilters) > 0 {
			query += " and " + strings.Join(authFilters, " and ")
			args = append(args, authArgs...)
		}

		rows, err := c.Query(query, args...)
		if err != nil {
			panic(err)
		}
		resultIds := []string{}
		for rows.Next() {
			var id string
			rows.Scan(&id)
			resultIds = append(resultIds, id)
		}

		for _, id := range ids {
			found := false
			for _, result := range resultIds {
				if result == id {
					found = true
				}
			}
			if !found {
				return false
			}
		}
	}

	return true
}
