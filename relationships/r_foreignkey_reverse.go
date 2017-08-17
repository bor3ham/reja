package relationships

import (
	"errors"
	"fmt"
	"github.com/bor3ham/reja/schema"
	"github.com/bor3ham/reja/utils"
	"strings"
)

type ForeignKeyReverse struct {
	RelationshipStub
	Key            string
	SourceTable    string
	SourceIDColumn string
	ColumnName     string
	Type           string
	Nullable       bool
	Default        func(schema.Context, interface{}) PointerSet
}

func (fkr ForeignKeyReverse) GetKey() string {
	return fkr.Key
}
func (fkr ForeignKeyReverse) GetType() string {
	return fkr.Type
}

func (fkr ForeignKeyReverse) GetDefaultValue() interface{} {
	return schema.Page{}
}
func (fkr ForeignKeyReverse) GetValues(
	c schema.Context,
	m *schema.Model,
	ids []string,
	extra [][]interface{},
	offset int,
	pageSize int,
) (
	map[string]interface{},
	map[string]map[string][]string,
) {
	if len(ids) == 0 {
		return map[string]interface{}{}, map[string]map[string][]string{}
	}
	args := []interface{}{}
	spots := []string{}
	for index, id := range ids {
		spots = append(spots, fmt.Sprintf("$%d", index+1))
		args = append(args, id)
	}
	filter := fmt.Sprintf("%s in (%s)", fkr.ColumnName, strings.Join(spots, ", "))

	server := c.GetServer()
	otherModel := server.GetModel(fkr.Type)
	order, _, err := otherModel.GetOrderQuery(otherModel.DefaultOrder)
	if err != nil {
		panic(err)
	}

	query := fmt.Sprintf(
		`
			select
				%s,
				%s
			from %s
			where %s
			%s
		`,
		fkr.SourceIDColumn,
		fkr.ColumnName,
		fkr.SourceTable,
		filter,
		order,
	)
	rows, err := c.Query(query, args...)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	values := map[string]schema.Page{}
	maps := map[string]map[string][]string{}
	// fill in initial page data
	for _, id := range ids {
		value := schema.Page{
			Metadata: map[string]interface{}{},
			Data:     []interface{}{},
		}
		value.Metadata["total"] = 0
		value.Metadata["count"] = 0
		values[id] = value
	}
	// go through result data
	for rows.Next() {
		var otherId, ownId string
		rows.Scan(&otherId, &ownId)
		value, exists := values[ownId]
		if !exists {
			panic("Found unexpected id in results")
		}
		total, ok := value.Metadata["total"].(int)
		if !ok {
			panic("Bad total received")
		}
		count, ok := value.Metadata["count"].(int)
		if !ok {
			panic("Bad count received")
		}
		if total >= offset && (pageSize < 0 || count <= pageSize) {
			count += 1
			value.Data = append(value.Data, schema.InstancePointer{
				ID:   &otherId,
				Type: fkr.Type,
			})
			value.Metadata["count"] = count
		}
		total += 1
		value.Metadata["total"] = total

		// add to maps
		_, exists = maps[ownId]
		if !exists {
			maps[ownId] = map[string][]string{}
			maps[ownId][fkr.Type] = []string{}
		}
		maps[ownId][fkr.Type] = append(maps[ownId][fkr.Type], otherId)
		// update value
		values[ownId] = value
	}
	// create the links
	for id, value := range values {
		total, ok := value.Metadata["total"].(int)
		if !ok {
			panic("Bad total received")
		}
		value.Links = utils.GetPaginationLinks(
			relationLink(c, m.Type, id, fkr.Key),
			1,
			pageSize,
			server.GetDefaultDirectPageSize(),
			total,
			map[string][]string{},
		)
		values[id] = value
	}
	// generalise values
	generalValues := map[string]interface{}{}
	for id, value := range values {
		generalValues[id] = value
	}
	return generalValues, maps
}

func (fkr *ForeignKeyReverse) DefaultFallback(
	c schema.Context,
	val interface{},
	instance interface{},
) (
	interface{},
	error,
) {
	var fkrVal PointerSet
	if val == nil {
		fkrVal = PointerSet{Provided: false}
	} else {
		var err error
		fkrVal, err = ParsePagePointerSet(val)
		if err != nil {
			return nil, err
		}
	}

	if !fkrVal.Provided {
		if fkr.Default != nil {
			return fkr.Default(c, instance), nil
		}
		return nil, nil
	}
	return fkrVal, nil
}
func (fkr *ForeignKeyReverse) Validate(c schema.Context, val interface{}) (interface{}, error) {
	fkrVal := AssertPointerSet(val)

	// validate the types are correct
	for _, pointer := range fkrVal.Data {
		if pointer.Type != fkr.Type {
			return nil, errors.New(fmt.Sprintf(
				"Relationship '%s' invalid: Incorrect type in set.",
				fkr.Key,
			))
		}
	}
	// find duplicates
	ids := map[string]bool{}
	for _, pointer := range fkrVal.Data {
		_, exists := ids[*pointer.ID]
		if exists {
			return nil, errors.New(fmt.Sprintf(
				"Relationship '%s' invalid: Duplicate object in set.",
				fkr.Key,
			))
		}
		ids[*pointer.ID] = true
	}
	// extract ids
	var instanceIds []string
	for _, pointer := range fkrVal.Data {
		instanceIds = append(instanceIds, *pointer.ID)
	}

	// check that the objects exist
	model := c.GetServer().GetModel(fkr.Type)
	include := schema.Include{
		Children: map[string]*schema.Include{},
	}
	instances, _, err := c.GetObjectsByIDs(model, instanceIds, &include)
	if err != nil {
		panic(err)
	}
	if len(instances) < len(ids) {
		return nil, errors.New(fmt.Sprintf(
			"Relationship '%s' invalid: Not all objects in set exist",
			fkr.Key,
		))
	}
	// check that the user has access to the objects
	canAccess := c.CanAccessAllInstances(instances)
	if !canAccess {
		return nil, errors.New(fmt.Sprintf(
			"Relationship '%s' invalid: You do not have access to all objects in set.",
			fkr.Key,
		))
	}
	return fkrVal, nil
}
func (fkr *ForeignKeyReverse) ValidateUpdate(
	c schema.Context,
	newVal interface{},
	oldVal interface{},
) (
	interface{},
	error,
) {
	// extract new value
	newPointer, err := ParsePagePointerSet(newVal)
	if err != nil {
		return nil, err
	}
	// if not provided, return nothing
	if !newPointer.Provided {
		return nil, nil
	}
	// clean and check validity of new value
	valid, err := fkr.Validate(c, newPointer)
	if err != nil {
		return nil, err
	}
	validNewPointerSet := AssertPointerSet(valid)

	// extract old value
	oldValue := pointerSetFromPage(oldVal)

	// return nothing if no changes
	if validNewPointerSet.Equal(oldValue) {
		return nil, nil
	}
	// check if a value has been removed if not nullable
	if !fkr.Nullable {
		oldCounts := oldValue.Counts()
		newCounts := validNewPointerSet.Counts()
		for key, _ := range oldCounts {
			_, exists := newCounts[key]
			if !exists {
				return nil, errors.New(fmt.Sprintf(
					"Relationship '%s' invalid: Cannot remove item from non nullable reverse relation.",
					fkr.Key,
				))
			}
		}
	}
	// otherwise return new validated value
	return validNewPointerSet, nil
}

func (fkr *ForeignKeyReverse) GetInsertQueries(newId string, val interface{}) []schema.Query {
	fkrVal, ok := val.(PointerSet)
	if !ok {
		panic("Bad pointer set value")
	}

	spots := []string{}
	args := []interface{}{}
	for index, pointer := range fkrVal.Data {
		spots = append(spots, fmt.Sprintf("$%d", index+2))
		args = append(args, *pointer.ID)
	}

	if len(spots) == 0 {
		return []schema.Query{}
	}

	query := fmt.Sprintf(
		`update %s set %s = $1 where %s in (%s);`,
		fkr.SourceTable,
		fkr.ColumnName,
		fkr.SourceIDColumn,
		strings.Join(spots, ", "),
	)
	return []schema.Query{
		schema.Query{
			Query: query,
			Args:  append([]interface{}{newId}, args...),
		},
	}
}

func (fkr *ForeignKeyReverse) GetUpdateQueries(id string, oldVal interface{}, newVal interface{}) []schema.Query {
	oldSet := pointerSetFromPage(oldVal)
	newSet, ok := newVal.(PointerSet)
	if !ok {
		panic("Bad pointer set value")
	}

	queries := []schema.Query{}

	oldCount := oldSet.Counts()
	newCount := newSet.Counts()

	nulling := []string{}
	for key, _ := range oldCount {
		_, exists := newCount[key]
		if !exists {
			splitKey := strings.Split(key, ":")
			nulling = append(nulling, splitKey[1])
		}
	}
	if len(nulling) > 0 {
		spots := []string{}
		args := []interface{}{}
		for index, id := range nulling {
			spots = append(spots, fmt.Sprintf("$%d", index+1))
			args = append(args, id)
		}
		queries = append(queries, schema.Query{
			Query: fmt.Sprintf(
				"update %s set %s = null where %s in (%s)",
				fkr.SourceTable,
				fkr.ColumnName,
				fkr.SourceIDColumn,
				strings.Join(spots, ", "),
			),
			Args: args,
		})
	}

	adding := []string{}
	for key, _ := range newCount {
		_, exists := oldCount[key]
		if !exists {
			splitKey := strings.Split(key, ":")
			adding = append(adding, splitKey[1])
		}
	}
	if len(adding) > 0 {
		spots := []string{}
		args := []interface{}{}
		for index, id := range adding {
			spots = append(spots, fmt.Sprintf("$%d", index+1))
			args = append(args, id)
		}
		queries = append(queries, schema.Query{
			Query: fmt.Sprintf(
				"update %s set %s = %s where %s in (%s)",
				fkr.SourceTable,
				fkr.ColumnName,
				id,
				fkr.SourceIDColumn,
				strings.Join(spots, ", "),
			),
			Args: args,
		})
	}

	return queries
}

func AssertForeignKeyReverse(val interface{}) schema.Page {
	fkrVal, ok := val.(schema.Page)
	if !ok {
		panic("Bad foreign key reverse value")
	}
	return fkrVal
}
