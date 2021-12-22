package schema_json

import (
	"encoding/json"
	"fmt"
	"github.com/mattfenwick/kube-utils/go/pkg/utils"
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
	"github.com/r3labs/diff/v2"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"regexp"
	"sort"
	"strings"
)

func Executable() {
	//mode := "find-by-path-nested-items"
	//mode := "parse"
	mode := "diff"

	switch mode {
	case "diff":
		RunDiff()
	case "parse":
		RunParse()
	case "find-by-path-nested-items":
		RunFindByPathNestedItems()
	case "find-by-path":
		logrus.SetLevel(logrus.DebugLevel)
		RunFindByPath()
	case "find-by-regex":
		RunFindByRegex()
	}
}

func RunDiff() {
	path1, path2 := os.Args[1], os.Args[2]

	if true {
		spec1 := map[string]interface{}{}
		err := utils.ReadJson(path1, &spec1)
		utils.DoOrDie(err)
		spec2 := map[string]interface{}{}
		err = utils.ReadJson(path2, &spec2)
		utils.DoOrDie(err)
		diffs := &utils.Diffs{}
		utils.JsonDiff(spec1, spec2, []string{}, diffs)
		for _, d := range diffs.Elements {
			fmt.Printf("%s at %+v\n", d.Type, d.Path)
		}
	} else {
		spec1, err := ReadSwaggerSpecs(path1)
		utils.DoOrDie(err)
		spec2, err := ReadSwaggerSpecs(path2)
		utils.DoOrDie(err)

		// TODO this library only seems to report leaf properties that have been removed, even if whole branches are gone
		changelog, err := diff.Diff(spec1, spec2)
		utils.DoOrDie(err)

		//fmt.Printf("%+v\n", changelog)
		for _, change := range changelog {
			if change.Type == "update" && change.Path[len(change.Path)-1] == "Description" && strings.Contains(change.To.(string), "eprecate") {
				// TODO this logic doesn't really work
				fmt.Printf("found something getting deprecated: %+v\n", change.Path)
			} else {
				fmt.Printf("found a %s change: %+v\n", change.Type, change.Path)
			}
		}
	}
}

func RunParse() {
	path := os.Args[1]
	spec, err := ReadSwaggerSpecs(path)
	utils.DoOrDie(err)

	for name, t := range spec.Definitions {
		for propName, prop := range t.Properties {
			fmt.Printf("%s, %s: %+v\n<<>>\n", name, propName, prop.Items)
		}
	}

	outputPath := "out-sorted.json"
	err = utils.WriteJson(outputPath, spec)
	utils.DoOrDie(err)
	// must again unmarshal/marshal to get struct keys sorted
	err = utils.JsonUnmarshalMarshal(outputPath)
	utils.DoOrDie(err)

	fmt.Printf("%s\n", strings.Join(spec.VersionKindLengths(), "\n"))

	//fmt.Printf("spec:\n%+v\n", spec)
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

			bytes, err := ioutil.ReadFile(configPath)
			utils.DoOrDie(errors.Wrapf(err, "unable to read file %s", configPath))

			err = json.Unmarshal(bytes, &args)
			utils.DoOrDie(errors.Wrapf(err, "unable to unmarshal json"))

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
	err := errors.Wrapf(command.Execute(), "run root command")
	utils.DoOrDie(err)
}

func RunFindInJsonByRegex(args *FindByRegexArgs) {
	logrus.Infof("configuration: %s", args.Json())

	path := args.File
	regexString := args.Regex

	bytes, err := ioutil.ReadFile(path)
	utils.DoOrDie(errors.Wrapf(err, "unable to read file %s", path))

	var obj map[string]interface{}
	err = json.Unmarshal(bytes, &obj)
	utils.DoOrDie(errors.Wrapf(err, "unable to unmarshal json"))

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
