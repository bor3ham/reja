package attributes

import (
	"fmt"
	"github.com/bor3ham/reja/schema"
	"github.com/bor3ham/reja/filters"
	"strings"
)

type BoolNullFilter struct {
	*schema.BaseFilter
	null   bool
	column string
}

func (f BoolNullFilter) GetWhereQueries(nextArg int) []string {
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
func (f BoolNullFilter) GetWhereArgs() []interface{} {
	return []interface{}{}
}

type BoolExactFilter struct {
	*schema.BaseFilter
	value  bool
	column string
}

func (f BoolExactFilter) GetWhereQueries(nextArg int) []string {
	return []string{
		fmt.Sprintf("%s = $%d", f.column, nextArg),
	}
}
func (f BoolExactFilter) GetWhereArgs() []interface{} {
	return []interface{}{
		f.value,
	}
}

func (b Bool) AvailableFilters() []string {
	return []string{
		b.Key,
		b.Key + filters.ISNULL_SUFFIX,
	}
}
func (b Bool) ValidateFilters(queries map[string][]string) ([]schema.Filter, error) {
	valids := []schema.Filter{}

	// null check
	nullsOnly := false
	nonNullsOnly := false

	nullKey := b.Key + filters.ISNULL_SUFFIX
	nullStrings, exists := queries[nullKey]
	if exists {
		if len(nullStrings) != 1 {
			return filters.Exception(
				"Cannot null check attribute '%s' against more than one value.",
				b.Key,
			)
		}
		isNullString := strings.ToLower(nullStrings[0])
		if isNullString == "true" {
			nullsOnly = true
			valids = append(valids, BoolNullFilter{
				BaseFilter: &schema.BaseFilter{
					QArgKey:    nullKey,
					QArgValues: []string{"true"},
				},
				null:   true,
				column: b.ColumnName,
			})
		} else if isNullString == "false" {
			nonNullsOnly = true
			_ = nonNullsOnly
			valids = append(valids, BoolNullFilter{
				BaseFilter: &schema.BaseFilter{
					QArgKey:    nullKey,
					QArgValues: []string{"false"},
				},
				null:   false,
				column: b.ColumnName,
			})
		} else {
			return filters.Exception(
				"Invalid null check value on attribute '%s'. Must be boolean.",
				b.Key,
			)
		}
	}

	exactKey := b.Key
	exactStrings, exists := queries[exactKey]
	if exists {
		if len(exactStrings) != 1 {
			return filters.Exception(
				"Cannot compare attribute '%s' against more than one value.",
				b.Key,
			)
		}

		if nullsOnly {
			return filters.Exception(
				"Cannot match attribute '%s' to an exact value and null.",
				b.Key,
			)
		}

		compareValue := strings.ToLower(exactStrings[0])
		if compareValue == "true" {
			valids = append(valids, BoolExactFilter{
				BaseFilter: &schema.BaseFilter{
					QArgKey:    exactKey,
					QArgValues: []string{"true"},
				},
				value:  true,
				column: b.ColumnName,
			})
		} else if compareValue == "false" {
			valids = append(valids, BoolExactFilter{
				BaseFilter: &schema.BaseFilter{
					QArgKey:    exactKey,
					QArgValues: []string{"false"},
				},
				value:  false,
				column: b.ColumnName,
			})
		} else {
			return filters.Exception(
				"Invalid comparison value on attribute '%s'. Must be boolean.",
				b.Key,
			)
		}
	}

	return valids, nil
}
