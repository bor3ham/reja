package relationships

import (
	"fmt"
	"strings"
	"github.com/bor3ham/reja/schema"
	"github.com/bor3ham/reja/filters"
)

type ForeignKeyNullFilter struct {
	*schema.BaseFilter
	null   bool
	column string
}
func (f ForeignKeyNullFilter) GetWhereQueries(nextArg int) []string {
	if f.null {
		return []string{
			fmt.Sprintf("%s is null", f.column),
		}
	} else {
		return []string{
			fmt.Sprintf("%s is not null", f.column),
		}
	}
}
func (f ForeignKeyNullFilter) GetWhereArgs() []interface{} {
	return []interface{}{}
}

type ForeignKeyExactFilter struct {
	*schema.BaseFilter
	values  []string
	column string
}
func (f ForeignKeyExactFilter) GetWhereQueries(nextArg int) []string {
	args := []string{}
	for _, _ = range f.values {
		args = append(args, fmt.Sprintf("$%d", nextArg))
		nextArg += 1
	}
	return []string{
		fmt.Sprintf("%s in (%s)", f.column, strings.Join(args, ", ")),
	}
}
func (f ForeignKeyExactFilter) GetWhereArgs() []interface{} {
	args := []interface{}{}
	for _, value := range f.values {
		args = append(args, value)
	}
	return args
}

func (fk ForeignKey) AvailableFilters() []string {
	return []string{
		fk.Key,
		fk.Key + filters.ISNULL_SUFFIX,
	}
}
func (fk ForeignKey) ValidateFilters(queries map[string][]string) ([]schema.Filter, error) {
	valids := []schema.Filter{}

	// null check
	nullsOnly := false
	nonNullsOnly := false

	nullKey := fk.Key + filters.ISNULL_SUFFIX
	nullStrings, exists := queries[nullKey]
	if exists {
		if len(nullStrings) != 1 {
			return filters.Exception(
				"Cannot null check attribute '%s' against more than one value.",
				fk.Key,
			)
		}
		isNullString := strings.ToLower(nullStrings[0])
		if isNullString == "true" {
			nullsOnly = true
			valids = append(valids, ForeignKeyNullFilter{
				BaseFilter: &schema.BaseFilter{
					QArgKey:    nullKey,
					QArgValues: []string{"true"},
				},
				null:   true,
				column: fk.ColumnName,
			})
		} else if isNullString == "false" {
			nonNullsOnly = true
			_ = nonNullsOnly
			valids = append(valids, ForeignKeyNullFilter{
				BaseFilter: &schema.BaseFilter{
					QArgKey:    nullKey,
					QArgValues: []string{"false"},
				},
				null:   false,
				column: fk.ColumnName,
			})
		} else {
			return filters.Exception(
				"Invalid null check value on attribute '%s'. Must be boolean.",
				fk.Key,
			)
		}
	}

	exactKey := fk.Key
	exactStrings, exists := queries[exactKey]
	if exists {
		if nullsOnly {
			return filters.Exception(
				"Cannot match attribute '%s' to an exact value and null.",
				fk.Key,
			)
		}

		compareValues := []string{}
		for _, value := range exactStrings {
			compareValues = append(compareValues, strings.ToLower(strings.TrimSpace(value)))
		}

		valids = append(valids, ForeignKeyExactFilter{
			BaseFilter: &schema.BaseFilter{
				QArgKey:    exactKey,
				QArgValues: compareValues,
			},
			values:  compareValues,
			column: fk.ColumnName,
		})
	}

	return valids, nil
}
