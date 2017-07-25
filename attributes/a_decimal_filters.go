package attributes

import (
	"fmt"
	"github.com/bor3ham/reja/filters"
	"github.com/bor3ham/reja/schema"
	"github.com/shopspring/decimal"
	"strings"
)

type DecimalNullFilter struct {
	*schema.BaseFilter
	null   bool
	column string
}

func (f DecimalNullFilter) GetWhere(
	c schema.Context,
	modelTable string,
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

type DecimalExactFilter struct {
	*schema.BaseFilter
	value  decimal.Decimal
	column string
}

func (f DecimalExactFilter) GetWhere(
	c schema.Context,
	modelTable string,
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

type DecimalLesserFilter struct {
	*schema.BaseFilter
	value  decimal.Decimal
	column string
	lesser bool
}

func (f DecimalLesserFilter) GetWhere(
	c schema.Context,
	modelTable string,
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

func (d Decimal) AvailableFilters() []interface{} {
	return []interface{}{
		filters.FilterDescription{
			Key:         d.Key,
			Description: "Exact match on decimal value. Single value decimal.",
			Examples: []string{
				fmt.Sprintf("?%s=1.20", d.Key),
			},
		},
		filters.FilterDescription{
			Key:         d.Key + filters.ISNULL_SUFFIX,
			Description: "Whether decimal value exists. Single value boolean.",
			Examples: []string{
				fmt.Sprintf("?%s=true", d.Key+filters.ISNULL_SUFFIX),
				fmt.Sprintf("?%s=false", d.Key+filters.ISNULL_SUFFIX),
			},
		},
		filters.FilterDescription{
			Key:         d.Key + filters.LT_SUFFIX,
			Description: "Any value less than given decimal. Single value decimal.",
			Examples: []string{
				fmt.Sprintf("?%s=5.4", d.Key+filters.LT_SUFFIX),
			},
		},
		filters.FilterDescription{
			Key:         d.Key + filters.GT_SUFFIX,
			Description: "Any value greater than given decimal. Single value decimal.",
			Examples: []string{
				fmt.Sprintf("?%s=5.4", d.Key+filters.GT_SUFFIX),
			},
		},
	}
}
func (d Decimal) ValidateFilters(queries map[string][]string) ([]schema.Filter, error) {
	valids := []schema.Filter{}

	nullKey := d.Key + filters.ISNULL_SUFFIX
	nullStrings, exists := queries[nullKey]
	if exists {
		if len(nullStrings) != 1 {
			return filters.Exception(
				"Cannot null check attribute '%s' against more than one value.",
				d.Key,
			)
		}
		isNullString := strings.ToLower(nullStrings[0])
		if isNullString == "true" {
			valids = append(valids, DecimalNullFilter{
				BaseFilter: &schema.BaseFilter{
					QArgKey:    nullKey,
					QArgValues: []string{"true"},
				},
				null:   true,
				column: d.ColumnName,
			})
		} else if isNullString == "false" {
			valids = append(valids, DecimalNullFilter{
				BaseFilter: &schema.BaseFilter{
					QArgKey:    nullKey,
					QArgValues: []string{"false"},
				},
				null:   false,
				column: d.ColumnName,
			})
		} else {
			return filters.Exception(
				"Invalid null check value on attribute '%s'. Must be boolean.",
				d.Key,
			)
		}
	}

	exactKey := d.Key
	exactStrings, exists := queries[exactKey]
	if exists {
		if len(exactStrings) != 1 {
			return filters.Exception(
				"Cannot compare attribute '%s' against more than one value.",
				d.Key,
			)
		}

		compareValue, err := decimal.NewFromString(exactStrings[0])
		if err == nil {
			compareValue = compareValue.Truncate(d.DecimalPlaces)
			valids = append(valids, DecimalExactFilter{
				BaseFilter: &schema.BaseFilter{
					QArgKey:    exactKey,
					QArgValues: []string{compareValue.String()},
				},
				value:  compareValue,
				column: d.ColumnName,
			})
		} else {
			return filters.Exception(
				"Invalid exact value on attribute '%s'. Must be decimal.",
				d.Key,
			)
		}
	}

	lesserKey := d.Key + filters.LT_SUFFIX
	lesserStrings, exists := queries[lesserKey]
	if exists {
		if len(lesserStrings) != 1 {
			return filters.Exception(
				"Cannot compare attribute '%s' to be lesser than more than one value.",
				d.Key,
			)
		}

		lesserValue, err := decimal.NewFromString(lesserStrings[0])
		if err == nil {
			lesserValue = lesserValue.Truncate(d.DecimalPlaces)
			valids = append(valids, DecimalLesserFilter{
				BaseFilter: &schema.BaseFilter{
					QArgKey:    lesserKey,
					QArgValues: []string{lesserValue.String()},
				},
				value:  lesserValue,
				column: d.ColumnName,
				lesser: true,
			})
		} else {
			return filters.Exception(
				"Invalid lesser than comparison value on attribute '%s'. Must be decimal.",
				d.Key,
			)
		}
	}

	greaterKey := d.Key + filters.GT_SUFFIX
	greaterStrings, exists := queries[greaterKey]
	if exists {
		if len(greaterStrings) != 1 {
			return filters.Exception(
				"Cannot compare attribute '%s' to be greater than more than one value.",
				d.Key,
			)
		}

		greaterValue, err := decimal.NewFromString(greaterStrings[0])
		if err == nil {
			greaterValue = greaterValue.Truncate(d.DecimalPlaces)
			valids = append(valids, DecimalLesserFilter{
				BaseFilter: &schema.BaseFilter{
					QArgKey:    greaterKey,
					QArgValues: []string{greaterValue.String()},
				},
				value:  greaterValue,
				column: d.ColumnName,
				lesser: false,
			})
		} else {
			return filters.Exception(
				"Invalid greater than comparison value on attribute '%s'. Must be decimal.",
				d.Key,
			)
		}
	}

	return valids, nil
}
