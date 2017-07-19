package attributes

import (
	"fmt"
	"github.com/bor3ham/reja/schema"
	"strings"
)

type BoolNullFilter struct {
	*BaseFilter
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
	*BaseFilter
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
		b.Key + ISNULL_SUFFIX,
		b.Key,
	}
}
func (b Bool) ValidateFilters(queries map[string][]string) ([]schema.Filter, error) {
	valids := []schema.Filter{}

	// null check
	nullsOnly := false
	nonNullsOnly := false

	nullKey := b.Key + ISNULL_SUFFIX
	nullStrings, exists := queries[nullKey]
	if exists {
		if len(nullStrings) != 1 {
			return filterException(
				"Cannot null check attribute '%s' against more than one value.",
				b.Key,
			)
		}
		isNullString := strings.ToLower(nullStrings[0])
		if isNullString == "true" {
			nullsOnly = true
			valids = append(valids, BoolNullFilter{
				BaseFilter: &BaseFilter{
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
				BaseFilter: &BaseFilter{
					QArgKey:    nullKey,
					QArgValues: []string{"false"},
				},
				null:   false,
				column: b.ColumnName,
			})
		} else {
			return filterException(
				"Invalid null check value on attribute '%s'. Must be boolean.",
				b.Key,
			)
		}
	}

	exactKey := b.Key
	exactStrings, exists := queries[exactKey]
	if exists {
		if len(exactStrings) != 1 {
			return filterException(
				"Cannot compare attribute '%s' against more than one value.",
				b.Key,
			)
		}

		if nullsOnly {
			return filterException(
				"Cannot match attribute '%s' to an exact value and null.",
				b.Key,
			)
		}

		compareValue := strings.ToLower(exactStrings[0])
		if compareValue == "true" {
			valids = append(valids, BoolExactFilter{
				BaseFilter: &BaseFilter{
					QArgKey:    exactKey,
					QArgValues: []string{"true"},
				},
				value:  true,
				column: b.ColumnName,
			})
		} else if compareValue == "false" {
			valids = append(valids, BoolExactFilter{
				BaseFilter: &BaseFilter{
					QArgKey:    exactKey,
					QArgValues: []string{"false"},
				},
				value:  false,
				column: b.ColumnName,
			})
		} else {
			return filterException(
				"Invalid comparison value on attribute '%s'. Must be boolean.",
				b.Key,
			)
		}
	}

	return valids, nil
}
