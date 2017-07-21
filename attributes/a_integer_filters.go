package attributes

import (
	"fmt"
	"github.com/bor3ham/reja/filters"
	"github.com/bor3ham/reja/schema"
	"strconv"
	"strings"
)

type IntegerNullFilter struct {
	*schema.BaseFilter
	null   bool
	column string
}

func (f IntegerNullFilter) GetWhere(
	c schema.Context,
	idColumn string,
	nextArg int,
) (
	[]string,
	[]interface{},
) {
	if f.null {
		return []string{
			fmt.Sprintf("%s is null", f.column),
		}, []interface{}{}
	} else {
		return []string{
			fmt.Sprintf("%s is not null", f.column),
		}, []interface{}{}
	}
}

type IntegerExactFilter struct {
	*schema.BaseFilter
	value  int
	column string
}

func (f IntegerExactFilter) GetWhere(
	c schema.Context,
	idColumn string,
	nextArg int,
) (
	[]string,
	[]interface{},
) {
	return []string{
			fmt.Sprintf("%s = $%d", f.column, nextArg),
		}, []interface{}{
			f.value,
		}
}

type IntegerLesserFilter struct {
	*schema.BaseFilter
	value  int
	column string
	lesser bool
}

func (f IntegerLesserFilter) GetWhere(
	c schema.Context,
	idColumn string,
	nextArg int,
) (
	[]string,
	[]interface{},
) {
	operator := "<"
	if !f.lesser {
		operator = ">"
	}
	return []string{
			fmt.Sprintf("%s %s $%d", f.column, operator, nextArg),
		}, []interface{}{
			f.value,
		}
}

func (i Integer) AvailableFilters() []interface{} {
	return []interface{}{
		filters.FilterDescription{
			Key:         i.Key,
			Description: "Exact match on integer value. Single value integer.",
			Examples: []string{
				fmt.Sprintf("?%s=1", i.Key),
			},
		},
		filters.FilterDescription{
			Key:         i.Key + filters.ISNULL_SUFFIX,
			Description: "Whether integer value exists. Single value boolean.",
			Examples: []string{
				fmt.Sprintf("?%s=true", i.Key+filters.ISNULL_SUFFIX),
				fmt.Sprintf("?%s=false", i.Key+filters.ISNULL_SUFFIX),
			},
		},
		filters.FilterDescription{
			Key:         i.Key + filters.LT_SUFFIX,
			Description: "Any value less than given integer. Single value integer.",
			Examples: []string{
				fmt.Sprintf("?%s=5", i.Key+filters.LT_SUFFIX),
			},
		},
		filters.FilterDescription{
			Key:         i.Key + filters.GT_SUFFIX,
			Description: "Any value greater than given integer. Single value integer.",
			Examples: []string{
				fmt.Sprintf("?%s=5", i.Key+filters.GT_SUFFIX),
			},
		},
	}
}
func (i Integer) ValidateFilters(queries map[string][]string) ([]schema.Filter, error) {
	valids := []schema.Filter{}

	// null check
	nullsOnly := false
	nonNullsOnly := false

	nullKey := i.Key + filters.ISNULL_SUFFIX
	nullStrings, exists := queries[nullKey]
	if exists {
		if len(nullStrings) != 1 {
			return filters.Exception(
				"Cannot null check attribute '%s' against more than one value.",
				i.Key,
			)
		}
		isNullString := strings.ToLower(nullStrings[0])
		if isNullString == "true" {
			nullsOnly = true
			valids = append(valids, IntegerNullFilter{
				BaseFilter: &schema.BaseFilter{
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
				BaseFilter: &schema.BaseFilter{
					QArgKey:    nullKey,
					QArgValues: []string{"false"},
				},
				null:   false,
				column: i.ColumnName,
			})
		} else {
			return filters.Exception(
				"Invalid null check value on attribute '%s'. Must be boolean.",
				i.Key,
			)
		}
	}

	exactKey := i.Key
	exactStrings, exists := queries[exactKey]
	if exists {
		if len(exactStrings) != 1 {
			return filters.Exception(
				"Cannot compare attribute '%s' against more than one value.",
				i.Key,
			)
		}

		if nullsOnly {
			return filters.Exception(
				"Cannot match attribute '%s' to an exact value and null.",
				i.Key,
			)
		}

		compareValue, err := strconv.Atoi(exactStrings[0])
		if err == nil {
			valids = append(valids, IntegerExactFilter{
				BaseFilter: &schema.BaseFilter{
					QArgKey:    exactKey,
					QArgValues: []string{strconv.Itoa(compareValue)},
				},
				value:  compareValue,
				column: i.ColumnName,
			})
		} else {
			return filters.Exception(
				"Invalid exact value on attribute '%s'. Must be integer.",
				i.Key,
			)
		}
	}

	filteringLesser := false
	var filteringLesserValue int

	lesserKey := i.Key + filters.LT_SUFFIX
	lesserStrings, exists := queries[lesserKey]
	if exists {
		if len(lesserStrings) != 1 {
			return filters.Exception(
				"Cannot compare attribute '%s' to be lesser than more than one value.",
				i.Key,
			)
		}

		if nullsOnly {
			return filters.Exception(
				"Cannot compare attribute '%s' to a value and null.",
				i.Key,
			)
		}

		lesserValue, err := strconv.Atoi(lesserStrings[0])
		if err == nil {
			filteringLesser = true
			filteringLesserValue = lesserValue

			valids = append(valids, IntegerLesserFilter{
				BaseFilter: &schema.BaseFilter{
					QArgKey:    lesserKey,
					QArgValues: []string{strconv.Itoa(lesserValue)},
				},
				value:  lesserValue,
				column: i.ColumnName,
				lesser: true,
			})
		} else {
			return filters.Exception(
				"Invalid lesser than comparison value on attribute '%s'. Must be integer.",
				i.Key,
			)
		}
	}

	greaterKey := i.Key + filters.GT_SUFFIX
	greaterStrings, exists := queries[greaterKey]
	if exists {
		if len(greaterStrings) != 1 {
			return filters.Exception(
				"Cannot compare attribute '%s' to be greater than more than one value.",
				i.Key,
			)
		}

		if nullsOnly {
			return filters.Exception(
				"Cannot compare attribute '%s' to a value and null.",
				i.Key,
			)
		}

		greaterValue, err := strconv.Atoi(greaterStrings[0])
		if err == nil {
			if filteringLesser && greaterValue > filteringLesserValue {
				return filters.Exception(
					"Cannot compare attribute '%s' to value greater than additional lesser than filter.",
					i.Key,
				)
			}

			valids = append(valids, IntegerLesserFilter{
				BaseFilter: &schema.BaseFilter{
					QArgKey:    greaterKey,
					QArgValues: []string{strconv.Itoa(greaterValue)},
				},
				value:  greaterValue,
				column: i.ColumnName,
				lesser: false,
			})
		} else {
			return filters.Exception(
				"Invalid greater than comparison value on attribute '%s'. Must be integer.",
				i.Key,
			)
		}
	}

	return valids, nil
}
