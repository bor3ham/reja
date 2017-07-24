package relationships

import (
	"fmt"
	"strings"
	"strconv"
	"github.com/bor3ham/reja/filters"
	"github.com/bor3ham/reja/schema"
)

type ManyToManyContainsFilter struct {
	*schema.BaseFilter

	table string
	ownIDColumn string
	otherIDColumn string

	values []string
	exclude bool
}
func (f ManyToManyContainsFilter) GetWhere(
	c schema.Context,
	modelTable string,
	idColumn string,
	nextArg int,
) (
	[]string,
	[]interface{},
) {
	argSpots := []string{}
	argVals := []interface{}{}
	for index, value := range f.values {
		argSpots = append(argSpots, fmt.Sprintf("$%d", index+1))
		argVals = append(argVals, value)
	}

	query := fmt.Sprintf(
		"select %s from %s where %s in (%s)",
		f.ownIDColumn,
		f.table,
		f.otherIDColumn,
		strings.Join(argSpots, ", "),
	)
	rows, err := c.Query(query, argVals...)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	ids := []string{}
	for rows.Next() {
		var id string
		err := rows.Scan(&id)
		if err != nil {
			panic(err)
		}
		ids = append(ids, id)
	}

	if f.exclude {
		if len(ids) > 0 {
			return []string{
				fmt.Sprintf("%s not in (%s)", idColumn, strings.Join(ids, ", ")),
			}, []interface{}{}
		} else {
			return []string{}, []interface{}{}
		}
	} else {
		if len(ids) > 0 {
			return []string{
				fmt.Sprintf("%s in (%s)", idColumn, strings.Join(ids, ", ")),
			}, []interface{}{}
		} else {
			return []string{"true is false"}, []interface{}{}
		}
	}
}

type ManyToManyCountFilter struct {
	*schema.BaseFilter

	table string
	ownIDColumn string
	otherIDColumn string
	key string

	value int
	operator string
}
func (f ManyToManyCountFilter) GetWhere(
	c schema.Context,
	modelTable string,
	idColumn string,
	nextArg int,
) (
	[]string,
	[]interface{},
) {
	query := fmt.Sprintf(
		`
			select %s from (
				select
					%s,
					coalesce(countSelect.count, 0) as total
				from
					%s
				left join (
					select
						%s,
						count(%s) as count
					from
						%s
					group by
						%s
				)
				as countSelect
				on countSelect.%s = %s.%s
			) as totalSelect where totalSelect.total %s $1
		`,
		idColumn,
		idColumn,
		modelTable,
		f.ownIDColumn,
		f.otherIDColumn,
		f.table,
		f.ownIDColumn,
		f.ownIDColumn,
		modelTable,
		idColumn,
		f.operator,
	)

	rows, err := c.Query(query, f.value)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	ids := []string{}
	for rows.Next() {
		var id string
		err := rows.Scan(&id)
		if err != nil {
			panic(err)
		}
		ids = append(ids, id)
	}

	if len(ids) > 0 {
		return []string{
			fmt.Sprintf("%s in (%s)", idColumn, strings.Join(ids, ", ")),
		}, []interface{}{}
	} else {
		return []string{"true is false"}, []interface{}{}
	}
}

func (m2m ManyToMany) ValidateFilters(queries map[string][]string) ([]schema.Filter, error) {
	valids := []schema.Filter{}

	containsKey := m2m.Key + filters.CONTAINS_SUFFIX
	containsStrings, exists := queries[containsKey]
	if exists {
		compareValues := []string{}
		for _, value := range containsStrings {
			compareValues = append(compareValues, strings.ToLower(strings.TrimSpace(value)))
		}

		valids = append(valids, ManyToManyContainsFilter{
			BaseFilter: &schema.BaseFilter{
				QArgKey:    containsKey,
				QArgValues: compareValues,
			},

			table: m2m.Table,
			ownIDColumn: m2m.OwnIDColumn,
			otherIDColumn: m2m.OtherIDColumn,

			values:         compareValues,
			exclude:        false,
		})
	}

	excludesKey := m2m.Key + filters.EXCLUDES_SUFFIX
	excludesStrings, exists := queries[excludesKey]
	if exists {
		compareValues := []string{}
		for _, value := range excludesStrings {
			compareValues = append(compareValues, strings.ToLower(strings.TrimSpace(value)))
		}

		valids = append(valids, ManyToManyContainsFilter{
			BaseFilter: &schema.BaseFilter{
				QArgKey:    excludesKey,
				QArgValues: compareValues,
			},

			table: m2m.Table,
			ownIDColumn: m2m.OwnIDColumn,
			otherIDColumn: m2m.OtherIDColumn,

			values:         compareValues,
			exclude:        true,
		})
	}

	exactCountKey := m2m.Key + filters.COUNT_SUFFIX
	exactCountStrings, exists := queries[exactCountKey]
	if exists {
		if len(exactCountStrings) != 1 {
			return filters.Exception(
				"Cannot compare count of relationship '%s' to more than one value.",
				m2m.Key,
			)
		}

		stringValue := exactCountStrings[0]
		intValue, err := strconv.Atoi(stringValue)
		if err != nil {
			return filters.Exception(
				"Invalid count comparison value on relationship '%s'. Must be integer.",
				m2m.Key,
			)
		}
		valids = append(valids, ManyToManyCountFilter{
			BaseFilter: &schema.BaseFilter{
				QArgKey:    exactCountKey,
				QArgValues: []string{strconv.Itoa(intValue)},
			},

			table: m2m.Table,
			ownIDColumn: m2m.OwnIDColumn,
			otherIDColumn: m2m.OtherIDColumn,
			key:            m2m.Key,

			value:          intValue,
			operator:       "=",
		})
	}

	lesserCountKey := m2m.Key + filters.COUNT_SUFFIX + filters.LT_SUFFIX
	lesserCountStrings, exists := queries[lesserCountKey]
	if exists {
		if len(lesserCountStrings) != 1 {
			return filters.Exception(
				"Cannot compare count of relationship '%s' to more than one value.",
				m2m.Key,
			)
		}

		stringValue := lesserCountStrings[0]
		intValue, err := strconv.Atoi(stringValue)
		if err != nil {
			return filters.Exception(
				"Invalid count comparison value on relationship '%s'. Must be integer.",
				m2m.Key,
			)
		}
		valids = append(valids, ManyToManyCountFilter{
			BaseFilter: &schema.BaseFilter{
				QArgKey:    lesserCountKey,
				QArgValues: []string{strconv.Itoa(intValue)},
			},

			table: m2m.Table,
			ownIDColumn: m2m.OwnIDColumn,
			otherIDColumn: m2m.OtherIDColumn,
			key:            m2m.Key,

			value:          intValue,
			operator:       "<",
		})
	}

	greaterCountKey := m2m.Key + filters.COUNT_SUFFIX + filters.GT_SUFFIX
	greaterCountStrings, exists := queries[greaterCountKey]
	if exists {
		if len(greaterCountStrings) != 1 {
			return filters.Exception(
				"Cannot compare count of relationship '%s' to more than one value.",
				m2m.Key,
			)
		}

		stringValue := greaterCountStrings[0]
		intValue, err := strconv.Atoi(stringValue)
		if err != nil {
			return filters.Exception(
				"Invalid count comparison value on relationship '%s'. Must be integer.",
				m2m.Key,
			)
		}
		valids = append(valids, ManyToManyCountFilter{
			BaseFilter: &schema.BaseFilter{
				QArgKey:    greaterCountKey,
				QArgValues: []string{strconv.Itoa(intValue)},
			},

			table: m2m.Table,
			ownIDColumn: m2m.OwnIDColumn,
			otherIDColumn: m2m.OtherIDColumn,
			key:            m2m.Key,

			value:          intValue,
			operator:       ">",
		})
	}

	return valids, nil
}
