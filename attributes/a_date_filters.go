package attributes

import (
	"fmt"
	"github.com/bor3ham/reja/filters"
	"github.com/bor3ham/reja/schema"
	"strings"
	"time"
)

type DateNullFilter struct {
	*schema.BaseFilter
	null   bool
	column string
}

func (f DateNullFilter) GetWhereQueries(nextArg int) []string {
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
func (f DateNullFilter) GetWhereArgs() []interface{} {
	return []interface{}{}
}

type DateExactFilter struct {
	*schema.BaseFilter
	value  time.Time
	column string
}

func (f DateExactFilter) GetWhereQueries(nextArg int) []string {
	return []string{
		fmt.Sprintf("%s = $%d", f.column, nextArg),
	}
}
func (f DateExactFilter) GetWhereArgs() []interface{} {
	return []interface{}{
		f.value,
	}
}

type DateAfterFilter struct {
	*schema.BaseFilter
	value  time.Time
	column string
}

func (f DateAfterFilter) GetWhereQueries(nextArg int) []string {
	return []string{
		fmt.Sprintf("%s > $%d", f.column, nextArg),
	}
}
func (f DateAfterFilter) GetWhereArgs() []interface{} {
	return []interface{}{
		f.value,
	}
}

type DateBeforeFilter struct {
	*schema.BaseFilter
	value  time.Time
	column string
}

func (f DateBeforeFilter) GetWhereQueries(nextArg int) []string {
	return []string{
		fmt.Sprintf("%s < $%d", f.column, nextArg),
	}
}
func (f DateBeforeFilter) GetWhereArgs() []interface{} {
	return []interface{}{
		f.value,
	}
}

func (d Date) AvailableFilters() []interface{} {
	return []interface{}{
		filters.FilterDescription{
			Key: d.Key,
			Description: fmt.Sprintf(
				"Exact match on date value. Single value date in format '%s'.",
				DATE_LAYOUT,
			),
			Examples: []string{
				fmt.Sprintf("?%s=%s", d.Key, time.Now().Format(DATE_LAYOUT)),
			},
		},
		filters.FilterDescription{
			Key:         d.Key + filters.ISNULL_SUFFIX,
			Description: "Whether date value exists. Single value boolean.",
			Examples: []string{
				fmt.Sprintf("?%s=true", d.Key+filters.ISNULL_SUFFIX),
				fmt.Sprintf("?%s=false", d.Key+filters.ISNULL_SUFFIX),
			},
		},
		filters.FilterDescription{
			Key: d.Key + filters.AFTER_SUFFIX,
			Description: fmt.Sprintf(
				"Any date values after given date. Single value date in format '%s'.",
				DATE_LAYOUT,
			),
			Examples: []string{
				fmt.Sprintf("?%s=%s", d.Key+filters.AFTER_SUFFIX, time.Now().Format(DATE_LAYOUT)),
			},
		},
		filters.FilterDescription{
			Key: d.Key + filters.BEFORE_SUFFIX,
			Description: fmt.Sprintf(
				"Any date values before given date. Single value date in format '%s'.",
				DATE_LAYOUT,
			),
			Examples: []string{
				fmt.Sprintf("?%s=%s", d.Key+filters.BEFORE_SUFFIX, time.Now().Format(DATE_LAYOUT)),
			},
		},
	}
}
func (d Date) ValidateFilters(queries map[string][]string) ([]schema.Filter, error) {
	valids := []schema.Filter{}

	// null check
	nullsOnly := false
	nonNullsOnly := false

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
			nullsOnly = true
			valids = append(valids, DateNullFilter{
				BaseFilter: &schema.BaseFilter{
					QArgKey:    nullKey,
					QArgValues: []string{"true"},
				},
				null:   true,
				column: d.ColumnName,
			})
		} else if isNullString == "false" {
			nonNullsOnly = true
			_ = nonNullsOnly
			valids = append(valids, DateNullFilter{
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

		if nullsOnly {
			return filters.Exception(
				"Cannot match attribute '%s' to an exact value and null.",
				d.Key,
			)
		}

		compareValue, err := time.Parse(DATE_LAYOUT, exactStrings[0])
		if err == nil {
			valids = append(valids, DateExactFilter{
				BaseFilter: &schema.BaseFilter{
					QArgKey:    exactKey,
					QArgValues: []string{compareValue.Format(DATE_LAYOUT)},
				},
				value:  compareValue,
				column: d.ColumnName,
			})
		} else {
			return filters.Exception(
				"Invalid exact value on attribute '%s'. Must be date in format %s.",
				d.Key,
				DATE_LAYOUT,
			)
		}
	}

	filteringAfter := false
	var filteringAfterValue time.Time

	afterKey := d.Key + filters.AFTER_SUFFIX
	afterStrings, exists := queries[afterKey]
	if exists {
		if len(afterStrings) != 1 {
			return filters.Exception(
				"Cannot compare attribute '%s' to be after more than one value.",
				d.Key,
			)
		}

		if nullsOnly {
			return filters.Exception(
				"Cannot compare attribute '%s' to a value and null.",
				d.Key,
			)
		}

		afterValue, err := time.Parse(DATE_LAYOUT, afterStrings[0])
		if err == nil {
			filteringAfter = true
			filteringAfterValue = afterValue

			valids = append(valids, DateAfterFilter{
				BaseFilter: &schema.BaseFilter{
					QArgKey:    afterKey,
					QArgValues: []string{afterValue.Format(DATE_LAYOUT)},
				},
				value:  afterValue,
				column: d.ColumnName,
			})
		} else {
			return filters.Exception(
				"Invalid after comparison value on attribute '%s'. Must be date in format %s.",
				d.Key,
				DATE_LAYOUT,
			)
		}
	}

	beforeKey := d.Key + filters.BEFORE_SUFFIX
	beforeStrings, exists := queries[beforeKey]
	if exists {
		if len(beforeStrings) != 1 {
			return filters.Exception(
				"Cannot compare attribute '%s' to be before more than one value.",
				d.Key,
			)
		}

		if nullsOnly {
			return filters.Exception(
				"Cannot compare attribute '%s' to a value and null.",
				d.Key,
			)
		}

		beforeValue, err := time.Parse(DATE_LAYOUT, beforeStrings[0])
		if err == nil {
			if filteringAfter && beforeValue.Before(filteringAfterValue) {
				return filters.Exception(
					"Cannot compare attribute '%s' to value before additional after filter.",
					d.Key,
				)
			}

			valids = append(valids, DateBeforeFilter{
				BaseFilter: &schema.BaseFilter{
					QArgKey:    beforeKey,
					QArgValues: []string{beforeValue.Format(DATE_LAYOUT)},
				},
				value:  beforeValue,
				column: d.ColumnName,
			})
		} else {
			return filters.Exception(
				"Invalid before comparison value on attribute '%s'. Must be date in format %s.",
				d.Key,
				DATE_LAYOUT,
			)
		}
	}

	return valids, nil
}
