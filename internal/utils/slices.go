package utils

import (
	"slices"
	"strings"
)

func SlicesHas[S ~[]E, E comparable](s S, es ...E) bool {
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

func SlicesHasSubstrings(s []string, substrings ...string) bool {
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

func SlicesMap[S ~[]EIn, EIn, EOut any](s S, f func(e EIn) EOut) []EOut {
	return slices.Collect(IterMap(slices.Values(s), f))
}

func SlicesFilter[S ~[]E, E any](f func(E) bool, s S) S {
	return slices.Collect(Filter(f, slices.Values(s)))
}
