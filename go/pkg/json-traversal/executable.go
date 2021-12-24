package json_traversal

import (
	"fmt"
	"github.com/mattfenwick/kube-utils/go/pkg/utils"
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"regexp"
	"sort"
	"strings"
)

func Executable() {
	mode := "find-by-path-nested-items"

	switch mode {
	case "find-by-path-nested-items":
		RunFindByPathNestedItems()
	case "find-by-path":
		logrus.SetLevel(logrus.DebugLevel)
		RunFindByPath()
	case "find-by-regex":
		RunFindByRegex()
	default:
		panic("invalid mode")
	}
}
func RunFindByPathNestedItems() {
	path := os.Args[1]

	logrus.SetLevel(logrus.DebugLevel)

	nestedItemsSelector := []*Selector{
		{Key: utils.Pointer("definitions")},
		{IsGlob: true},
		{Key: utils.Pointer("properties")},
		{Key: utils.Pointer("items")},
		{Key: utils.Pointer("items")},
	}

	var obj map[string]interface{}
	err := utils.ReadJson(path, &obj)
	utils.DoOrDie(errors.Wrapf(err, "unable to read json from %s", path))

	results := JsonFindBySelector(obj, nestedItemsSelector, []*PathComponent{})

	if len(results) == 0 {
		fmt.Println("found 0 results")
	}

	for _, result := range results {
		fmt.Printf("result: - %s\n - %+v\n", PathString(result.Path), result.Value)
	}
}

func RunFindByPath() {
	path := os.Args[1]

	selector := []*Selector{
		{Key: utils.Pointer("definitions")},
		{IsGlob: true},
		{Key: utils.Pointer("x-kubernetes-group-version-kind")},
		{Key: utils.Pointer("0"), IsArray: true},
		{Key: utils.Pointer("kind")},
	}
	// ["definitions"]["io.k8s.api.extensions.v1beta1.Ingress"]["x-kubernetes-group-version-kind"][0]["kind"]

	var obj map[string]interface{}
	err := utils.ReadJson(path, obj)
	utils.DoOrDie(errors.Wrapf(err, "unable to read json from %s", path))

	results := JsonFindBySelector(obj, selector, []*PathComponent{})

	sort.Slice(results, func(i, j int) bool {
		return results[i].Value.(string) < results[j].Value.(string)
	})

	tableString := &strings.Builder{}
	table := tablewriter.NewWriter(tableString)
	table.SetAutoWrapText(false)
	table.SetRowLine(true)
	table.SetAutoMergeCells(true)
	table.SetHeader([]string{"group", "type"})
	for _, result := range results {
		//fmt.Printf("value %+v at %s\n", result.Value, PathString(result.Path))
		//fmt.Printf("%s/%s\n", result.Path[1].RawString(), result.Value)
		table.Append([]string{result.Path[1].RawString(), result.Value.(string)})
	}
	table.Render()
	fmt.Printf("%s\n", tableString)
}

type FindByRegexArgs struct {
	File      string
	Regex     string
	StartPath []string
}

func (a *FindByRegexArgs) Json() string {
	bytes, err := utils.MarshalIndent(a, "", "  ")
	utils.DoOrDie(errors.Wrapf(err, "unable to marshal json for FindByRegexArgs"))
	return string(bytes)
}

func (a *FindByRegexArgs) StartPathSelector() []*Selector {
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

func setupFindByRegexCommand() *cobra.Command {
	var configPath string

	command := &cobra.Command{
		Use:   "find-json",
		Short: "find strings in a JSON blob",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, as []string) {
			args := &FindByRegexArgs{}
			utils.DoOrDie(utils.ReadJson(configPath, &args))
			RunFindInJsonByRegex(args)
		},
	}

	command.Flags().StringVar(&configPath, "config-path", "", "path to json config file")
	utils.DoOrDie(command.MarkFlagRequired("config-path"))

	//command.Flags().StringVar(&args.File, "file", "", "json file in which to search")
	//
	//command.Flags().StringVar(&args.Regex, "regex", "", "regex to search for")
	//utils.DoOrDie(command.MarkFlagRequired("regex"))
	//
	//command.Flags().StringSliceVar(&args.StartPath, "start-path", []string{}, "path to search under")

	return command
}

func RunFindByRegex() {
	logrus.SetLevel(logrus.DebugLevel)

	command := setupFindByRegexCommand()
	utils.DoOrDie(errors.Wrapf(command.Execute(), "run root command"))
}

func RunFindInJsonByRegex(args *FindByRegexArgs) {
	logrus.Infof("configuration: %s", args.Json())

	path := args.File
	regexString := args.Regex

	var obj map[string]interface{}

	utils.DoOrDie(utils.ReadJson(path, &obj))

	re := regexp.MustCompile(regexString)
	var matches []*KeyMatch

	if len(args.StartPath) > 0 {
		pathSelectorResults := JsonFindBySelector(obj, args.StartPathSelector(), []*PathComponent{})

		for _, result := range pathSelectorResults {
			logrus.Infof("searching under: %s", PathString(result.Path))
			matches = append(matches, JsonFindByRegex(JsonFindByPath(obj, result.Path), result.Path, re)...)
		}
	} else {
		matches = append(matches, JsonFindByRegex(obj, []*PathComponent{}, re)...)
	}

	for _, match := range matches {
		fmt.Printf("%s: %s\n", strings.Join(match.PathString(), ""), match.Value)
	}
}
