package relationships

import (
	"fmt"
	"strings"
	"github.com/bor3ham/reja/schema"
	"github.com/bor3ham/reja/filters"
)

func stringifyPointers(pointers []schema.InstancePointer) []string {
	strings := []string{}
	for _, pointer := range pointers {
		strings = append(strings, fmt.Sprintf("%s:%s", pointer.Type, *pointer.ID))
	}
	return strings
}

type GenericForeignKeyNullFilter struct {
	*schema.BaseFilter
	null   bool
	idColumn string
}
func (f GenericForeignKeyNullFilter) GetWhereQueries(nextArg int) []string {
	if f.null {
		return []string{
			fmt.Sprintf("%s is null", f.idColumn),
		}
	} else {
		return []string{
			fmt.Sprintf("%s is not null", f.idColumn),
		}
	}
}
func (f GenericForeignKeyNullFilter) GetWhereArgs() []interface{} {
	return []interface{}{}
}

type GenericForeignKeyTypeFilter struct {
	*schema.BaseFilter
	values  []string
	typeColumn string
}
func (f GenericForeignKeyTypeFilter) GetWhereQueries(nextArg int) []string {
	query := ""
	if len(f.values) > 1 {
		query += "("
	}
	for valueIndex, _ := range f.values {
		if valueIndex > 0 {
			query += " or "
		}
		query += fmt.Sprintf(
			"%s = $%d",
			f.typeColumn,
			nextArg,
		)
		nextArg += 1
	}
	if len(f.values) > 1 {
		query += ")"
	}
	return []string{query}
}
func (f GenericForeignKeyTypeFilter) GetWhereArgs() []interface{} {
	args := []interface{}{}
	for _, value := range f.values {
		args = append(args, value)
	}
	return args
}

type GenericForeignKeyIDFilter struct {
	*schema.BaseFilter
	values  []string
	idColumn string
}
func (f GenericForeignKeyIDFilter) GetWhereQueries(nextArg int) []string {
	query := ""
	if len(f.values) > 1 {
		query += "("
	}
	for valueIndex, _ := range f.values {
		if valueIndex > 0 {
			query += " or "
		}
		query += fmt.Sprintf(
			"%s = $%d",
			f.idColumn,
			nextArg,
		)
		nextArg += 1
	}
	if len(f.values) > 1 {
		query += ")"
	}
	return []string{query}
}
func (f GenericForeignKeyIDFilter) GetWhereArgs() []interface{} {
	args := []interface{}{}
	for _, value := range f.values {
		args = append(args, value)
	}
	return args
}

type GenericForeignKeyExactFilter struct {
	*schema.BaseFilter
	values  []schema.InstancePointer
	typeColumn string
	idColumn string
}
func (f GenericForeignKeyExactFilter) GetWhereQueries(nextArg int) []string {
	query := ""
	if len(f.values) > 0 {
		if len(f.values) > 1 {
			query = "("
		}
		for valueIndex, _ := range f.values {
			if valueIndex > 0 {
				query += " or "
			}
			query += fmt.Sprintf(
				"(%s = $%d and %s = $%d)",
				f.typeColumn,
				nextArg,
				f.idColumn,
				nextArg + 1,
			)
			nextArg += 2
		}
		if len(f.values) > 1 {
			query += ")"
		}
	}
	return []string{query}
}
func (f GenericForeignKeyExactFilter) GetWhereArgs() []interface{} {
	args := []interface{}{}
	for _, value := range f.values {
		args = append(args, value.Type)
		args = append(args, *value.ID)
	}
	return args
}

func (gfk GenericForeignKey) ValidateFilters(queries map[string][]string) ([]schema.Filter, error) {
	valids := []schema.Filter{}

	// null check
	nullsOnly := false
	nonNullsOnly := false

	nullKey := gfk.Key + filters.ISNULL_SUFFIX
	nullStrings, exists := queries[nullKey]
	if exists {
		if len(nullStrings) != 1 {
			return filters.Exception(
				"Cannot null check attribute '%s' against more than one value.",
				gfk.Key,
			)
		}
		isNullString := strings.ToLower(nullStrings[0])
		if isNullString == "true" {
			nullsOnly = true
			valids = append(valids, GenericForeignKeyNullFilter{
				BaseFilter: &schema.BaseFilter{
					QArgKey:    nullKey,
					QArgValues: []string{"true"},
				},
				null:   true,
				idColumn: gfk.IDColumnName,
			})
		} else if isNullString == "false" {
			nonNullsOnly = true
			_ = nonNullsOnly
			valids = append(valids, GenericForeignKeyNullFilter{
				BaseFilter: &schema.BaseFilter{
					QArgKey:    nullKey,
					QArgValues: []string{"false"},
				},
				null:   false,
				idColumn: gfk.IDColumnName,
			})
		} else {
			return filters.Exception(
				"Invalid null check value on attribute '%s'. Must be boolean.",
				gfk.Key,
			)
		}
	}

	typeKey := gfk.Key + filters.TYPE_SUFFIX
	typeStrings, exists := queries[typeKey]
	if exists {
		if nullsOnly {
			return filters.Exception(
				"Cannot match attribute '%s' to a type value and null.",
				gfk.Key,
			)
		}

		compareValues := []string{}
		for _, value := range typeStrings {
			cleanValue := strings.TrimSpace(value)
			if len(gfk.ValidTypes) > 0 {
				valid := false
				for _, relationType := range gfk.ValidTypes {
					if cleanValue == relationType {
						valid = true
					}
				}
				if !valid {
					return filters.Exception(
						"Cannot match attribute '%s' type to invalid value.",
						gfk.Key,
					)
				}
			}
			compareValues = append(compareValues, cleanValue)
		}

		valids = append(valids, GenericForeignKeyTypeFilter{
			BaseFilter: &schema.BaseFilter{
				QArgKey:    typeKey,
				QArgValues: compareValues,
			},
			values:  compareValues,
			typeColumn: gfk.TypeColumnName,
		})
	}

	idKey := gfk.Key + filters.ID_SUFFIX
	idStrings, exists := queries[idKey]
	if exists {
		if nullsOnly {
			return filters.Exception(
				"Cannot match attribute '%s' to an ID value and null.",
				gfk.Key,
			)
		}

		compareValues := []string{}
		for _, value := range idStrings {
			compareValues = append(compareValues, strings.ToLower(strings.TrimSpace(value)))
		}

		valids = append(valids, GenericForeignKeyIDFilter{
			BaseFilter: &schema.BaseFilter{
				QArgKey:    idKey,
				QArgValues: compareValues,
			},
			values:  compareValues,
			idColumn: gfk.IDColumnName,
		})
	}

	exactKey := gfk.Key
	exactStrings, exists := queries[exactKey]
	if exists {
		if nullsOnly {
			return filters.Exception(
				"Cannot match attribute '%s' to an exact value and null.",
				gfk.Key,
			)
		}

		comparePointers := []schema.InstancePointer{}
		for _, value := range exactStrings {
			cleanValue := strings.TrimSpace(value)
			splitValue := strings.Split(cleanValue, ":")
			if len(splitValue) != 2 {
				return filters.Exception(
					"Invalid exact match on attribute '%s'. Must be instance pointer in format Type:ID.",
					gfk.Key,
				)
			}
			stringType := splitValue[0]
			if len(gfk.ValidTypes) > 0 {
				valid := false
				for _, relationType := range gfk.ValidTypes {
					if stringType == relationType {
						valid = true
					}
				}
				if !valid {
					return filters.Exception(
						"Invalid exact match on attribute '%s'. Must be valid type choice for relationship.",
						gfk.Key,
					)
				}
			}
			stringID := splitValue[1]
			comparePointers = append(comparePointers, schema.InstancePointer{
				Type: stringType,
				ID: &stringID,
			})
		}

		valids = append(valids, GenericForeignKeyExactFilter{
			BaseFilter: &schema.BaseFilter{
				QArgKey:    exactKey,
				QArgValues: stringifyPointers(comparePointers),
			},
			values:  comparePointers,
			typeColumn: gfk.TypeColumnName,
			idColumn: gfk.IDColumnName,
		})
	}

	return valids, nil
}
