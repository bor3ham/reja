package attributes

import (
	"fmt"
	"github.com/bor3ham/reja/filters"
	"github.com/bor3ham/reja/schema"
	"net/url"
	"strings"
	"time"
)

type DatetimeNullFilter struct {
	*schema.BaseFilter
	null   bool
	column string
}

func (f DatetimeNullFilter) GetWhere(
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

type DatetimeExactFilter struct {
	*schema.BaseFilter
	value  time.Time
	column string
}

func (f DatetimeExactFilter) GetWhere(
	c schema.Context,
	idColumn string,
	nextArg int,
) (
	[]string,
	[]interface{},
) {
	return []string{
		fmt.Sprintf("date_trunc('second', %s) = $%d", f.column, nextArg),
	}, []interface{}{
		f.value,
	}
}

type DatetimeAfterFilter struct {
	*schema.BaseFilter
	value  time.Time
	column string
	after bool
}

func (f DatetimeAfterFilter) GetWhere(
	c schema.Context,
	idColumn string,
	nextArg int,
) (
	[]string,
	[]interface{},
) {
	operator := ">"
	if !f.after {
		operator = "<"
	}
	return []string{
		fmt.Sprintf("date_trunc('second', %s) %s $%d", f.column, operator, nextArg),
	}, []interface{}{
		f.value,
	}
}

func (dt Datetime) AvailableFilters() []interface{} {
	return []interface{}{
		filters.FilterDescription{
			Key:         dt.Key,
			Description: "Exact match on datetime value. Single value datetime in RFC3339 format.",
			Examples: []string{
				fmt.Sprintf(
					"?%s=%s",
					dt.Key,
					url.QueryEscape(time.Now().Format(time.RFC3339)),
				),
			},
		},
		filters.FilterDescription{
			Key:         dt.Key + filters.ISNULL_SUFFIX,
			Description: "Whether datetime value exists. Single value boolean.",
			Examples: []string{
				fmt.Sprintf("?%s=true", dt.Key+filters.ISNULL_SUFFIX),
				fmt.Sprintf("?%s=false", dt.Key+filters.ISNULL_SUFFIX),
			},
		},
		filters.FilterDescription{
			Key:         dt.Key + filters.AFTER_SUFFIX,
			Description: "Any datetime values after given time. Single value datetime in RFC3339 format.",
			Examples: []string{
				fmt.Sprintf(
					"?%s=%s",
					dt.Key+filters.AFTER_SUFFIX,
					url.QueryEscape(time.Now().Format(time.RFC3339)),
				),
			},
		},
		filters.FilterDescription{
			Key:         dt.Key + filters.BEFORE_SUFFIX,
			Description: "Any date values before given date. Single value date in RFC3339 format.",
			Examples: []string{
				fmt.Sprintf(
					"?%s=%s",
					dt.Key+filters.BEFORE_SUFFIX,
					url.QueryEscape(time.Now().Format(time.RFC3339)),
				),
			},
		},
	}
}
func (dt Datetime) ValidateFilters(queries map[string][]string) ([]schema.Filter, error) {
	valids := []schema.Filter{}

	// null check
	nullsOnly := false
	nonNullsOnly := false

	nullKey := dt.Key + filters.ISNULL_SUFFIX
	nullStrings, exists := queries[nullKey]
	if exists {
		if len(nullStrings) != 1 {
			return filters.Exception(
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
			return filters.Exception(
				"Invalid null check value on attribute '%s'. Must be boolean.",
				dt.Key,
			)
		}
	}

	exactKey := dt.Key
	exactStrings, exists := queries[exactKey]
	if exists {
		if len(exactStrings) != 1 {
			return filters.Exception(
				"Cannot compare attribute '%s' against more than one value.",
				dt.Key,
			)
		}

		if nullsOnly {
			return filters.Exception(
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
			return filters.Exception(
				"Invalid exact value on attribute '%s'. Must be datetime in RFC3339 format.",
				dt.Key,
			)
		}
	}

	filteringAfter := false
	var filteringAfterValue time.Time

	afterKey := dt.Key + filters.AFTER_SUFFIX
	afterStrings, exists := queries[afterKey]
	if exists {
		if len(afterStrings) != 1 {
			return filters.Exception(
				"Cannot compare attribute '%s' to be after more than one value.",
				dt.Key,
			)
		}

		if nullsOnly {
			return filters.Exception(
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
				after: true,
			})
		} else {
			return filters.Exception(
				"Invalid after comparison value on attribute '%s'. Must be datetime in RFC3339 format.",
				dt.Key,
			)
		}
	}

	beforeKey := dt.Key + filters.BEFORE_SUFFIX
	beforeStrings, exists := queries[beforeKey]
	if exists {
		if len(beforeStrings) != 1 {
			return filters.Exception(
				"Cannot compare attribute '%s' to be before more than one value.",
				dt.Key,
			)
		}

		if nullsOnly {
			return filters.Exception(
				"Cannot compare attribute '%s' to a value and null.",
				dt.Key,
			)
		}

		beforeValue, err := time.Parse(time.RFC3339, beforeStrings[0])
		if err == nil {
			if filteringAfter && beforeValue.Before(filteringAfterValue) {
				return filters.Exception(
					"Cannot compare attribute '%s' to value before additional after filter.",
					dt.Key,
				)
			}

			valids = append(valids, DatetimeAfterFilter{
				BaseFilter: &schema.BaseFilter{
					QArgKey:    beforeKey,
					QArgValues: []string{beforeValue.Format(time.RFC3339)},
				},
				value:  beforeValue,
				column: dt.ColumnName,
				after: false,
			})
		} else {
			return filters.Exception(
				"Invalid before comparison value on attribute '%s'. Must be datetime in RFC3339 format.",
				dt.Key,
			)
		}
	}

	return valids, nil
}
