package relationships

import (
	"fmt"
	"github.com/bor3ham/reja/filters"
	"github.com/bor3ham/reja/schema"
	"strconv"
	"strings"
)

type GenericForeignKeyReverseContainsFilter struct {
	*schema.BaseFilter

	table         string
	ownTypeColumn string
	ownIDColumn   string
	ownType       string
	otherIDColumn string

	values  []string
	exclude bool
}

func (f GenericForeignKeyReverseContainsFilter) GetWhere(
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
		argSpots = append(argSpots, fmt.Sprintf("$%d", index+2))
		argVals = append(argVals, value)
	}

	query := fmt.Sprintf(
		"select %s from %s where %s = $1 and %s in (%s)",
		f.ownIDColumn,
		f.table,
		f.ownTypeColumn,
		f.otherIDColumn,
		strings.Join(argSpots, ", "),
	)
	argVals = append([]interface{}{f.ownType}, argVals...)
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

type GenericForeignKeyReverseCountFilter struct {
	*schema.BaseFilter
	key string

	table         string
	ownTypeColumn string
	ownIDColumn   string
	ownType       string

	value    int
	operator string
}

func (f GenericForeignKeyReverseCountFilter) GetWhere(
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
					coalesce(countselect.count, 0) as total
				from %s
					left join (
						select
							%s,
							count(*) as count
						from
							%s
						where
							%s = $1
						group by
							%s
					)
					as countselect
					on countselect.%s = %s.%s
			) as totalSelect where totalSelect.total %s $2
		`,
		idColumn,
		idColumn,
		modelTable,
		f.ownIDColumn,
		f.table,
		f.ownTypeColumn,
		f.ownIDColumn,
		f.ownIDColumn,
		modelTable,
		idColumn,
		f.operator,
	)

	rows, err := c.Query(query, f.ownType, f.value)
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

func (gfkr GenericForeignKeyReverse) AvailableFilters() []interface{} {
	return []interface{}{
		filters.FilterDescription{
			Key:         gfkr.Key + filters.CONTAINS_SUFFIX,
			Description: "Related items to search for in set. One or more IDs.",
			Examples: []string{
				fmt.Sprintf("?%s=1", gfkr.Key+filters.CONTAINS_SUFFIX),
				fmt.Sprintf("?%s=1&%s=2", gfkr.Key+filters.CONTAINS_SUFFIX, gfkr.Key+filters.CONTAINS_SUFFIX),
			},
		},
		filters.FilterDescription{
			Key:         gfkr.Key + filters.EXCLUDES_SUFFIX,
			Description: "Related items to exclude if appearing in set. One or more IDs.",
			Examples: []string{
				fmt.Sprintf("?%s=1", gfkr.Key+filters.EXCLUDES_SUFFIX),
				fmt.Sprintf("?%s=1&%s=2", gfkr.Key+filters.EXCLUDES_SUFFIX, gfkr.Key+filters.EXCLUDES_SUFFIX),
			},
		},
		filters.FilterDescription{
			Key:         gfkr.Key + filters.COUNT_SUFFIX,
			Description: "Count of related items. Single value integer.",
			Examples: []string{
				fmt.Sprintf("?%s=5", gfkr.Key+filters.COUNT_SUFFIX),
			},
		},
		filters.FilterDescription{
			Key:         gfkr.Key + filters.COUNT_SUFFIX + filters.LT_SUFFIX,
			Description: "Maximum count of related items. Single value integer.",
			Examples: []string{
				fmt.Sprintf("?%s=5", gfkr.Key+filters.COUNT_SUFFIX+filters.LT_SUFFIX),
			},
		},
		filters.FilterDescription{
			Key:         gfkr.Key + filters.COUNT_SUFFIX + filters.GT_SUFFIX,
			Description: "Minimum count of related items. Single value integer.",
			Examples: []string{
				fmt.Sprintf("?%s=5", gfkr.Key+filters.COUNT_SUFFIX+filters.GT_SUFFIX),
			},
		},
	}
}
func (gfkr GenericForeignKeyReverse) ValidateFilters(queries map[string][]string) ([]schema.Filter, error) {
	valids := []schema.Filter{}

	containsKey := gfkr.Key + filters.CONTAINS_SUFFIX
	containsStrings, exists := queries[containsKey]
	if exists {
		compareValues := []string{}
		for _, value := range containsStrings {
			compareValues = append(compareValues, strings.ToLower(strings.TrimSpace(value)))
		}

		valids = append(valids, GenericForeignKeyReverseContainsFilter{
			BaseFilter: &schema.BaseFilter{
				QArgKey:    containsKey,
				QArgValues: compareValues,
			},

			table:         gfkr.Table,
			ownTypeColumn: gfkr.OwnTypeColumn,
			ownIDColumn:   gfkr.OwnIDColumn,
			ownType:       gfkr.OwnType,
			otherIDColumn: gfkr.OtherIDColumn,

			values:  compareValues,
			exclude: false,
		})
	}

	excludesKey := gfkr.Key + filters.EXCLUDES_SUFFIX
	excludesStrings, exists := queries[excludesKey]
	if exists {
		compareValues := []string{}
		for _, value := range excludesStrings {
			compareValues = append(compareValues, strings.ToLower(strings.TrimSpace(value)))
		}

		valids = append(valids, GenericForeignKeyReverseContainsFilter{
			BaseFilter: &schema.BaseFilter{
				QArgKey:    excludesKey,
				QArgValues: compareValues,
			},

			table:         gfkr.Table,
			ownTypeColumn: gfkr.OwnTypeColumn,
			ownIDColumn:   gfkr.OwnIDColumn,
			ownType:       gfkr.OwnType,
			otherIDColumn: gfkr.OtherIDColumn,

			values:  compareValues,
			exclude: true,
		})
	}

	exactCountKey := gfkr.Key + filters.COUNT_SUFFIX
	exactCountStrings, exists := queries[exactCountKey]
	if exists {
		if len(exactCountStrings) != 1 {
			return filters.Exception(
				"Cannot compare count of relationship '%s' to more than one value.",
				gfkr.Key,
			)
		}

		stringValue := exactCountStrings[0]
		intValue, err := strconv.Atoi(stringValue)
		if err != nil {
			return filters.Exception(
				"Invalid count comparison value on relationship '%s'. Must be integer.",
				gfkr.Key,
			)
		}
		valids = append(valids, GenericForeignKeyReverseCountFilter{
			BaseFilter: &schema.BaseFilter{
				QArgKey:    exactCountKey,
				QArgValues: []string{strconv.Itoa(intValue)},
			},

			table:         gfkr.Table,
			ownTypeColumn: gfkr.OwnTypeColumn,
			ownIDColumn:   gfkr.OwnIDColumn,
			ownType:       gfkr.OwnType,

			value:    intValue,
			operator: "=",
		})
	}

	lesserCountKey := gfkr.Key + filters.COUNT_SUFFIX + filters.LT_SUFFIX
	lesserCountStrings, exists := queries[lesserCountKey]
	if exists {
		if len(lesserCountStrings) != 1 {
			return filters.Exception(
				"Cannot compare count of relationship '%s' to more than one value.",
				gfkr.Key,
			)
		}

		stringValue := lesserCountStrings[0]
		intValue, err := strconv.Atoi(stringValue)
		if err != nil {
			return filters.Exception(
				"Invalid count comparison value on relationship '%s'. Must be integer.",
				gfkr.Key,
			)
		}
		valids = append(valids, GenericForeignKeyReverseCountFilter{
			BaseFilter: &schema.BaseFilter{
				QArgKey:    lesserCountKey,
				QArgValues: []string{strconv.Itoa(intValue)},
			},

			table:         gfkr.Table,
			ownTypeColumn: gfkr.OwnTypeColumn,
			ownIDColumn:   gfkr.OwnIDColumn,
			ownType:       gfkr.OwnType,

			value:    intValue,
			operator: "<",
		})
	}

	greaterCountKey := gfkr.Key + filters.COUNT_SUFFIX + filters.GT_SUFFIX
	greaterCountStrings, exists := queries[greaterCountKey]
	if exists {
		if len(greaterCountStrings) != 1 {
			return filters.Exception(
				"Cannot compare count of relationship '%s' to more than one value.",
				gfkr.Key,
			)
		}

		stringValue := greaterCountStrings[0]
		intValue, err := strconv.Atoi(stringValue)
		if err != nil {
			return filters.Exception(
				"Invalid count comparison value on relationship '%s'. Must be integer.",
				gfkr.Key,
			)
		}
		valids = append(valids, GenericForeignKeyReverseCountFilter{
			BaseFilter: &schema.BaseFilter{
				QArgKey:    greaterCountKey,
				QArgValues: []string{strconv.Itoa(intValue)},
			},

			table:         gfkr.Table,
			ownTypeColumn: gfkr.OwnTypeColumn,
			ownIDColumn:   gfkr.OwnIDColumn,
			ownType:       gfkr.OwnType,

			value:    intValue,
			operator: ">",
		})
	}

	return valids, nil
}
