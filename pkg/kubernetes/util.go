package kubernetes

import (
	"github.com/mattfenwick/collections/pkg/set"
	"github.com/mattfenwick/collections/pkg/slice"
)

type KeySetComparison struct {
	JustA []string
	Both  []string
	JustB []string
}

func CompareKeySets(a *set.Set[string], b *set.Set[string]) *KeySetComparison {
	return &KeySetComparison{
		JustA: slice.Sort(a.Difference(b).ToSlice()),
		Both:  slice.Sort(a.Intersect(b).ToSlice()),
		JustB: slice.Sort(b.Difference(a).ToSlice()),
	}
}
