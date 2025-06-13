package utils

import "iter"

func IterMap[In, Out any](seq iter.Seq[In], f func(In) Out) iter.Seq[Out] {
	return func(yield func(Out) bool) {
		for in := range seq {
			if !yield(f(in)) {
				return
			}
		}
	}
}

func Seq2To1[KIn, VIn, VOut any](seq iter.Seq2[KIn, VIn], f func(KIn, VIn) VOut) iter.Seq[VOut] {
	return func(yield func(VOut) bool) {
		for k, v := range seq {
			if !yield(f(k, v)) {
				return
			}
		}
	}
}

func Reduce[Sum, V any](f func(Sum, V) Sum, sum Sum, seq iter.Seq[V]) Sum {
	for v := range seq {
		sum = f(sum, v)
	}
	return sum
}
