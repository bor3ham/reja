package attributes

import (
	"fmt"
	"github.com/bor3ham/reja/schema"
	"strings"
	"time"
)

type DateNullFilter struct {
	*BaseFilter
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
	*BaseFilter
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
	*BaseFilter
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
	*BaseFilter
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

func (d Date) AvailableFilters() []string {
	return []string{
		d.Key,
		d.Key + ISNULL_SUFFIX,
		d.Key + AFTER_SUFFIX,
		d.Key + BEFORE_SUFFIX,
	}
}
func (d Date) ValidateFilters(queries map[string][]string) ([]schema.Filter, error) {
	valids := []schema.Filter{}

	// null check
	nullsOnly := false
	nonNullsOnly := false

	nullKey := d.Key + ISNULL_SUFFIX
	nullStrings, exists := queries[nullKey]
	if exists {
		if len(nullStrings) != 1 {
			return filterException(
				"Cannot null check attribute '%s' against more than one value.",
				d.Key,
			)
		}
		isNullString := strings.ToLower(nullStrings[0])
		if isNullString == "true" {
			nullsOnly = true
			valids = append(valids, DateNullFilter{
				BaseFilter: &BaseFilter{
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
				BaseFilter: &BaseFilter{
					QArgKey:    nullKey,
					QArgValues: []string{"false"},
				},
				null:   false,
				column: d.ColumnName,
			})
		} else {
			return filterException(
				"Invalid null check value on attribute '%s'. Must be boolean.",
				d.Key,
			)
		}
	}

	exactKey := d.Key
	exactStrings, exists := queries[exactKey]
	if exists {
		if len(exactStrings) != 1 {
			return filterException(
				"Cannot compare attribute '%s' against more than one value.",
				d.Key,
			)
		}

		if nullsOnly {
			return filterException(
				"Cannot match attribute '%s' to an exact value and null.",
				d.Key,
			)
		}

		compareValue, err := time.Parse(DATE_LAYOUT, exactStrings[0])
		if err == nil {
			valids = append(valids, DateExactFilter{
				BaseFilter: &BaseFilter{
					QArgKey:    exactKey,
					QArgValues: []string{compareValue.Format(DATE_LAYOUT)},
				},
				value:  compareValue,
				column: d.ColumnName,
			})
		} else {
			return filterException(
				"Invalid exact value on attribute '%s'. Must be date in format %s.",
				d.Key,
				DATE_LAYOUT,
			)
		}
	}

	filteringAfter := false
	var filteringAfterValue time.Time

	afterKey := d.Key + AFTER_SUFFIX
	afterStrings, exists := queries[afterKey]
	if exists {
		if len(afterStrings) != 1 {
			return filterException(
				"Cannot compare attribute '%s' to be after more than one value.",
				d.Key,
			)
		}

		if nullsOnly {
			return filterException(
				"Cannot compare attribute '%s' to a value and null.",
				d.Key,
			)
		}

		afterValue, err := time.Parse(DATE_LAYOUT, afterStrings[0])
		if err == nil {
			filteringAfter = true
			filteringAfterValue = afterValue

			valids = append(valids, DateAfterFilter{
				BaseFilter: &BaseFilter{
					QArgKey:    exactKey,
					QArgValues: []string{afterValue.Format(DATE_LAYOUT)},
				},
				value:  afterValue,
				column: d.ColumnName,
			})
		} else {
			return filterException(
				"Invalid after comparison value on attribute '%s'. Must be date in format %s.",
				d.Key,
				DATE_LAYOUT,
			)
		}
	}

	beforeKey := d.Key + BEFORE_SUFFIX
	beforeStrings, exists := queries[beforeKey]
	if exists {
		if len(beforeStrings) != 1 {
			return filterException(
				"Cannot compare attribute '%s' to be before more than one value.",
				d.Key,
			)
		}

		if nullsOnly {
			return filterException(
				"Cannot compare attribute '%s' to a value and null.",
				d.Key,
			)
		}

		beforeValue, err := time.Parse(DATE_LAYOUT, beforeStrings[0])
		if err == nil {
			if filteringAfter && beforeValue.Before(filteringAfterValue) {
				return filterException(
					"Cannot compare attribute '%s' to value before additional after filter.",
					d.Key,
				)
			}

			valids = append(valids, DateBeforeFilter{
				BaseFilter: &BaseFilter{
					QArgKey:    exactKey,
					QArgValues: []string{beforeValue.Format(DATE_LAYOUT)},
				},
				value:  beforeValue,
				column: d.ColumnName,
			})
		} else {
			return filterException(
				"Invalid before comparison value on attribute '%s'. Must be date in format %s.",
				d.Key,
				DATE_LAYOUT,
			)
		}
	}

	return valids, nil
}
