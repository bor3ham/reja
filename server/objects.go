package server

import (
	"fmt"
	"github.com/bor3ham/reja/schema"
	"strings"
	"sync"
)

const USE_OBJECT_CACHE = true

type RelationResult struct {
	Key          string
	Index        int
	Default      interface{}
	Values       map[string]interface{}
	RelationMaps map[string]map[string][]string
}

type IncludeResult struct {
	Instances []schema.Instance
	Included  []schema.Instance
	Error     error
}

func flattened(fields [][]interface{}) []interface{} {
	var flatList []interface{}
	for _, relation := range fields {
		flatList = append(flatList, relation...)
	}
	return flatList
}

func combineRelations(
	maps ...map[string]map[string][]string,
) map[string]map[string][]string {
	combinedMap := map[string]map[string][]string{}
	for _, relations := range maps {
		for key, models := range relations {
			_, exists := combinedMap[key]
			if !exists {
				combinedMap[key] = map[string][]string{}
			}
			for model, ids := range models {
				_, exists = combinedMap[key][model]
				if !exists {
					combinedMap[key][model] = []string{}
				}
				combinedMap[key][model] = append(combinedMap[key][model], ids...)
			}
		}
	}
	return combinedMap
}

func (rc *RequestContext) GetObjects(
	m *schema.Model,
	objectIds []string,
	offset int,
	limit int,
	include *schema.Include,
) (
	[]schema.Instance,
	[]schema.Instance,
	error,
) {
	var cacheHits []schema.Instance
	var cacheMaps []map[string]map[string][]string

	var query string
	columns := m.FieldColumns()
	columns = append(m.FieldColumns(), m.ExtraColumns()...)
	if len(objectIds) > 0 {
		// attempt to use cache
		var newIds []string

		for _, id := range objectIds {
			instance, relationMap := rc.GetCachedObject(m.Type, id)
			if instance != nil && USE_OBJECT_CACHE {
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

	instances := []schema.Instance{}
	listRelations := map[string]map[string][]string{}

	if len(query) > 0 {
		rows, err := rc.Query(query)
		if err != nil {
			return []schema.Instance{}, []schema.Instance{}, err
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
				return []schema.Instance{}, []schema.Instance{}, err
			}

			instance := m.Manager.Create()
			instance.SetID(id)
			instances = append(instances, instance)

			ids = append(ids, id)
		}

		var wg sync.WaitGroup
		relationResults := make(chan RelationResult)
		relationships := m.Relationships
		wg.Add(len(relationships))
		for relationIndex, relationship := range relationships {
			go func(wg *sync.WaitGroup, index int, relation schema.Relationship) {
				defer wg.Done()
				var relationExtras [][]interface{}
				for _, result := range extraFields {
					relationExtras = append(relationExtras, result[index])
				}

				values, maps := relation.GetValues(rc, ids, relationExtras)
				relationResults <- RelationResult{
					Index:        index,
					Key:          relation.GetKey(),
					Default:      relation.GetDefaultValue(),
					Values:       values,
					RelationMaps: maps,
				}
			}(&wg, relationIndex, relationship)
		}
		go func(wg *sync.WaitGroup) {
			wg.Wait()
			close(relationResults)
		}(&wg)

		relationDefaults := make([]interface{}, len(relationships))
		relationValues := make([]map[string]interface{}, len(relationships))
		relationMaps := make([]map[string]map[string][]string, len(relationships))

		for result := range relationResults {
			// re order relation results
			relationDefaults[result.Index] = result.Default
			relationValues[result.Index] = result.Values
			relationMaps[result.Index] = result.RelationMaps
		}

		for index, instance := range instances {
			instanceRelations := map[string]map[string][]string{}
			for relationIndex, value := range relationValues {
				key := relationships[relationIndex].GetKey()
				// get value or default
				id := instance.GetID()
				item, exists := value[id]
				if exists {
					instanceFields[index] = append(instanceFields[index], item)
				} else {
					instanceFields[index] = append(instanceFields[index], relationDefaults[relationIndex])
				}
				// add to instance relation map
				_, exists = relationMaps[relationIndex][id]
				if exists {
					_, exists = instanceRelations[key]
					if !exists {
						instanceRelations[key] = map[string][]string{}
					}
					for model, ids := range relationMaps[relationIndex][id] {
						_, exists = instanceRelations[key][model]
						if !exists {
							instanceRelations[key][model] = []string{}
						}
						instanceRelations[key][model] = append(instanceRelations[key][model], ids...)
					}
				}
			}
			instance.SetValues(instanceFields[index])
			// add complete relation map to flat map
			listRelations = combineRelations(listRelations, instanceRelations)
			// add instance to cache
			rc.CacheObject(instance, instanceRelations)
		}
	}

	// add cached instances and maps
	instances = append(instances, cacheHits...)
	for _, cacheMap := range cacheMaps {
		listRelations = combineRelations(listRelations, cacheMap)
	}

	var wg sync.WaitGroup
	includedResults := make(chan IncludeResult)
	for attribute, modelTypes := range listRelations {
		for modelType, ids := range modelTypes {
			childModel := rc.GetServer().GetModel(modelType)
			if childModel == nil {
				panic(fmt.Sprintf("Could not find model for model: %s", modelType))
			}

			wg.Add(1)
			go func(
				wg *sync.WaitGroup,
				rc *RequestContext,
				include *schema.Include,
				model *schema.Model,
				attribute string,
			) {
				defer wg.Done()

				childIncludes, exists := include.Children[attribute]
				if exists {
					childInstances, childIncluded, err := rc.GetObjects(
						model,
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
							Included:  childIncluded,
							Error:     nil,
						}
					}

				}
			}(&wg, rc, include, childModel, attribute)
		}
	}
	go func(wg *sync.WaitGroup) {
		wg.Wait()
		close(includedResults)
	}(&wg)
	var included []schema.Instance
	for result := range includedResults {
		if result.Error != nil {
			return []schema.Instance{}, []schema.Instance{}, result.Error
		}
		included = append(included, result.Instances...)
		included = append(included, result.Included...)
	}

	return instances, included, nil
}

func UniqueInstances(set []schema.Instance) []schema.Instance {
	var unique []schema.Instance
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
