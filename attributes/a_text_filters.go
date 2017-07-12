package attributes

import (
	"errors"
	"fmt"
	"strconv"
)

func filterException(text string, args ...interface{}) (map[string][]string, error) {
	return map[string][]string{}, errors.New(fmt.Sprintf(text, args...))
}

func (t Text) ValidateFilters(queries map[string][]string) (map[string][]string, error) {
	valids := map[string][]string{}

	// exact match
	matchingExact := false
	exacts, exists := queries[t.Key]
	if exists {
		matchingExact = true

		if len(exacts) != 1 {
			return filterException(
				"Cannot exact match attribute '%s' to more than one value.",
				t.Key,
			)
		}
		valids["exact"] = exacts
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

		if matchingExact && len(valids["exact"][0]) != lengthInt {
			return filterException(
				"Cannot exact match attribute '%s' and also match different length.",
				t.Key,
			)
		}

		valids["length"] = lengths
	}

	contains, exists := queries[t.Key+"__contains"]
	if exists {
		if matchingExact {
			return filterException(
				"Cannot exact match attribute '%s' and also look for contained value.",
				t.Key,
			)
		}
		valids["contains"] = contains
	}

	lts, exists := queries[t.Key+"__length__lt"]
	if exists {
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
		valids["lt"] = lts
	}

	gts, exists := queries[t.Key+"__length__gt"]
	if exists {
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
		_ = gtInt
		if err != nil {
			return filterException(
				"Invalid length comparison specified on attribute '%s'.",
				t.Key,
			)
		}
		valids["gt"] = gts
	}

	return valids, nil
}
