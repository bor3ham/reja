package models

import (
	"fmt"
	"github.com/bor3ham/reja/context"
	rejaInstances "github.com/bor3ham/reja/instances"
	"github.com/bor3ham/reja/relationships"
	"strings"
	"sync"
	// "github.com/davecgh/go-spew/spew"
)

type RelationResult struct {
	Key string
	Index int
	Values  map[string]interface{}
	Default interface{}
	Map map[string][]string
}

type IncludeResult struct {
	Instances []rejaInstances.Instance
	Included []rejaInstances.Instance
	Error error
}

func GetObjects(
	rc context.Context,
	m Model,
	objectIds []string,
	offset int,
	limit int,
	include *Include,
) (
	[]rejaInstances.Instance,
	[]rejaInstances.Instance,
	error,
) {
	var cacheHits []rejaInstances.Instance
	var cacheMaps []map[string]map[string][]string

	var query string
	columns := m.FieldNames()
	columns = append(m.FieldNames(), m.ExtraNames()...)
	if len(objectIds) > 0 {
		// attempt to use cache
		var newIds []string

		for _, id := range objectIds {
			instance, relationMap := rc.GetCachedObject(m.Type, id)
			if instance != nil {
				cacheHits = append(cacheHits, instance)
				cacheMaps = append(cacheMaps, relationMap)
			} else {
				newIds = append(newIds, id)
			}
		}

		if len(newIds) > 0 {
			query = fmt.Sprintf(
				`
					select
						%s,
						%s
					from %s
					where %s
		    	`,
				m.IDColumn,
				strings.Join(columns, ","),
				m.Table,
				fmt.Sprintf("%s in (%s)", m.IDColumn, strings.Join(newIds, ", ")),
			)
		}
	} else {
		query = fmt.Sprintf(
			`
				select
					%s,
					%s
				from %s
				limit %d
				offset %d
	    	`,
			m.IDColumn,
			strings.Join(columns, ","),
			m.Table,
			limit,
			offset,
		)
	}

	instances := []rejaInstances.Instance{}
	relationshipMap := map[string]map[string][]string{}

	if len(query) > 0 {
		rows, err := rc.Query(query)
		if err != nil {
			return []rejaInstances.Instance{}, []rejaInstances.Instance{}, err
		}
		defer rows.Close()

		ids := []string{}
		instanceFields := [][]interface{}{}
		extraFields := [][][]interface{}{}
		for rows.Next() {
			var id string
			fields := m.FieldVariables()
			instanceFields = append(instanceFields, fields)
			extras := m.ExtraVariables()
			extraFields = append(extraFields, extras)
			flatExtras := flattened(extras)

			scanFields := []interface{}{}
			scanFields = append(scanFields, &id)
			scanFields = append(scanFields, fields...)
			scanFields = append(scanFields, flatExtras...)
			err := rows.Scan(scanFields...)
			if err != nil {
				return []rejaInstances.Instance{}, []rejaInstances.Instance{}, err
			}

			instance := m.Manager.Create()
			instance.SetID(id)
			instances = append(instances, instance)

			ids = append(ids, id)
		}


		var wg sync.WaitGroup
		relationResults := make(chan RelationResult)
		wg.Add(len(m.Relationships))
		for relationIndex, relationship := range m.Relationships {
			go func(wg *sync.WaitGroup, index int, relation relationships.Relationship) {
				defer wg.Done()
				var relationExtras [][]interface{}
				for _, result := range extraFields {
					relationExtras = append(relationExtras, result[index])
				}

				values, relationMap := relation.GetValues(rc, ids, relationExtras)
				relationResults <- RelationResult{
					Index: index,
					Key: relation.GetKey(),
					Values: values,
					Default: relation.GetDefaultValue(),
					Map: relationMap,
				}
			}(&wg, relationIndex, relationship)
		}
		go func(wg *sync.WaitGroup) {
			wg.Wait()
			close(relationResults)
		}(&wg)

		relationValues := make([]map[string]interface{}, len(m.Relationships))
		relationDefaults := make([]interface{}, len(m.Relationships))

		for result := range relationResults {
			// re order relation results
			relationValues[result.Index] = result.Values
			relationDefaults[result.Index] = result.Default
			// and add to running map
			for modelType, ids := range result.Map {
				_, exists := relationshipMap[modelType]
				if !exists {
					relationshipMap[modelType] = map[string][]string{}
				}
				relationshipMap[modelType][result.Key] = ids
			}
		}

		for index, instance := range instances {
			for relationIndex, value := range relationValues {
				item, exists := value[instance.GetID()]
				if exists {
					instanceFields[index] = append(instanceFields[index], item)
				} else {
					instanceFields[index] = append(instanceFields[index], relationDefaults[relationIndex])
				}
			}
			instance.SetValues(instanceFields[index])
			// add instance to cache
			rc.CacheObject(instance, relationshipMap)
		}
	}

	// add cached instances and maps
	instances = append(instances, cacheHits...)
	for _, cacheMap := range cacheMaps {
		for modelType, attributes := range cacheMap {
			_, exists := relationshipMap[modelType]
			if !exists {
				relationshipMap[modelType] = map[string][]string{}
			}
			for attribute, ids := range attributes {
				_, exists = relationshipMap[modelType][attribute]
				if !exists {
					relationshipMap[modelType][attribute] = []string{}
				}
				relationshipMap[modelType][attribute] = append(relationshipMap[modelType][attribute], ids...)
			}
		}
	}

	var wg sync.WaitGroup
	includedResults := make(chan IncludeResult)
	wg.Add(len(relationshipMap))
	for modelType, attributes := range relationshipMap {
		childModel := GetModel(modelType)

		go func(
			wg *sync.WaitGroup,
			rc context.Context,
			include *Include,
			model *Model,
			attributes map[string][]string,
		) {
			defer wg.Done()

			for attribute, ids := range attributes {
				childIncludes, exists := include.Children[attribute]
				if exists {
					childInstances, childIncluded, err := GetObjects(
						rc,
						*model,
						ids,
						0,
						0,
						childIncludes,
					)
					if err != nil {
						includedResults <- IncludeResult{
							Error: err,
						}
					} else {
						includedResults <- IncludeResult{
							Instances: childInstances,
							Included: childIncluded,
							Error: nil,
						}
					}

				}
			}
		}(&wg, rc, include, childModel, attributes)
	}
	go func(wg *sync.WaitGroup) {
		wg.Wait()
		close(includedResults)
	}(&wg)
	var included []rejaInstances.Instance
	for result := range includedResults {
		if result.Error != nil {
			return []rejaInstances.Instance{}, []rejaInstances.Instance{}, result.Error
		}
		included = append(included, result.Instances...)
		included = append(included, result.Included...)
	}

	return instances, included, nil
}

func UniqueInstances(set []rejaInstances.Instance) []rejaInstances.Instance {
	var unique []rejaInstances.Instance
	known := map[string]map[string]bool{}
	for _, instance := range set {
		instanceType := instance.GetType()
		instanceId := instance.GetID()
		_, exists := known[instanceType]
		if !exists {
			known[instanceType] = map[string]bool{}
		}
		_, exists = known[instanceType][instanceId]
		if !exists {
			unique = append(unique, instance)
			known[instanceType][instanceId] = true
		}
	}
	return unique
}
