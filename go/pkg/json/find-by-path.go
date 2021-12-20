package json

import (
	"encoding/json"
	"fmt"
	"github.com/mattfenwick/kube-utils/go/pkg/simulator"
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
	"sort"
	"strings"
)

func Pointer(s string) *string {
	return &s
}

func RunFindByPath() {
	path := os.Args[1]

	selector := []*Selector{
		{Key: Pointer("definitions")},
		{IsGlob: true},
		{Key: Pointer("x-kubernetes-group-version-kind")},
		{Key: Pointer("0"), IsArray: true},
		{Key: Pointer("kind")},
	}
	// ["definitions"]["io.k8s.api.extensions.v1beta1.Ingress"]["x-kubernetes-group-version-kind"][0]["kind"]

	bytes, err := ioutil.ReadFile(path)
	simulator.DoOrDie(errors.Wrapf(err, "unable to read file %s", path))

	var obj map[string]interface{}
	err = json.Unmarshal(bytes, &obj)
	simulator.DoOrDie(errors.Wrapf(err, "unable to unmarshal json"))

	results := FindPathInObject(obj, selector, []*PathComponent{})

	var values []string
	for _, result := range results {
		fmt.Printf("value %+v at %s\n", result.Value, PathString(result.Path))
		values = append(values, result.Value.(string))
	}
	sort.Strings(values)

	fmt.Printf("%s\n", strings.Join(values, "\n"))
}
