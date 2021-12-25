package swagger

import (
	"fmt"
	"github.com/mattfenwick/kube-utils/go/pkg/utils"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"sort"
	"strings"
)

func setupRootCommand() *cobra.Command {
	var verbosity string

	command := &cobra.Command{
		Use:   "swagger",
		Short: "work with kube swagger spec",
		Args:  cobra.ExactArgs(0),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return SetUpLogger(verbosity)
		},
	}

	command.PersistentFlags().StringVar(&verbosity, "log-level", "info", "log level; one of [trace, debug, info, warn, error, fatal, panic]")

	command.AddCommand(setupExplainCommand())
	command.AddCommand(setupCompareCommand())
	command.AddCommand(setupParseCommand())

	return command
}

type ExplainArgs struct {
	Format        string
	GroupVersions []string
	TypeNames     []string
	Version       string
}

func setupExplainCommand() *cobra.Command {
	args := &ExplainArgs{}

	command := &cobra.Command{
		Use:   "explain",
		Short: "explain types from a swagger spec",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, as []string) {
			RunExplain(args)
		},
	}

	command.Flags().StringVar(&args.Format, "format", "condensed", "output format")
	command.Flags().StringSliceVar(&args.GroupVersions, "group-version", []string{}, "group/versions to look for type under; looks under all if not specified")
	command.Flags().StringSliceVar(&args.TypeNames, "type", []string{}, "kubernetes types to explain")
	command.Flags().StringVar(&args.Version, "version", "1.23.0", "kubernetes spec version")

	return command
}

func RunExplain(args *ExplainArgs) {
	swaggerSpec := MustReadSwaggerSpec(args.Version)

	// no types specified?  use them all
	//   otherwise, filter down to just the ones requested
	var typeNames []string
	if len(args.TypeNames) == 0 {
		for name := range swaggerSpec.DefinitionsByNameByGroup() {
			typeNames = append(typeNames, name)
		}
		sort.Strings(typeNames)
	} else {
		typeNames = args.TypeNames // TODO should this be sorted, or respect the input order?
	}

	for _, typeName := range typeNames {
		logrus.Debugf("analysing type %s", typeName)
		analyses := swaggerSpec.AnalyzeType(typeName)

		// no group/versions specified?  use them all
		//   otherwise, filter down to just the ones requested
		if len(args.GroupVersions) > 0 {
			filteredAnalyses := map[string]interface{}{}
			for _, groupVersion := range args.GroupVersions {
				if analysis, ok := analyses[groupVersion]; ok {
					filteredAnalyses[groupVersion] = analysis
				} else {
					logrus.Debugf("type %s not found under group/version %s (%+v)", typeName, groupVersion, SortedKeys(analyses))
				}
			}
			analyses = filteredAnalyses
		}

		gvks := SortedKeys(analyses)
		if len(gvks) == 0 {
			logrus.Debugf("no group/versions found for %s", typeName)
			continue
		}
		for _, groupVersion := range gvks {
			analysis := analyses[groupVersion]
			switch args.Format {
			case "table":
				fmt.Printf("%s.%s:\n%s\n", groupVersion, typeName, AnalysisTypeTable(analysis))
			case "condensed":
				fmt.Printf("%s.%s:\n%s\n", groupVersion, typeName, strings.Join(AnalysisTypeSummary(analysis), "\n"))
			default:
				panic(errors.Errorf("invalid output format: %s", args.Format))
			}
			fmt.Println()
		}
		fmt.Println()
	}
}

type CompareArgs struct {
	Versions []string
	//GroupVersions []string // TODO ?
	TypeNames        []string
	SkipDescriptions bool
	PrintValues      bool
}

func setupCompareCommand() *cobra.Command {
	args := &CompareArgs{}

	command := &cobra.Command{
		Use:   "compare",
		Short: "compare types from across swagger specs",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, as []string) {
			RunCompare(args)
		},
	}

	//command.Flags().StringSliceVar(&args.GroupVersions, "group-version", []string{}, "group/versions to look for type under; looks under all if not specified")
	//utils.DoOrDie(command.MarkFlagRequired("group-version"))

	command.Flags().StringSliceVar(&args.Versions, "version", []string{"1.18.19", "1.23.0"}, "kubernetes versions")

	command.Flags().StringSliceVar(&args.TypeNames, "type", []string{"Pod"}, "types to compare")

	command.Flags().BoolVar(&args.SkipDescriptions, "skip-descriptions", true, "if true, skip comparing descriptions (since these often change for non-functional reasons)")

	command.Flags().BoolVar(&args.PrintValues, "print-values", false, "if true, print values (in addition to just the path and change type)")

	return command
}

func RunCompare(args *CompareArgs) {
	if len(args.Versions) != 2 {
		panic(errors.Errorf("expected 2 kube versions, found %+v", args.Versions))
	}

	swaggerSpec1 := MustReadSwaggerSpec(args.Versions[0])
	swaggerSpec2 := MustReadSwaggerSpec(args.Versions[1])

	typeNames := map[string]interface{}{}
	if len(args.TypeNames) > 0 {
		for _, name := range args.TypeNames {
			typeNames[name] = true
		}
	} else {
		for name := range swaggerSpec1.DefinitionsByNameByGroup() {
			typeNames[name] = true
		}
		for name := range swaggerSpec2.DefinitionsByNameByGroup() {
			typeNames[name] = true
		}
	}

	for _, typeName := range SortedKeys(typeNames) {
		fmt.Printf("inspecting type %s\n", typeName)

		//resolved1 := swaggerSpec1.ResolveToJsonBlob(typeName)
		//resolved2 := swaggerSpec2.ResolveToJsonBlob(typeName)
		resolved1 := swaggerSpec1.AnalyzeType(typeName)
		resolved2 := swaggerSpec2.AnalyzeType(typeName)

		logrus.Infof("group/versions for kube %s: %+v", args.Versions[0], SortedKeys(resolved1))
		logrus.Infof("group/versions for kube %s: %+v", args.Versions[1], SortedKeys(resolved2))

		for _, groupName1 := range SortedKeys(resolved1) {
			type1 := resolved1[groupName1]
			for _, groupName2 := range SortedKeys(resolved2) {
				type2 := resolved2[groupName2]
				fmt.Printf("comparing %s: %s@%s vs. %s@%s\n", typeName, args.Versions[0], groupName1, args.Versions[1], groupName2)
				for _, e := range CompareAnalysisTypes(type1, type2).Elements {
					//for _, e := range utils.DiffJsonValues(utils.MustJsonRemarshal(type1), utils.MustJsonRemarshal(type2)).Elements {
					if len(e.Path) > 0 && e.Path[len(e.Path)-1] == "description" && args.SkipDescriptions {
						logrus.Debugf("skipping description at %+v", e.Path)
					} else {
						fmt.Printf("  %-20s    %+v\n", e.Type, strings.Join(e.Path, "."))
						if args.PrintValues {
							fmt.Printf("  - old: %+v\n  - new: %+v\n", e.Old, e.New)
						}
					}
				}
				fmt.Println()
			}
		}
	}
}

type ParseArgs struct {
	Version string
}

func setupParseCommand() *cobra.Command {
	args := &ParseArgs{}

	command := &cobra.Command{
		Use:   "parse",
		Short: "parse and serialize openapi spec for comparison (test command)",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, as []string) {
			RunParse(args)
		},
	}

	command.Flags().StringVar(&args.Version, "version", "1.23.0", "kubernetes version")

	return command
}

func RunParse(args *ParseArgs) {
	spec := MustReadSwaggerSpec(args.Version)

	for name, t := range spec.Definitions {
		for propName, prop := range t.Properties {
			logrus.Debugf("%s, %s: %+v\n<<>>\n", name, propName, prop.Items)
		}
	}

	// must do weird marshal/unmarshal/marshal dance to get struct keys sorted
	normalizedSpec := utils.MustJsonRemarshal(spec)
	bytes, err := utils.MarshalIndent(normalizedSpec, "", "  ")
	utils.DoOrDie(err)

	fmt.Printf("%s\n", bytes)
}
