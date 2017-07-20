package attributes

import (
	"fmt"
	"github.com/bor3ham/reja/schema"
	"strings"
	"time"
)

type DatetimeNullFilter struct {
	*schema.BaseFilter
	null   bool
	column string
}

func (f DatetimeNullFilter) GetWhereQueries(nextArg int) []string {
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
func (f DatetimeNullFilter) GetWhereArgs() []interface{} {
	return []interface{}{}
}

type DatetimeExactFilter struct {
	*schema.BaseFilter
	value  time.Time
	column string
}

func (f DatetimeExactFilter) GetWhereQueries(nextArg int) []string {
	return []string{
		fmt.Sprintf("date_trunc('second', %s) = $%d", f.column, nextArg),
	}
}
func (f DatetimeExactFilter) GetWhereArgs() []interface{} {
	return []interface{}{
		f.value,
	}
}

type DatetimeAfterFilter struct {
	*schema.BaseFilter
	value  time.Time
	column string
}

func (f DatetimeAfterFilter) GetWhereQueries(nextArg int) []string {
	return []string{
		fmt.Sprintf("date_trunc('second', %s) > $%d", f.column, nextArg),
	}
}
func (f DatetimeAfterFilter) GetWhereArgs() []interface{} {
	return []interface{}{
		f.value,
	}
}

type DatetimeBeforeFilter struct {
	*schema.BaseFilter
	value  time.Time
	column string
}

func (f DatetimeBeforeFilter) GetWhereQueries(nextArg int) []string {
	return []string{
		fmt.Sprintf("date_trunc('second', %s) < $%d", f.column, nextArg),
	}
}
func (f DatetimeBeforeFilter) GetWhereArgs() []interface{} {
	return []interface{}{
		f.value,
	}
}

func (dt Datetime) AvailableFilters() []string {
	return []string{
		dt.Key,
		dt.Key + ISNULL_SUFFIX,
		dt.Key + AFTER_SUFFIX,
		dt.Key + BEFORE_SUFFIX,
	}
}
func (dt Datetime) ValidateFilters(queries map[string][]string) ([]schema.Filter, error) {
	valids := []schema.Filter{}

	// null check
	nullsOnly := false
	nonNullsOnly := false

	nullKey := dt.Key + ISNULL_SUFFIX
	nullStrings, exists := queries[nullKey]
	if exists {
		if len(nullStrings) != 1 {
			return filterException(
				"Cannot null check attribute '%s' against more than one value.",
				dt.Key,
			)
		}
		isNullString := strings.ToLower(nullStrings[0])
		if isNullString == "true" {
			nullsOnly = true
			valids = append(valids, DatetimeNullFilter{
				BaseFilter: &schema.BaseFilter{
					QArgKey:    nullKey,
					QArgValues: []string{"true"},
				},
				null:   true,
				column: dt.ColumnName,
			})
		} else if isNullString == "false" {
			nonNullsOnly = true
			_ = nonNullsOnly
			valids = append(valids, DatetimeNullFilter{
				BaseFilter: &schema.BaseFilter{
					QArgKey:    nullKey,
					QArgValues: []string{"false"},
				},
				null:   false,
				column: dt.ColumnName,
			})
		} else {
			return filterException(
				"Invalid null check value on attribute '%s'. Must be boolean.",
				dt.Key,
			)
		}
	}

	exactKey := dt.Key
	exactStrings, exists := queries[exactKey]
	if exists {
		if len(exactStrings) != 1 {
			return filterException(
				"Cannot compare attribute '%s' against more than one value.",
				dt.Key,
			)
		}

		if nullsOnly {
			return filterException(
				"Cannot match attribute '%s' to an exact value and null.",
				dt.Key,
			)
		}

		compareValue, err := time.Parse(time.RFC3339, exactStrings[0])
		if err == nil {
			valids = append(valids, DatetimeExactFilter{
				BaseFilter: &schema.BaseFilter{
					QArgKey:    exactKey,
					QArgValues: []string{compareValue.Format(time.RFC3339)},
				},
				value:  compareValue,
				column: dt.ColumnName,
			})
		} else {
			return filterException(
				"Invalid exact value on attribute '%s'. Must be datetime in RFC3339 format.",
				dt.Key,
			)
		}
	}

	filteringAfter := false
	var filteringAfterValue time.Time

	afterKey := dt.Key + AFTER_SUFFIX
	afterStrings, exists := queries[afterKey]
	if exists {
		if len(afterStrings) != 1 {
			return filterException(
				"Cannot compare attribute '%s' to be after more than one value.",
				dt.Key,
			)
		}

		if nullsOnly {
			return filterException(
				"Cannot compare attribute '%s' to a value and null.",
				dt.Key,
			)
		}

		afterValue, err := time.Parse(time.RFC3339, afterStrings[0])
		if err == nil {
			filteringAfter = true
			filteringAfterValue = afterValue

			valids = append(valids, DatetimeAfterFilter{
				BaseFilter: &schema.BaseFilter{
					QArgKey:    afterKey,
					QArgValues: []string{afterValue.Format(time.RFC3339)},
				},
				value:  afterValue,
				column: dt.ColumnName,
			})
		} else {
			return filterException(
				"Invalid after comparison value on attribute '%s'. Must be datetime in RFC3339 format.",
				dt.Key,
			)
		}
	}

	beforeKey := dt.Key + BEFORE_SUFFIX
	beforeStrings, exists := queries[beforeKey]
	if exists {
		if len(beforeStrings) != 1 {
			return filterException(
				"Cannot compare attribute '%s' to be before more than one value.",
				dt.Key,
			)
		}

		if nullsOnly {
			return filterException(
				"Cannot compare attribute '%s' to a value and null.",
				dt.Key,
			)
		}

		beforeValue, err := time.Parse(time.RFC3339, beforeStrings[0])
		if err == nil {
			if filteringAfter && beforeValue.Before(filteringAfterValue) {
				return filterException(
					"Cannot compare attribute '%s' to value before additional after filter.",
					dt.Key,
				)
			}

			valids = append(valids, DatetimeBeforeFilter{
				BaseFilter: &schema.BaseFilter{
					QArgKey:    beforeKey,
					QArgValues: []string{beforeValue.Format(time.RFC3339)},
				},
				value:  beforeValue,
				column: dt.ColumnName,
			})
		} else {
			return filterException(
				"Invalid before comparison value on attribute '%s'. Must be datetime in RFC3339 format.",
				dt.Key,
			)
		}
	}

	return valids, nil
}
