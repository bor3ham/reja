package http

import (
	"errors"
	"strconv"
	"fmt"
)

func GetIntParam(
	params map[string][]string,
	key string,
	name string,
	defaultVal int,
	min *int,
	max *int,
) (int, error) {
	var value int
	values, provided := params[key]

	// not provided
	if !provided {
		return defaultVal, nil
	}

	// how could this even happen?
	if len(values) == 0 {
		panic("Provided but empty get parameter?")
	}
	// check for multiple values
	if len(values) > 1 {
		return 1, errors.New(fmt.Sprintf("There can only be one %s value.", name))
	}
	// check validity
	var err error
	value, err = strconv.Atoi(values[0])
	if err != nil {
		return 1, errors.New(fmt.Sprintf("%s must be an integer.", name))
	}
	// check maximum
	if max != nil {
		if value > *max {
			return 1, errors.New(
				fmt.Sprintf(
					"%s must be less than or equal to the maximum (%d).",
					name,
					*max,
				),
			)
		}
	}
	// check minimum
	if min != nil {
		if value < *min {
			return 1, errors.New(
				fmt.Sprintf(
					"%s must be greater than or equal to the minimum (%d).",
					name,
					*min,
				),
			)
		}
	}
	// success
	return value, nil
}
