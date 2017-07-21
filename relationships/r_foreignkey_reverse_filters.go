package relationships

import (
	"fmt"
	"strings"
	"github.com/bor3ham/reja/filters"
	"github.com/bor3ham/reja/schema"
)

type ForeignKeyReverseContainsFilter struct {
	*schema.BaseFilter
	columnName string
	sourceTable string
	sourceIDColumn string
	values []string
	exclude bool
}
func (f ForeignKeyReverseContainsFilter) GetWhere(
	c schema.Context,
	idColumn string,
	nextArg int,
) (
	[]string,
	[]interface{},
) {
	argSpots := []string{}
	argVals := []interface{}{}
	for index, value := range f.values {
		argSpots = append(argSpots, fmt.Sprintf("$%d", index + 1))
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

func (fkr ForeignKeyReverse) AvailableFilters() []interface{} {
	return []interface{}{
		filters.FilterDescription{
			Key:         fkr.Key + filters.CONTAINS_SUFFIX,
			Description: "Related items to filter for. One or more IDs.",
			Examples: []string{
				fmt.Sprintf("?%s=1", fkr.Key + filters.CONTAINS_SUFFIX),
				fmt.Sprintf("?%s=1&%s=2", fkr.Key + filters.CONTAINS_SUFFIX, fkr.Key + filters.CONTAINS_SUFFIX),
			},
		},
		filters.FilterDescription{
			Key:         fkr.Key + filters.EXCLUDES_SUFFIX,
			Description: "Related items to exclude. One or more IDs.",
			Examples: []string{
				fmt.Sprintf("?%s=1", fkr.Key + filters.EXCLUDES_SUFFIX),
				fmt.Sprintf("?%s=1&%s=2", fkr.Key + filters.EXCLUDES_SUFFIX, fkr.Key + filters.EXCLUDES_SUFFIX),
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
			columnName: fkr.ColumnName,
			sourceTable: fkr.SourceTable,
			sourceIDColumn: fkr.SourceIDColumn,
			values: compareValues,
			exclude: false,
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
			columnName: fkr.ColumnName,
			sourceTable: fkr.SourceTable,
			sourceIDColumn: fkr.SourceIDColumn,
			values: compareValues,
			exclude: true,
		})
	}

	return valids, nil
}
