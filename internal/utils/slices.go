package utils

import "strings"

func SlicesHaveSubstrings(s []string, substrings ...string) bool {
	if len(s) == 0 {
		return false
	}

	if len(substrings) == 0 {
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
