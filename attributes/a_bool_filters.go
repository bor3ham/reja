package attributes

import (
	"fmt"
	"github.com/bor3ham/reja/filters"
	"github.com/bor3ham/reja/schema"
	"strings"
)

type BoolNullFilter struct {
	*schema.BaseFilter
	null   bool
	column string
}

func (f BoolNullFilter) GetWhere(
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

type BoolExactFilter struct {
	*schema.BaseFilter
	value  bool
	column string
}

func (f BoolExactFilter) GetWhere(
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

func (b Bool) AvailableFilters() []interface{} {
	return []interface{}{
		filters.FilterDescription{
			Key:         b.Key,
			Description: "Exact match on bool value. Single value boolean.",
			Examples: []string{
				fmt.Sprintf("?%s=true", b.Key),
				fmt.Sprintf("?%s=false", b.Key),
			},
		},
		filters.FilterDescription{
			Key:         b.Key + filters.ISNULL_SUFFIX,
			Description: "Whether bool value exists. Single value boolean.",
			Examples: []string{
				fmt.Sprintf("?%s=true", b.Key+filters.ISNULL_SUFFIX),
				fmt.Sprintf("?%s=false", b.Key+filters.ISNULL_SUFFIX),
			},
		},
	}
}
func (b Bool) ValidateFilters(queries map[string][]string) ([]schema.Filter, error) {
	valids := []schema.Filter{}

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
			valids = append(valids, BoolNullFilter{
				BaseFilter: &schema.BaseFilter{
					QArgKey:    nullKey,
					QArgValues: []string{"true"},
				},
				null:   true,
				column: b.ColumnName,
			})
		} else if isNullString == "false" {
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
