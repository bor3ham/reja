package relationships

import (
	"fmt"
	"github.com/bor3ham/reja/filters"
	"github.com/bor3ham/reja/schema"
	"strconv"
	"strings"
)

type ForeignKeyReverseContainsFilter struct {
	*schema.BaseFilter
	columnName     string
	sourceTable    string
	sourceIDColumn string
	values         []string
	exclude        bool
}

func (f ForeignKeyReverseContainsFilter) GetWhere(
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
		f.columnName,
		f.sourceTable,
		f.sourceIDColumn,
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

type ForeignKeyReverseCountFilter struct {
	*schema.BaseFilter
	key         string
	columnName  string
	sourceTable string
	value       int
	operator    string
}

func (f ForeignKeyReverseCountFilter) GetWhere(
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
						group by
							%s
					)
					as countselect
					on countselect.%s = %s.%s
			) as totalSelect where totalSelect.total %s $1
		`,
		idColumn,
		idColumn,
		modelTable,
		f.columnName,
		f.sourceTable,
		f.columnName,
		f.columnName,
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

func (fkr ForeignKeyReverse) AvailableFilters() []interface{} {
	return []interface{}{
		filters.FilterDescription{
			Key:         fkr.Key + filters.CONTAINS_SUFFIX,
			Description: "Related items to search for in set. One or more IDs.",
			Examples: []string{
				fmt.Sprintf("?%s=1", fkr.Key+filters.CONTAINS_SUFFIX),
				fmt.Sprintf("?%s=1&%s=2", fkr.Key+filters.CONTAINS_SUFFIX, fkr.Key+filters.CONTAINS_SUFFIX),
			},
		},
		filters.FilterDescription{
			Key:         fkr.Key + filters.EXCLUDES_SUFFIX,
			Description: "Related items to exclude if appearing in set. One or more IDs.",
			Examples: []string{
				fmt.Sprintf("?%s=1", fkr.Key+filters.EXCLUDES_SUFFIX),
				fmt.Sprintf("?%s=1&%s=2", fkr.Key+filters.EXCLUDES_SUFFIX, fkr.Key+filters.EXCLUDES_SUFFIX),
			},
		},
		filters.FilterDescription{
			Key:         fkr.Key + filters.COUNT_SUFFIX,
			Description: "Count of related items. Single value integer.",
			Examples: []string{
				fmt.Sprintf("?%s=5", fkr.Key+filters.COUNT_SUFFIX),
			},
		},
		filters.FilterDescription{
			Key:         fkr.Key + filters.COUNT_SUFFIX + filters.LT_SUFFIX,
			Description: "Maximum count of related items. Single value integer.",
			Examples: []string{
				fmt.Sprintf("?%s=5", fkr.Key+filters.COUNT_SUFFIX+filters.LT_SUFFIX),
			},
		},
		filters.FilterDescription{
			Key:         fkr.Key + filters.COUNT_SUFFIX + filters.GT_SUFFIX,
			Description: "Minimum count of related items. Single value integer.",
			Examples: []string{
				fmt.Sprintf("?%s=5", fkr.Key+filters.COUNT_SUFFIX+filters.GT_SUFFIX),
			},
		},
	}
}
func (fkr ForeignKeyReverse) ValidateFilters(queries map[string][]string) ([]schema.Filter, error) {
	valids := []schema.Filter{}

	containsKey := fkr.Key + filters.CONTAINS_SUFFIX
	containsStrings, exists := queries[containsKey]
	if exists {
		compareValues := []string{}
		for _, value := range containsStrings {
			compareValues = append(compareValues, strings.ToLower(strings.TrimSpace(value)))
		}

		valids = append(valids, ForeignKeyReverseContainsFilter{
			BaseFilter: &schema.BaseFilter{
				QArgKey:    containsKey,
				QArgValues: compareValues,
			},
			columnName:     fkr.ColumnName,
			sourceTable:    fkr.SourceTable,
			sourceIDColumn: fkr.SourceIDColumn,
			values:         compareValues,
			exclude:        false,
		})
	}

	excludesKey := fkr.Key + filters.EXCLUDES_SUFFIX
	excludesStrings, exists := queries[excludesKey]
	if exists {
		compareValues := []string{}
		for _, value := range excludesStrings {
			compareValues = append(compareValues, strings.ToLower(strings.TrimSpace(value)))
		}

		valids = append(valids, ForeignKeyReverseContainsFilter{
			BaseFilter: &schema.BaseFilter{
				QArgKey:    excludesKey,
				QArgValues: compareValues,
			},
			columnName:     fkr.ColumnName,
			sourceTable:    fkr.SourceTable,
			sourceIDColumn: fkr.SourceIDColumn,
			values:         compareValues,
			exclude:        true,
		})
	}

	exactCountKey := fkr.Key + filters.COUNT_SUFFIX
	exactCountStrings, exists := queries[exactCountKey]
	if exists {
		if len(exactCountStrings) != 1 {
			return filters.Exception(
				"Cannot compare count of relationship '%s' to more than one value.",
				fkr.Key,
			)
		}

		stringValue := exactCountStrings[0]
		intValue, err := strconv.Atoi(stringValue)
		if err != nil {
			return filters.Exception(
				"Invalid count comparison value on relationship '%s'. Must be integer.",
				fkr.Key,
			)
		}
		valids = append(valids, ForeignKeyReverseCountFilter{
			BaseFilter: &schema.BaseFilter{
				QArgKey:    exactCountKey,
				QArgValues: []string{strconv.Itoa(intValue)},
			},
			key:         fkr.Key,
			columnName:  fkr.ColumnName,
			sourceTable: fkr.SourceTable,
			value:       intValue,
			operator:    "=",
		})
	}

	lesserCountKey := fkr.Key + filters.COUNT_SUFFIX + filters.LT_SUFFIX
	lesserCountStrings, exists := queries[lesserCountKey]
	if exists {
		if len(lesserCountStrings) != 1 {
			return filters.Exception(
				"Cannot compare count of relationship '%s' to more than one value.",
				fkr.Key,
			)
		}

		stringValue := lesserCountStrings[0]
		intValue, err := strconv.Atoi(stringValue)
		if err != nil {
			return filters.Exception(
				"Invalid count comparison value on relationship '%s'. Must be integer.",
				fkr.Key,
			)
		}
		valids = append(valids, ForeignKeyReverseCountFilter{
			BaseFilter: &schema.BaseFilter{
				QArgKey:    lesserCountKey,
				QArgValues: []string{strconv.Itoa(intValue)},
			},
			key:         fkr.Key,
			columnName:  fkr.ColumnName,
			sourceTable: fkr.SourceTable,
			value:       intValue,
			operator:    "<",
		})
	}

	greaterCountKey := fkr.Key + filters.COUNT_SUFFIX + filters.GT_SUFFIX
	greaterCountStrings, exists := queries[greaterCountKey]
	if exists {
		if len(greaterCountStrings) != 1 {
			return filters.Exception(
				"Cannot compare count of relationship '%s' to more than one value.",
				fkr.Key,
			)
		}

		stringValue := greaterCountStrings[0]
		intValue, err := strconv.Atoi(stringValue)
		if err != nil {
			return filters.Exception(
				"Invalid count comparison value on relationship '%s'. Must be integer.",
				fkr.Key,
			)
		}
		valids = append(valids, ForeignKeyReverseCountFilter{
			BaseFilter: &schema.BaseFilter{
				QArgKey:    greaterCountKey,
				QArgValues: []string{strconv.Itoa(intValue)},
			},
			key:         fkr.Key,
			columnName:  fkr.ColumnName,
			sourceTable: fkr.SourceTable,
			value:       intValue,
			operator:    ">",
		})
	}

	return valids, nil
}
