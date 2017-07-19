package attributes

import (
	"fmt"
	"github.com/bor3ham/reja/schema"
	"strconv"
	"strings"
)

type IntegerNullFilter struct {
	*BaseFilter
	null   bool
	column string
}

func (f IntegerNullFilter) GetWhereQueries(nextArg int) []string {
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
func (f IntegerNullFilter) GetWhereArgs() []interface{} {
	return []interface{}{}
}

type IntegerExactFilter struct {
	*BaseFilter
	value  int
	column string
}

func (f IntegerExactFilter) GetWhereQueries(nextArg int) []string {
	return []string{
		fmt.Sprintf("%s = $%d", f.column, nextArg),
	}
}
func (f IntegerExactFilter) GetWhereArgs() []interface{} {
	return []interface{}{
		f.value,
	}
}

type IntegerLesserFilter struct {
	*BaseFilter
	value  int
	column string
}

func (f IntegerLesserFilter) GetWhereQueries(nextArg int) []string {
	return []string{
		fmt.Sprintf("%s < $%d", f.column, nextArg),
	}
}
func (f IntegerLesserFilter) GetWhereArgs() []interface{} {
	return []interface{}{
		f.value,
	}
}

type IntegerGreaterFilter struct {
	*BaseFilter
	value  int
	column string
}

func (f IntegerGreaterFilter) GetWhereQueries(nextArg int) []string {
	return []string{
		fmt.Sprintf("%s > $%d", f.column, nextArg),
	}
}
func (f IntegerGreaterFilter) GetWhereArgs() []interface{} {
	return []interface{}{
		f.value,
	}
}

func (i Integer) AvailableFilters() []string {
	return []string{
		i.Key,
		i.Key + ISNULL_SUFFIX,
		i.Key + LT_SUFFIX,
		i.Key + GT_SUFFIX,
	}
}
func (i Integer) ValidateFilters(queries map[string][]string) ([]schema.Filter, error) {
	valids := []schema.Filter{}

	// null check
	nullsOnly := false
	nonNullsOnly := false

	nullKey := i.Key + ISNULL_SUFFIX
	nullStrings, exists := queries[nullKey]
	if exists {
		if len(nullStrings) != 1 {
			return filterException(
				"Cannot null check attribute '%s' against more than one value.",
				i.Key,
			)
		}
		isNullString := strings.ToLower(nullStrings[0])
		if isNullString == "true" {
			nullsOnly = true
			valids = append(valids, IntegerNullFilter{
				BaseFilter: &BaseFilter{
					QArgKey:    nullKey,
					QArgValues: []string{"true"},
				},
				null:   true,
				column: i.ColumnName,
			})
		} else if isNullString == "false" {
			nonNullsOnly = true
			_ = nonNullsOnly
			valids = append(valids, IntegerNullFilter{
				BaseFilter: &BaseFilter{
					QArgKey:    nullKey,
					QArgValues: []string{"false"},
				},
				null:   false,
				column: i.ColumnName,
			})
		} else {
			return filterException(
				"Invalid null check value on attribute '%s'. Must be boolean.",
				i.Key,
			)
		}
	}

	exactKey := i.Key
	exactStrings, exists := queries[exactKey]
	if exists {
		if len(exactStrings) != 1 {
			return filterException(
				"Cannot compare attribute '%s' against more than one value.",
				i.Key,
			)
		}

		if nullsOnly {
			return filterException(
				"Cannot match attribute '%s' to an exact value and null.",
				i.Key,
			)
		}

		compareValue, err := strconv.Atoi(exactStrings[0])
		if err == nil {
			valids = append(valids, IntegerExactFilter{
				BaseFilter: &BaseFilter{
					QArgKey:    exactKey,
					QArgValues: []string{strconv.Itoa(compareValue)},
				},
				value:  compareValue,
				column: i.ColumnName,
			})
		} else {
			return filterException(
				"Invalid exact value on attribute '%s'. Must be integer.",
				i.Key,
			)
		}
	}

	filteringLesser := false
	var filteringLesserValue int

	lesserKey := i.Key + LT_SUFFIX
	lesserStrings, exists := queries[lesserKey]
	if exists {
		if len(lesserStrings) != 1 {
			return filterException(
				"Cannot compare attribute '%s' to be lesser than more than one value.",
				i.Key,
			)
		}

		if nullsOnly {
			return filterException(
				"Cannot compare attribute '%s' to a value and null.",
				i.Key,
			)
		}

		lesserValue, err := strconv.Atoi(lesserStrings[0])
		if err == nil {
			filteringLesser = true
			filteringLesserValue = lesserValue

			valids = append(valids, IntegerLesserFilter{
				BaseFilter: &BaseFilter{
					QArgKey:    lesserKey,
					QArgValues: []string{strconv.Itoa(lesserValue)},
				},
				value:  lesserValue,
				column: i.ColumnName,
			})
		} else {
			return filterException(
				"Invalid lesser than comparison value on attribute '%s'. Must be integer.",
				i.Key,
			)
		}
	}

	greaterKey := i.Key + GT_SUFFIX
	greaterStrings, exists := queries[greaterKey]
	if exists {
		if len(greaterStrings) != 1 {
			return filterException(
				"Cannot compare attribute '%s' to be greater than more than one value.",
				i.Key,
			)
		}

		if nullsOnly {
			return filterException(
				"Cannot compare attribute '%s' to a value and null.",
				i.Key,
			)
		}

		greaterValue, err := strconv.Atoi(greaterStrings[0])
		if err == nil {
			if filteringLesser && greaterValue > filteringLesserValue {
				return filterException(
					"Cannot compare attribute '%s' to value greater than additional lesser than filter.",
					i.Key,
				)
			}

			valids = append(valids, IntegerGreaterFilter{
				BaseFilter: &BaseFilter{
					QArgKey:    greaterKey,
					QArgValues: []string{strconv.Itoa(greaterValue)},
				},
				value:  greaterValue,
				column: i.ColumnName,
			})
		} else {
			return filterException(
				"Invalid greater than comparison value on attribute '%s'. Must be integer.",
				i.Key,
			)
		}
	}

	return valids, nil
}
