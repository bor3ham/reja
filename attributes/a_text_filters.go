package attributes

import (
	"fmt"
	"github.com/bor3ham/reja/filters"
	"github.com/bor3ham/reja/schema"
	"strconv"
	"strings"
)

type TextNullFilter struct {
	*schema.BaseFilter
	null   bool
	column string
}

func (f TextNullFilter) GetWhere(
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

type TextExactFilter struct {
	*schema.BaseFilter
	matching string
	column   string
}

func (f TextExactFilter) GetWhere(
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
			f.matching,
		}
}

type TextContainsFilter struct {
	*schema.BaseFilter
	contains []string
	column   string
}

func (f TextContainsFilter) GetWhere(
	c schema.Context,
	modelTable string,
	idColumn string,
	nextArg int,
) (
	[]string,
	[]interface{},
) {
	args := []interface{}{}
	where := "("
	for matchIndex, match := range f.contains {
		if matchIndex > 0 {
			where += " or "
		}
		where += fmt.Sprintf(
			`%s ilike '%%' || $%d || '%%'`,
			f.column,
			nextArg,
		)
		args = append(args, strings.Replace(match, "%%", "\\%%", -1))
		nextArg += 1
	}
	where += ")"
	return []string{
		where,
	}, args
}

type TextLengthExactFilter struct {
	*schema.BaseFilter
	length int
	column string
}

func (f TextLengthExactFilter) GetWhere(
	c schema.Context,
	modelTable string,
	idColumn string,
	nextArg int,
) (
	[]string,
	[]interface{},
) {
	return []string{
			fmt.Sprintf("char_length(%s) = $%d", f.column, nextArg),
		}, []interface{}{
			f.length,
		}
}

type TextLengthLesserFilter struct {
	*schema.BaseFilter
	length int
	column string
	lesser bool
}

func (f TextLengthLesserFilter) GetWhere(
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
			fmt.Sprintf("char_length(%s) %s $%d", f.column, operator, nextArg),
		}, []interface{}{
			f.length,
		}
}

type TextLengthGreaterFilter struct {
	*schema.BaseFilter
	length int
	column string
}

func (f TextLengthGreaterFilter) GetWhere(
	c schema.Context,
	modelTable string,
	idColumn string,
	nextArg int,
) (
	[]string,
	[]interface{},
) {
	return []string{
			fmt.Sprintf("char_length(%s) > $%d", f.column, nextArg),
		}, []interface{}{
			f.length,
		}
}

func (t Text) AvailableFilters() []interface{} {
	return []interface{}{
		filters.FilterDescription{
			Key:         t.Key,
			Description: "Exact match on text value. Single value case sensitive freeform text.",
			Examples: []string{
				fmt.Sprintf("?%s=Foo", t.Key),
			},
		},
		filters.FilterDescription{
			Key:         t.Key + filters.ISNULL_SUFFIX,
			Description: "Whether text value exists. Single value boolean.",
			Examples: []string{
				fmt.Sprintf("?%s=true", t.Key+filters.ISNULL_SUFFIX),
				fmt.Sprintf("?%s=false", t.Key+filters.ISNULL_SUFFIX),
			},
		},
		filters.FilterDescription{
			Key:         t.Key + filters.LENGTH_SUFFIX,
			Description: "Exact match on text length. Single value integer.",
			Examples: []string{
				fmt.Sprintf("?%s=5", t.Key+filters.LENGTH_SUFFIX),
			},
		},
		filters.FilterDescription{
			Key:         t.Key + filters.CONTAINS_SUFFIX,
			Description: "Contains match on text value. Single value case insensitive freeform text.",
			Examples: []string{
				fmt.Sprintf("?%s=foo", t.Key+filters.CONTAINS_SUFFIX),
			},
		},
		filters.FilterDescription{
			Key:         t.Key + filters.LENGTH_SUFFIX + filters.LT_SUFFIX,
			Description: "Any text value with a length less than given integer. Single value integer.",
			Examples: []string{
				fmt.Sprintf("?%s=5", t.Key+filters.LENGTH_SUFFIX+filters.LT_SUFFIX),
			},
		},
		filters.FilterDescription{
			Key:         t.Key + filters.LENGTH_SUFFIX + filters.GT_SUFFIX,
			Description: "Any text value with a greater less than given integer. Single value integer.",
			Examples: []string{
				fmt.Sprintf("?%s=5", t.Key+filters.LENGTH_SUFFIX+filters.GT_SUFFIX),
			},
		},
	}
}
func (t Text) ValidateFilters(queries map[string][]string) ([]schema.Filter, error) {
	valids := []schema.Filter{}

	nullKey := t.Key + filters.ISNULL_SUFFIX
	nullStrings, exists := queries[nullKey]
	if exists {
		if len(nullStrings) != 1 {
			return filters.Exception(
				"Cannot null check attribute '%s' against more than one value.",
				t.Key,
			)
		}
		isNullString := strings.ToLower(nullStrings[0])
		if isNullString == "true" {
			valids = append(valids, TextNullFilter{
				BaseFilter: &schema.BaseFilter{
					QArgKey:    nullKey,
					QArgValues: []string{"true"},
				},
				null:   true,
				column: t.ColumnName,
			})
		} else if isNullString == "false" {
			valids = append(valids, TextNullFilter{
				BaseFilter: &schema.BaseFilter{
					QArgKey:    nullKey,
					QArgValues: []string{"false"},
				},
				null:   false,
				column: t.ColumnName,
			})
		} else {
			return filters.Exception(
				"Invalid null check value on attribute '%s'. Must be boolean.",
				t.Key,
			)
		}
	}

	exactKey := t.Key
	exacts, exists := queries[exactKey]
	if exists {
		if len(exacts) != 1 {
			return filters.Exception(
				"Cannot exact match attribute '%s' to more than one value.",
				t.Key,
			)
		}

		valids = append(valids, TextExactFilter{
			BaseFilter: &schema.BaseFilter{
				QArgKey:    exactKey,
				QArgValues: exacts,
			},
			matching: exacts[0],
			column:   t.ColumnName,
		})
	}

	lengthKey := t.Key + filters.LENGTH_SUFFIX
	lengths, exists := queries[lengthKey]
	if exists {
		if len(lengths) != 1 {
			return filters.Exception(
				"Cannot length compare attribute '%s' to more than one length.",
				t.Key,
			)
		}

		length := lengths[0]
		lengthInt, err := strconv.Atoi(length)
		if err != nil {
			return filters.Exception(
				"Invalid length match specified on attribute '%s'.",
				t.Key,
			)
		}

		valids = append(valids, TextLengthExactFilter{
			BaseFilter: &schema.BaseFilter{
				QArgKey:    lengthKey,
				QArgValues: []string{strconv.Itoa(lengthInt)},
			},
			length: lengthInt,
			column: t.ColumnName,
		})
	}

	containsKey := t.Key + filters.CONTAINS_SUFFIX
	contains, exists := queries[containsKey]
	if exists {
		lowerContains := []string{}
		for _, match := range contains {
			lowerContains = append(lowerContains, strings.ToLower(match))
		}

		valids = append(valids, TextContainsFilter{
			BaseFilter: &schema.BaseFilter{
				QArgKey:    containsKey,
				QArgValues: lowerContains,
			},
			contains: lowerContains,
			column:   t.ColumnName,
		})
	}

	lesserKey := t.Key + filters.LENGTH_SUFFIX + filters.LT_SUFFIX
	lts, exists := queries[lesserKey]
	if exists {
		if len(lts) != 1 {
			return filters.Exception(
				"Cannot compare length of attribute '%s' to more than one value.",
				t.Key,
			)
		}

		lt := lts[0]
		ltInt, err := strconv.Atoi(lt)
		if err != nil || ltInt < 1 {
			return filters.Exception(
				"Invalid length comparison specified on attribute '%s'.",
				t.Key,
			)
		}
		valids = append(valids, TextLengthLesserFilter{
			BaseFilter: &schema.BaseFilter{
				QArgKey:    lesserKey,
				QArgValues: []string{strconv.Itoa(ltInt)},
			},
			length: ltInt,
			column: t.ColumnName,
			lesser: true,
		})
	}

	greaterKey := t.Key + filters.LENGTH_SUFFIX + filters.GT_SUFFIX
	gts, exists := queries[greaterKey]
	if exists {
		if len(gts) != 1 {
			return filters.Exception(
				"Cannot compare length of attribute '%s' to more than one value.",
				t.Key,
			)
		}

		gt := gts[0]
		gtInt, err := strconv.Atoi(gt)
		if err != nil {
			return filters.Exception(
				"Invalid length comparison specified on attribute '%s'.",
				t.Key,
			)
		}
		valids = append(valids, TextLengthLesserFilter{
			BaseFilter: &schema.BaseFilter{
				QArgKey:    greaterKey,
				QArgValues: []string{strconv.Itoa(gtInt)},
			},
			length: gtInt,
			column: t.ColumnName,
			lesser: false,
		})
	}

	return valids, nil
}
