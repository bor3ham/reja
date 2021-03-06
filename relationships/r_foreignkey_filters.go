package relationships

import (
	"fmt"
	"github.com/bor3ham/reja/filters"
	"github.com/bor3ham/reja/schema"
	"strings"
)

type ForeignKeyNullFilter struct {
	*schema.BaseFilter
	null   bool
	column string
}

func (f ForeignKeyNullFilter) GetWhere(
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

type ForeignKeyExactFilter struct {
	*schema.BaseFilter
	values []string
	column string
}

func (f ForeignKeyExactFilter) GetWhere(
	c schema.Context,
	modelTable string,
	idColumn string,
	nextArg int,
) (
	[]string,
	[]interface{},
) {
	spots := []string{}
	args := []interface{}{}
	for _, value := range f.values {
		spots = append(spots, fmt.Sprintf("$%d", nextArg))
		args = append(args, value)
		nextArg += 1
	}
	return []string{
		fmt.Sprintf("%s in (%s)", f.column, strings.Join(spots, ", ")),
	}, args
}

func (fk ForeignKey) AvailableFilters() []interface{} {
	return []interface{}{
		filters.FilterDescription{
			Key:         fk.Key,
			Description: "Related item to filter for. One or more IDs.",
			Examples: []string{
				fmt.Sprintf("?%s=1", fk.Key),
				fmt.Sprintf("?%s=1&%s=2", fk.Key, fk.Key),
			},
		},
		filters.FilterDescription{
			Key:         fk.Key + filters.ISNULL_SUFFIX,
			Description: "Whether related item exists. Single value boolean.",
			Examples: []string{
				fmt.Sprintf("?%s=true", fk.Key+filters.ISNULL_SUFFIX),
				fmt.Sprintf("?%s=false", fk.Key+filters.ISNULL_SUFFIX),
			},
		},
	}
}
func (fk ForeignKey) ValidateFilters(queries map[string][]string) ([]schema.Filter, error) {
	valids := []schema.Filter{}

	// null check
	nullsOnly := false
	nonNullsOnly := false

	nullKey := fk.Key + filters.ISNULL_SUFFIX
	nullStrings, exists := queries[nullKey]
	if exists {
		if len(nullStrings) != 1 {
			return filters.Exception(
				"Cannot null check attribute '%s' against more than one value.",
				fk.Key,
			)
		}
		isNullString := strings.ToLower(nullStrings[0])
		if isNullString == "true" {
			nullsOnly = true
			valids = append(valids, ForeignKeyNullFilter{
				BaseFilter: &schema.BaseFilter{
					QArgKey:    nullKey,
					QArgValues: []string{"true"},
				},
				null:   true,
				column: fk.ColumnName,
			})
		} else if isNullString == "false" {
			nonNullsOnly = true
			_ = nonNullsOnly
			valids = append(valids, ForeignKeyNullFilter{
				BaseFilter: &schema.BaseFilter{
					QArgKey:    nullKey,
					QArgValues: []string{"false"},
				},
				null:   false,
				column: fk.ColumnName,
			})
		} else {
			return filters.Exception(
				"Invalid null check value on attribute '%s'. Must be boolean.",
				fk.Key,
			)
		}
	}

	exactKey := fk.Key
	exactStrings, exists := queries[exactKey]
	if exists {
		if nullsOnly {
			return filters.Exception(
				"Cannot match attribute '%s' to an exact value and null.",
				fk.Key,
			)
		}

		compareValues := []string{}
		for _, value := range exactStrings {
			compareValues = append(compareValues, strings.TrimSpace(value))
		}

		valids = append(valids, ForeignKeyExactFilter{
			BaseFilter: &schema.BaseFilter{
				QArgKey:    exactKey,
				QArgValues: compareValues,
			},
			values: compareValues,
			column: fk.ColumnName,
		})
	}

	return valids, nil
}
