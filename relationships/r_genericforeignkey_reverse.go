package relationships

import (
	"errors"
	"fmt"
	"github.com/bor3ham/reja/schema"
	"github.com/bor3ham/reja/utils"
	"strings"
)

type GenericForeignKeyReverse struct {
	RelationshipStub
	Key           string
	Table         string
	OwnType       string
	OwnTypeColumn string
	OwnIDColumn   string
	OtherIDColumn string
	OtherType     string
	Default       func(schema.Context, interface{}) PointerSet
	Nullable      bool
}

func (gfkr GenericForeignKeyReverse) GetKey() string {
	return gfkr.Key
}
func (gfkr GenericForeignKeyReverse) GetType() string {
	return gfkr.OtherType
}

func (gfkr GenericForeignKeyReverse) GetDefaultValue() interface{} {
	return schema.Page{}
}
func (gfkr GenericForeignKeyReverse) GetValues(
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

	server := c.GetServer()
	otherModel := server.GetModel(gfkr.OtherType)
	order, _, err := otherModel.GetOrderQuery(otherModel.DefaultOrder)
	if err != nil {
		panic(err)
	}

	spots := []string{}
	args := []interface{}{}
	for index, id := range ids {
		spots = append(spots, fmt.Sprintf("$%d", index+2))
		args = append(args, id)
	}
	idFilter := fmt.Sprintf("%s in (%s)", gfkr.OwnIDColumn, strings.Join(spots, ", "))
	typeFilter := fmt.Sprintf("%s = $1", gfkr.OwnTypeColumn)
	query := fmt.Sprintf(
		`
			select
				%s,
				%s
			from %s
			where (%s and %s)
			%s
	    `,
		gfkr.OtherIDColumn,
		gfkr.OwnIDColumn,
		gfkr.Table,
		idFilter,
		typeFilter,
		order,
	)
	rows, err := c.Query(query, append([]interface{}{gfkr.OwnType}, args...)...)
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
				Type: gfkr.OtherType,
			})
			value.Metadata["count"] = count
		}
		total += 1
		value.Metadata["total"] = total
		// update the value
		values[ownId] = value

		// add to maps
		_, exists = maps[ownId]
		if !exists {
			maps[ownId] = map[string][]string{}
			maps[ownId][gfkr.OtherType] = []string{}
		}
		maps[ownId][gfkr.OtherType] = append(maps[ownId][gfkr.OtherType], otherId)
	}
	// create the links
	for id, value := range values {
		total, ok := value.Metadata["total"].(int)
		if !ok {
			panic("Bad total received")
		}
		value.Links = utils.GetPaginationLinks(
			relationLink(c, m.Type, id, gfkr.Key),
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

func (gfkr *GenericForeignKeyReverse) DefaultFallback(
	c schema.Context,
	val interface{},
	instance interface{},
) (
	interface{},
	error,
) {
	var gfkrVal PointerSet
	if val == nil {
		gfkrVal = PointerSet{Provided: false}
	} else {
		var err error
		gfkrVal, err = ParsePagePointerSet(val)
		if err != nil {
			return nil, err
		}
	}

	if !gfkrVal.Provided {
		if gfkr.Default != nil {
			return gfkr.Default(c, instance), nil
		}
		return nil, nil
	}
	return gfkrVal, nil
}
func (gfkr *GenericForeignKeyReverse) Validate(
	c schema.Context,
	val interface{},
) (
	interface{},
	error,
) {
	gfkrVal := AssertPointerSet(val)

	// validate the types are correct
	for _, pointer := range gfkrVal.Data {
		if pointer.Type != gfkr.OtherType {
			return nil, errors.New(fmt.Sprintf(
				"Relationship '%s' invalid: Incorrect type in set.",
				gfkr.Key,
			))
		}
	}
	// find duplicates
	ids := map[string]bool{}
	for _, pointer := range gfkrVal.Data {
		_, exists := ids[*pointer.ID]
		if exists {
			return nil, errors.New(fmt.Sprintf(
				"Relationship '%s' invalid: Duplicate object in set.",
				gfkr.Key,
			))
		}
		ids[*pointer.ID] = true
	}
	// extract ids
	var instanceIds []string
	for _, pointer := range gfkrVal.Data {
		instanceIds = append(instanceIds, *pointer.ID)
	}

	// check that the objects exist
	model := c.GetServer().GetModel(gfkr.OtherType)
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
			gfkr.Key,
		))
	}
	// check that the user has access to the objects
	canAccess := c.CanAccessAllInstances(instances)
	if !canAccess {
		return nil, errors.New(fmt.Sprintf(
			"Relationship '%s' invalid: You do not have access to all objects in set.",
			gfkr.Key,
		))
	}
	return gfkrVal, nil
}
func (gfkr *GenericForeignKeyReverse) ValidateUpdate(
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
	valid, err := gfkr.Validate(c, newPointer)
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
	if !gfkr.Nullable {
		oldCounts := oldValue.Counts()
		newCounts := validNewPointerSet.Counts()
		for key, _ := range oldCounts {
			_, exists := newCounts[key]
			if !exists {
				return nil, errors.New(fmt.Sprintf(
					"Relationship '%s' invalid: Cannot remove item from non nullable reverse relation.",
					gfkr.Key,
				))
			}
		}
	}
	// otherwise return new validated value
	return validNewPointerSet, nil
}

func (gfkr *GenericForeignKeyReverse) GetInsertQueries(
	newId string,
	val interface{},
) []schema.Query {
	gfkrVal, ok := val.(PointerSet)
	if !ok {
		panic("Bad pointer set value")
	}

	spots := []string{}
	args := []interface{}{}
	for index, pointer := range gfkrVal.Data {
		spots = append(spots, fmt.Sprintf("$%d", index+3))
		args = append(args, *pointer.ID)
	}

	if len(spots) == 0 {
		return []schema.Query{}
	}

	query := fmt.Sprintf(
		`update %s set (%s, %s) = ($1, $2) where %s in (%s);`,
		gfkr.Table,
		gfkr.OwnTypeColumn,
		gfkr.OwnIDColumn,
		gfkr.OtherIDColumn,
		strings.Join(spots, ", "),
	)
	return []schema.Query{
		schema.Query{
			Query: query,
			Args: append([]interface{}{
				gfkr.OwnType,
				newId,
			}, args...),
		},
	}
}

func (gfkr *GenericForeignKeyReverse) GetUpdateQueries(id string, oldVal interface{}, newVal interface{}) []schema.Query {
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
				"update %s set (%s = null, %s = null) where %s in (%s)",
				gfkr.Table,
				gfkr.OwnIDColumn,
				gfkr.OwnTypeColumn,
				gfkr.OtherIDColumn,
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
			spots = append(spots, fmt.Sprintf("$%d", index+2))
			args = append(args, id)
		}
		queries = append(queries, schema.Query{
			Query: fmt.Sprintf(
				"update %s set %s = %s, %s = $1 where %s in (%s)",
				gfkr.Table,
				gfkr.OwnIDColumn,
				id,
				gfkr.OwnTypeColumn,
				gfkr.OtherIDColumn,
				strings.Join(spots, ", "),
			),
			Args: append([]interface{}{
				gfkr.OwnType,
			}, args...),
		})
	}

	return queries
}

func AssertGenericForeignKeyReverse(val interface{}) schema.Page {
	gfkrVal, ok := val.(schema.Page)
	if !ok {
		panic("Bad generic foreign key reverse value")
	}
	return gfkrVal
}
