package kubernetes

import (
	"github.com/mattfenwick/collections/pkg/set"
	"github.com/mattfenwick/collections/pkg/slice"
	"github.com/pkg/errors"
	goyaml "gopkg.in/yaml.v3"
	k8syaml "sigs.k8s.io/yaml"
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

// ParseObjectIntoType takes a weakly typed object -- such as an interface{}, a
//   map[string]interface{}, etc. -- marshals it into yaml, then unmarshals it
//   back into a strongly typed object -- such as a batchv1.Job
//   The goal is to convert a weakly typed object into a strongly typed one;
//   the yaml intermediary is just a convenient implementation detail.
func ParseObjectIntoType[A any](in interface{}) (*A, error) {
	yamlBytes, err := goyaml.Marshal(in)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to marshal yaml")
	}
	var out A
	err = k8syaml.UnmarshalStrict(yamlBytes, &out)
	return &out, errors.Wrapf(err, "unable to unmarshal k8s yaml")
}
