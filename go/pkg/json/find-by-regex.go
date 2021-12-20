package json

import (
	"encoding/json"
	"fmt"
	"github.com/mattfenwick/kube-utils/go/pkg/simulator"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"
)

type Args struct {
	File      string
	Regex     string
	StartPath []string
}

func (a *Args) Json() string {
	bytes, err := json.MarshalIndent(a, "", "  ")
	simulator.DoOrDie(errors.Wrapf(err, "unable to marshal json for Args"))
	return string(bytes)
}

func (a *Args) StartPathSelector() []*Selector {
	var selectors []*Selector
	for _, component := range a.StartPath {
		var selector *Selector
		if component == "*" {
			selector = &Selector{IsGlob: true}
		} else if component[0] == '"' {
			key := component[1 : len(component)-1]
			selector = &Selector{IsArray: false, IsGlob: false, Key: &key}
		} else {
			key := component
			selector = &Selector{IsArray: true, IsGlob: false, Key: &key}
		}
		selectors = append(selectors, selector)
	}
	return selectors
}

func setupCommand() *cobra.Command {
	var configPath string

	command := &cobra.Command{
		Use:   "find-json",
		Short: "find strings in a JSON blob",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, as []string) {
			args := &Args{}

			bytes, err := ioutil.ReadFile(configPath)
			simulator.DoOrDie(errors.Wrapf(err, "unable to read file %s", configPath))

			err = json.Unmarshal(bytes, &args)
			simulator.DoOrDie(errors.Wrapf(err, "unable to unmarshal json"))

			RunFindInJson(args)
		},
	}

	command.Flags().StringVar(&configPath, "config-path", "", "path to json config file")
	simulator.DoOrDie(command.MarkFlagRequired("config-path"))

	//command.Flags().StringVar(&args.File, "file", "", "json file in which to search")
	//
	//command.Flags().StringVar(&args.Regex, "regex", "", "regex to search for")
	//simulator.DoOrDie(command.MarkFlagRequired("regex"))
	//
	//command.Flags().StringSliceVar(&args.StartPath, "start-path", []string{}, "path to search under")

	return command
}

func RunFindByRegex() {
	logrus.SetLevel(logrus.DebugLevel)

	command := setupCommand()
	err := errors.Wrapf(command.Execute(), "run root command")
	simulator.DoOrDie(err)
}

func RunFindInJson(args *Args) {
	logrus.Infof("configuration: %s", args.Json())

	path := args.File
	regexString := args.Regex

	bytes, err := ioutil.ReadFile(path)
	simulator.DoOrDie(errors.Wrapf(err, "unable to read file %s", path))

	var obj map[string]interface{}
	err = json.Unmarshal(bytes, &obj)
	simulator.DoOrDie(errors.Wrapf(err, "unable to unmarshal json"))

	re := regexp.MustCompile(regexString)
	var matches []*KeyMatch

	if len(args.StartPath) > 0 {
		pathSelectorResults := FindPathInObject(obj, args.StartPathSelector(), []*PathComponent{})

		for _, result := range pathSelectorResults {
			logrus.Infof("searching under: %s", PathString(result.Path))
			matches = append(matches, FindInJson(Select(obj, result.Path), result.Path, re)...)
		}
	} else {
		matches = append(matches, FindInJson(obj, []*PathComponent{}, re)...)
	}

	for _, match := range matches {
		fmt.Printf("%s: %s\n", strings.Join(match.PathString(), ""), match.Value)
	}
}

func Select(obj interface{}, path []*PathComponent) interface{} {
	for _, component := range path {
		switch o := obj.(type) {
		case []interface{}:
			if component.ArrayIndex != nil {
				if len(o) > *component.ArrayIndex {
					obj = o[*component.ArrayIndex]
				} else {
					return nil
				}
			} else {
				return nil
			}
		case map[string]interface{}:
			if component.MapKey != nil {
				if _, ok := o[*component.MapKey]; ok {
					obj = *component.MapKey
				} else {
					return nil
				}
			} else if component.MapValue != nil {
				obj = o[*component.MapValue]
			} else {
				return nil
			}
		default:
			return nil
		}
	}
	return obj
}

type Selector struct {
	IsGlob  bool
	IsArray bool
	Key     *string
}

type Result struct {
	Path  []*PathComponent
	Value interface{}
}

func FindPathInObject(obj interface{}, selector []*Selector, context []*PathComponent) []*Result {
	logrus.Debugf("FindPathInObject: %s", PathString(context))
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
					FindPathInObject(e, selector[1:], append(context, NewArrayPathComponent(index)))...)
			}
		case map[string]interface{}:
			logrus.Debugf("searching under glob map")
			for k, v := range o {
				results = append(results,
					FindPathInObject(v, selector[1:], append(context, NewMapValuePathComponent(k)))...)
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
				simulator.DoOrDie(errors.Wrapf(err, "unable to ParseInt from %s", *next.Key))
				if index < len(o) {
					results = append(results,
						FindPathInObject(o[index], selector[1:], append(context, NewArrayPathComponent(index)))...)
				}
			}
		case map[string]interface{}:
			if !next.IsArray {
				logrus.Debugf("searching under map")
				v, ok := o[*next.Key]
				logrus.Debugf("key '%s' in map? %t", *next.Key, ok)
				if ok {
					results = append(results,
						FindPathInObject(v, selector[1:], append(context, NewMapValuePathComponent(*next.Key)))...)
				}

				// debug
				for k := range o {
					logrus.Infof("key in map: %s", k)
				}
			}
		default:
			logrus.Debugf("skipping type %T", o)
		}
	}
	return results
}

type PathComponent struct {
	ArrayIndex *int
	MapKey     *string
	MapValue   *string
}

func NewArrayPathComponent(index int) *PathComponent {
	return &PathComponent{ArrayIndex: &index}
}

func NewMapKeyPathComponent(key string) *PathComponent {
	return &PathComponent{MapKey: &key}
}

func NewMapValuePathComponent(key string) *PathComponent {
	return &PathComponent{MapValue: &key}
}

type KeyMatch struct {
	Path  []*PathComponent
	Value string
}

func (k *KeyMatch) PathString() []string {
	return PathString(k.Path)
}

func PathString(components []*PathComponent) []string {
	var path []string
	for _, component := range components {
		if component.MapKey != nil {
			path = append(path, fmt.Sprintf(`{"%s"}`, *component.MapKey))
		} else if component.MapValue != nil {
			path = append(path, fmt.Sprintf(`["%s"]`, *component.MapValue))
		} else if component.ArrayIndex != nil {
			path = append(path, fmt.Sprintf(`[%d]`, *component.ArrayIndex))
		} else {
			simulator.DoOrDie(errors.Errorf("invalid PathComponent: %+v", component))
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
		logrus.Tracef("nothing to find: %T (path: %+v)", o, PathString(path))
		return nil
	}
}
