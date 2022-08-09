package json_traversal

import (
	"github.com/mattfenwick/kube-utils/pkg/utils"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/maps"
	"regexp"
	"strconv"
)

type Selector struct {
	IsGlob  bool
	IsArray bool
	Key     *string
}

type Result struct {
	Path  []*PathComponent
	Value interface{}
}

func JsonFindBySelector(obj interface{}, selector []*Selector, context []*PathComponent) []*Result {
	logrus.Debugf("JsonFindBySelector: %s", PathString(context))
	if len(selector) == 0 {
		return []*Result{{Path: context, Value: obj}}
	}

	next := selector[0]
	var results []*Result

	if next.IsGlob {
		logrus.Debugf("searching under glob")
		switch o := obj.(type) {
		case []interface{}:
			logrus.Debugf("searching under glob array")
			for i, e := range o {
				index := i
				results = append(results,
					JsonFindBySelector(e, selector[1:], append(context, NewArrayPathComponent(index)))...)
			}
		case map[string]interface{}:
			logrus.Debugf("searching under glob map")
			for k, v := range o {
				results = append(results,
					JsonFindBySelector(v, selector[1:], append(context, NewMapValuePathComponent(k)))...)
			}
		default:
			panic(errors.Errorf("can only glob slice or map, found %T", o))
		}
	} else {
		logrus.Debugf("searching under key")
		switch o := obj.(type) {
		case []interface{}:
			if next.IsArray {
				logrus.Debugf("searching under array")
				index64, err := strconv.ParseInt(*next.Key, 10, 32)
				index := int(index64)
				utils.DoOrDie(errors.Wrapf(err, "unable to ParseInt from %s", *next.Key))
				if index < len(o) {
					results = append(results,
						JsonFindBySelector(o[index], selector[1:], append(context, NewArrayPathComponent(index)))...)
				}
			}
		case map[string]interface{}:
			if !next.IsArray {
				logrus.Debugf("searching under map")
				v, ok := o[*next.Key]
				logrus.Debugf("key '%s' in map? %t", *next.Key, ok)
				if ok {
					results = append(results,
						JsonFindBySelector(v, selector[1:], append(context, NewMapValuePathComponent(*next.Key)))...)
				} else {
					logrus.Infof("did not find key %s; keys: %+v", *next.Key, maps.Keys(o))
				}

				for k := range o {
					logrus.Debugf("key in map: %s", k)
				}
			}
		default:
			logrus.Debugf("skipping type %T", o)
		}
	}
	return results
}

type KeyMatch struct {
	Path  []*PathComponent
	Value string
}

func (k *KeyMatch) PathString() []string {
	return PathString(k.Path)
}

func JsonFindByRegex(obj interface{}, path []*PathComponent, re *regexp.Regexp) []*KeyMatch {
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
			matches = append(matches, JsonFindByRegex(e, append(path, &PathComponent{ArrayIndex: &index}), re)...)
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
			matches = append(matches, JsonFindByRegex(v, append(path, &PathComponent{MapValue: &key}), re)...)
		}
		return matches
	default:
		logrus.Tracef("nothing to find: %T (path: %+v)", o, PathString(path))
		return nil
	}
}
