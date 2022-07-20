package swagger

import (
	"fmt"
	"github.com/mattfenwick/collections/pkg/builtin"
	"github.com/mattfenwick/collections/pkg/function"
	"github.com/mattfenwick/collections/pkg/slice"
	"github.com/mattfenwick/kube-utils/go/pkg/kubernetes"
	"github.com/mattfenwick/kube-utils/go/pkg/utils"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/maps"
)

type JsonPaths struct {
	Paths [][]string
}

func (j *JsonPaths) Append(path []string) {
	j.Paths = append(j.Paths, slice.Map(function.Id[string], path))
}

func (j *JsonPaths) GetSortedPaths() [][]string {
	return slice.SortBy(slice.CompareSlicePairwiseBy(builtin.CompareOrdered[string]), j.Paths)
}

func JsonFindPaths(obj interface{}) [][]string {
	paths := &JsonPaths{}
	bouncedObj, err := kubernetes.BounceMarshalGeneric[map[string]interface{}](obj)
	utils.DoOrDie(err)
	JsonFindPathsHelper(*bouncedObj, []string{}, paths)
	return paths.GetSortedPaths()
}

func JsonFindPathsHelper(obj interface{}, pathContext []string, paths *JsonPaths) {
	path := make([]string, len(pathContext))
	copy(path, pathContext)

	logrus.Debugf("path: %+v", path)

	if obj == nil {
		panic(errors.Errorf("unexpected nil at %+v", path))
	} else {
		switch val := obj.(type) {
		case map[string]interface{}:
			for _, k := range maps.Keys(val) {
				JsonFindPathsHelper(val[k], append(path, fmt.Sprintf(`["%s"]`, k)), paths)
			}
		case []interface{}:
			for i := range val {
				newPath := append(path, fmt.Sprintf("[%d]", i))
				JsonFindPathsHelper(val[i], newPath, paths)
			}
		case int:
			paths.Append(append(path, fmt.Sprintf("%T", val)))
		case string:
			paths.Append(append(path, fmt.Sprintf("%T", val)))
		case bool:
			paths.Append(append(path, fmt.Sprintf("%T", val)))
		//case types.Nil: // TODO is this necessary?
		default:
			panic(errors.Errorf("unrecognized type: %+v, %T", path, val))
		}
	}
}
