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

func SlicesHasSubstringsFunc[S ~[]E, E comparable](s S, f func(e E) string, substrings ...string) bool {
	if len(s) < 1 || len(substrings) < 1 {
		return false
	}

	for _, sItem := range s {
		for _, sub := range substrings {
			if strings.Contains(f(sItem), sub) {
				return true
			}
		}
	}
	return false
}

func SlicesMap[S ~[]EIn, EIn, EOut any](s S, f func(e EIn) EOut) []EOut {
	result := make([]EOut, len(s))

	for i := range s {
		result[i] = f(s[i])
	}

	return result
}

func SlicesFilter[S ~[]E, E any](f func(E) bool, s S) S {
	result := make(S, 0, len(s))

	for i := range s {
		if f(s[i]) {
			result = append(result, s[i])
		}
	}

	return result
}
