package models

import (
	rejaInstances "github.com/bor3ham/reja/instances"
	"github.com/bor3ham/reja/context"
	"strings"
	"fmt"
)

func GetObjects(
	rc context.Context,
	m Model,
	objectIds []string,
	offset int,
	limit int,
	included *Include,
) (
	[]rejaInstances.Instance,
	error,
) {
	var query string
	columns := m.FieldNames()
	columns = append(m.FieldNames(), m.ExtraNames()...)
	if len(objectIds) > 0 {
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
			fmt.Sprintf("%s in (%s)", m.IDColumn, strings.Join(objectIds, ", ")),
		)
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

	rows, err :=rc.Query(query)
	if err != nil {
		return []rejaInstances.Instance{}, err
	}
	defer rows.Close()

	ids := []string{}
	instances := []rejaInstances.Instance{}
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
			return []rejaInstances.Instance{}, err
		}

		instance := m.Manager.Create()
		instance.SetID(id)
		instances = append(instances, instance)

		ids = append(ids, id)
	}

	// relation map
	relationshipMap := map[string]map[string][]string{}

	relationValues := []RelationResult{}
	for relationIndex, relationship := range m.Relationships {
		var relationExtras [][]interface{}
		for _, result := range extraFields {
			relationExtras = append(relationExtras, result[relationIndex])
		}

		values, relationMap := relationship.GetValues(rc, ids, relationExtras)
		relationValues = append(relationValues, RelationResult{
			Values:  values,
			Default: relationship.GetDefaultValue(),
		})
		for modelType, ids := range relationMap {
			_, exists := relationshipMap[modelType]
			if !exists {
				relationshipMap[modelType] = map[string][]string{}
			}
			relationshipMap[modelType][relationship.GetKey()] = ids
		}
	}

	for instance_index, instance := range instances {
		for _, value := range relationValues {
			item, exists := value.Values[instance.GetID()]
			if exists {
				instanceFields[instance_index] = append(instanceFields[instance_index], item)
			} else {
				instanceFields[instance_index] = append(instanceFields[instance_index], value.Default)
			}
		}
	}

	for instance_index, instance := range instances {
		instance.SetValues(instanceFields[instance_index])
	}

	return instances, nil
}
