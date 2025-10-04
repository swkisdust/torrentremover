package utils

import (
	"slices"
	"strings"
)

func SlicesHave[S ~[]E, E comparable](s S, es ...E) bool {
	if len(s) < 1 || len(es) < 1 {
		return false
	}

	for _, e := range es {
		if slices.Contains(s, e) {
			return true
		}
	}
	return false
}

func SlicesHaveSubstrings(s []string, substrings ...string) bool {
	if len(s) < 1 || len(substrings) < 1 {
		return false
	}

	for _, sItem := range s {
		for _, sub := range substrings {
			if strings.Contains(sItem, sub) {
				return true
			}
		}
	}
	return false
}

func SliceMap[S ~[]EIn, EIn, EOut any](s S, f func(e EIn) EOut) []EOut {
	return slices.Collect(IterMap(slices.Values(s), f))
}
