package utils

import "golang.org/x/exp/constraints"

func IfOr[T any](a bool, x, y T) T {
	if a {
		return x
	}
	return y
}

func SafeDivide[T constraints.Integer](x, y T) T {
	if y == 0 {
		return 0
	}

	return x / y
}
