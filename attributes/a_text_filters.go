package attributes

import (
	"fmt"
	"strings"
	"strconv"
	"github.com/bor3ham/reja/schema"
)

type TextNullFilter struct {
	*BaseFilter
	null bool
	column string
}
func (f TextNullFilter) GetWhereQueries(nextArg int) []string {
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
func (f TextNullFilter) GetWhereArgs() []interface{} {
	return []interface{}{}
}

type TextExactFilter struct {
	*BaseFilter
	matching string
	column string
}
func (f TextExactFilter) GetWhereQueries(nextArg int) []string {
	return []string{
		fmt.Sprintf("%s = $%d", f.column, nextArg),
	}
}
func (f TextExactFilter) GetWhereArgs() []interface{} {
	return []interface{}{
		f.matching,
	}
}

type TextContainsFilter struct {
	*BaseFilter
	contains []string
	column string
}
func (f TextContainsFilter) GetWhereQueries(nextArg int) []string {
	where := "("
	for matchIndex, _ := range f.contains {
		if matchIndex > 0 {
			where += " or "
		}
		where += fmt.Sprintf(
			`%s ilike '%%' || $%d || '%%'`,
			f.column,
			nextArg,
		)
		nextArg += 1
	}
	where += ")"
	return []string{
		where,
	}
}
func (f TextContainsFilter) GetWhereArgs() []interface{} {
	args := []interface{}{}
	for _, match := range f.contains {
		args = append(args, strings.Replace(match, "%%", "\\%%", -1))
	}
	return args
}

type TextLengthExactFilter struct {
	*BaseFilter
	length int
	column string
}
func (f TextLengthExactFilter) GetWhereQueries(nextArg int) []string {
	return []string{
		fmt.Sprintf("char_length(%s) = $%d", f.column, nextArg),
	}
}
func (f TextLengthExactFilter) GetWhereArgs() []interface{} {
	return []interface{}{
		f.length,
	}
}

type TextLengthLesserFilter struct {
	*BaseFilter
	length int
	column string
}
func (f TextLengthLesserFilter) GetWhereQueries(nextArg int) []string {
	return []string{
		fmt.Sprintf("char_length(%s) < $%d", f.column, nextArg),
	}
}
func (f TextLengthLesserFilter) GetWhereArgs() []interface{} {
	return []interface{}{
		f.length,
	}
}

type TextLengthGreaterFilter struct {
	*BaseFilter
	length int
	column string
}
func (f TextLengthGreaterFilter) GetWhereQueries(nextArg int) []string {
	return []string{
		fmt.Sprintf("char_length(%s) > $%d", f.column, nextArg),
	}
}
func (f TextLengthGreaterFilter) GetWhereArgs() []interface{} {
	return []interface{}{
		f.length,
	}
}

func (t Text) ValidateFilters(queries map[string][]string) ([]schema.Filter, error) {
	valids := []schema.Filter{}

	// null check
	nullsOnly := false
	nonNullsOnly := false

	nullKey := t.Key+"__is_null"
	nullStrings, exists := queries[nullKey]
	if exists {
		if len(nullStrings) != 1 {
			return filterException(
				"Cannot null check attribute '%s' against more than one value.",
				t.Key,
			)
		}
		isNullString := strings.ToLower(nullStrings[0])
		if isNullString == "true" {
			nullsOnly = true
			valids = append(valids, TextNullFilter{
				BaseFilter: &BaseFilter{
					QArgKey: nullKey,
					QArgValues: []string{"true"},
				},
				null: true,
				column: t.ColumnName,
			})
		} else if isNullString == "false" {
			nonNullsOnly = true
			_ = nonNullsOnly
			valids = append(valids, TextNullFilter{
				BaseFilter: &BaseFilter{
					QArgKey: nullKey,
					QArgValues: []string{"false"},
				},
				null: false,
				column: t.ColumnName,
			})
		} else {
			return filterException(
				"Invalid null check value on attribute '%s'.",
				t.Key,
			)
		}
	}

	// exact match
	matchingExact := false
	exactMatch := ""

	exacts, exists := queries[t.Key]
	if exists {
		matchingExact = true
		exactMatch = exacts[0]

		if len(exacts) != 1 {
			return filterException(
				"Cannot exact match attribute '%s' to more than one value.",
				t.Key,
			)
		}

		if nullsOnly {
			return filterException(
				"Cannot match attribute '%s' to an exact value and null.",
				t.Key,
			)
		}

		valids = append(valids, TextExactFilter{
			BaseFilter: &BaseFilter{
				QArgKey: t.Key,
				QArgValues: exacts,
			},
			matching: exactMatch,
			column: t.ColumnName,
		})
	}

	// length match
	matchingExactLength := false
	lengths, exists := queries[t.Key+"__length"]
	if exists {
		matchingExactLength = true

		if len(lengths) != 1 {
			return filterException(
				"Cannot length compare attribute '%s' to more than one length.",
				t.Key,
			)
		}

		length := lengths[0]
		lengthInt, err := strconv.Atoi(length)
		if err != nil {
			return filterException(
				"Invalid length match specified on attribute '%s'.",
				t.Key,
			)
		}

		if nullsOnly {
			return filterException(
				"Cannot match attribute '%s' to an exact length and null.",
				t.Key,
			)
		}
		if matchingExact && len(exactMatch) != lengthInt {
			return filterException(
				"Cannot exact match attribute '%s' and also match different length.",
				t.Key,
			)
		}

		valids = append(valids, TextLengthExactFilter{
			BaseFilter: &BaseFilter{
				QArgKey: t.Key+"__length",
				QArgValues: []string{strconv.Itoa(lengthInt)},
			},
			length: lengthInt,
			column: t.ColumnName,
		})
	}

	contains, exists := queries[t.Key+"__contains"]
	if exists {
		if nullsOnly {
			return filterException(
				"Cannot match attribute '%s' to a contained value and null.",
				t.Key,
			)
		}
		if matchingExact {
			return filterException(
				"Cannot exact match attribute '%s' and also look for contained value.",
				t.Key,
			)
		}

		lowerContains := []string{}
		for _, match := range contains {
			lowerContains = append(lowerContains, strings.ToLower(match))
		}

		valids = append(valids, TextContainsFilter{
			BaseFilter: &BaseFilter{
				QArgKey: t.Key+"__contains",
				QArgValues: lowerContains,
			},
			contains: lowerContains,
			column: t.ColumnName,
		})
	}

	lts, exists := queries[t.Key+"__length__lt"]
	if exists {
		if nullsOnly {
			return filterException(
				"Cannot match attribute '%s' to null and compare length.",
				t.Key,
			)
		}
		if matchingExact {
			return filterException(
				"Cannot exact match attribute '%s' and also compare length.",
				t.Key,
			)
		}
		if matchingExactLength {
			return filterException(
				"Cannot exact match attribute '%s' length and also compare length.",
				t.Key,
			)
		}

		if len(lts) != 1 {
			return filterException(
				"Cannot compare length of attribute '%s' to more than one value.",
				t.Key,
			)
		}

		lt := lts[0]
		ltInt, err := strconv.Atoi(lt)
		if err != nil || ltInt < 1 {
			return filterException(
				"Invalid length comparison specified on attribute '%s'.",
				t.Key,
			)
		}
		valids = append(valids, TextLengthLesserFilter{
			BaseFilter: &BaseFilter{
				QArgKey: t.Key+"__length__lt",
				QArgValues: []string{strconv.Itoa(ltInt)},
			},
			length: ltInt,
			column: t.ColumnName,
		})
	}

	gts, exists := queries[t.Key+"__length__gt"]
	if exists {
		if nullsOnly {
			return filterException(
				"Cannot match attribute '%s' to null and compare length.",
				t.Key,
			)
		}
		if matchingExact {
			return filterException(
				"Cannot exact match attribute '%s' and also compare length.",
				t.Key,
			)
		}
		if matchingExactLength {
			return filterException(
				"Cannot exact match attribute '%s' length and also compare length.",
				t.Key,
			)
		}

		if len(gts) != 1 {
			return filterException(
				"Cannot compare length of attribute '%s' to more than one value.",
				t.Key,
			)
		}

		gt := gts[0]
		gtInt, err := strconv.Atoi(gt)
		if err != nil {
			return filterException(
				"Invalid length comparison specified on attribute '%s'.",
				t.Key,
			)
		}
		valids = append(valids, TextLengthGreaterFilter{
			BaseFilter: &BaseFilter{
				QArgKey: t.Key+"__length__gt",
				QArgValues: []string{strconv.Itoa(gtInt)},
			},
			length: gtInt,
			column: t.ColumnName,
		})
	}

	return valids, nil
}
