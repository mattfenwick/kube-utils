package main

import (
	"encoding/json"
	"fmt"
	"github.com/mattfenwick/kube-utils/go/pkg/simulator"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
)

func main() {
	path := os.Args[1]
	regexString := os.Args[2]

	bytes, err := ioutil.ReadFile(path)
	simulator.DoOrDie(errors.Wrapf(err, "unable to read file %s", path))

	var obj map[string]interface{}
	err = json.Unmarshal(bytes, &obj)
	simulator.DoOrDie(errors.Wrapf(err, "unable to unmarshal json"))

	re := regexp.MustCompile(regexString)
	matches := FindInJson(obj, nil, re)

	for _, match := range matches {
		fmt.Printf("%s: %s\n", strings.Join(match.PathString(), ""), match.Value)
	}
}

type PathComponent struct {
	ArrayIndex *int
	MapKey     *string
	MapValue   *string
}

type KeyMatch struct {
	Path  []*PathComponent
	Value string
}

func (k *KeyMatch) PathString() []string {
	var path []string
	for _, component := range k.Path {
		if component.MapKey != nil {
			path = append(path, fmt.Sprintf(`{"%s"}`, *component.MapKey))
		} else if component.MapValue != nil {
			path = append(path, fmt.Sprintf(`["%s"]`, *component.MapValue))
		} else if component.ArrayIndex != nil {
			path = append(path, fmt.Sprintf(`[%d]`, *component.ArrayIndex))
		} else {
			simulator.DoOrDie(errors.Errorf("invalid PathComponent: %+v", k))
		}
	}
	return path
}

func FindInJson(obj interface{}, path []*PathComponent, re *regexp.Regexp) []*KeyMatch {
	switch o := obj.(type) {
	case string:
		if re.FindString(o) != "" {
			return []*KeyMatch{{
				Path:  path,
				Value: o,
			}}
		}
		return nil
	case []interface{}:
		var matches []*KeyMatch
		for i, e := range o {
			index := i
			matches = append(matches, FindInJson(e, append(path, &PathComponent{ArrayIndex: &index}), re)...)
		}
		return matches
	case map[string]interface{}:
		var matches []*KeyMatch
		for k, v := range o {
			key := k
			if re.FindString(k) != "" {
				matches = append(matches, &KeyMatch{
					Path:  append(path, &PathComponent{MapKey: &key}),
					Value: key,
				})
			}
			matches = append(matches, FindInJson(v, append(path, &PathComponent{MapValue: &key}), re)...)
		}
		return matches
	default:
		logrus.Debugf("nothing to find: %T (path: %+v)", o, path)
		return nil
	}
}
