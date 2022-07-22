package swagger

import (
	"github.com/mattfenwick/collections/pkg/slice"
)

type Version []string

var (
	CompareVersion = slice.CompareSlicePairwise[string]()
)
