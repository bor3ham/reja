package utils

import (
	"strings"
	"fmt"
)

func StringInAsArgs(nextArg int, in []string) (int, string) {
	spots := []string{}
	for _, _ = range in {
		spots = append(spots, fmt.Sprintf("$%d", nextArg))
		nextArg += 1
	}
	return nextArg, strings.Join(spots, ", ")
}
