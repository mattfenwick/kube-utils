package swagger

import (
	"fmt"
	"github.com/mattfenwick/kube-utils/go/pkg/utils"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"strings"
)

func setupSwaggerCommand() *cobra.Command {
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

	command.AddCommand(setupSwaggerResolveCommand())
	command.AddCommand(setupSwaggerCompareCommand())
	command.AddCommand(setupSwaggerParseCommand())

	return command
}

type ResolveArgs struct {
	Format        string
	GroupVersions []string
	TypeName      string
	Version       string
}

func setupSwaggerResolveCommand() *cobra.Command {
	args := &ResolveArgs{}

	command := &cobra.Command{
		Use:   "resolve",
		Short: "resolve types from a swagger spec",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, as []string) {
			RunResolve(args)
		},
	}

	command.Flags().StringVar(&args.Format, "format", "condensed", "output format")
	command.Flags().StringSliceVar(&args.GroupVersions, "group-version", []string{}, "group/versions to look for type under; looks under all if not specified")
	command.Flags().StringVar(&args.TypeName, "type", "", "kubernetes type")
	utils.DoOrDie(command.MarkFlagRequired("type"))
	command.Flags().StringVar(&args.Version, "version", "1.23.0", "kubernetes spec version")

	return command
}

func RunResolve(args *ResolveArgs) {
	// TODO either guarantee the data is present, or curl it
	path := MakePathFromKubeVersion(args.Version)
	typeName := args.TypeName

	swaggerSpec, err := ReadSwaggerSpecs(path)
	utils.DoOrDie(err)

	analyses := swaggerSpec.AnalyzeType(typeName)

	// no group/versions specified?  use them all
	//   otherwise, filter down to just the ones requested
	if len(args.GroupVersions) > 0 {
		filteredAnalyses := map[string]interface{}{}
		for _, groupVersion := range args.GroupVersions {
			if analysis, ok := analyses[groupVersion]; ok {
				filteredAnalyses[groupVersion] = analysis
			} else {
				panic(errors.Errorf("type %s not found under group/version %s", args.TypeName, groupVersion))
			}
		}
		analyses = filteredAnalyses
	}

	for groupVersion, analysis := range analyses {
		switch args.Format {
		case "table":
			fmt.Printf("%s.%s:\n%s\n", groupVersion, typeName, AnalysisTypeTable(analysis))
		case "condensed":
			fmt.Printf("%s.%s:\n%s\n", groupVersion, typeName, strings.Join(AnalysisTypeSummary(analysis), "\n"))
		default:
			panic(errors.Errorf("invalid output format: %s", args.Format))
		}
	}
}

type CompareArgs struct {
	Versions []string
	//GroupVersions []string // TODO ?
	TypeNames        []string
	SkipDescriptions bool
	PrintValues      bool
}

func setupSwaggerCompareCommand() *cobra.Command {
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
	path1, path2 := MakePathFromKubeVersion(args.Versions[0]), MakePathFromKubeVersion(args.Versions[1])

	swaggerSpec1, err := ReadSwaggerSpecs(path1)
	utils.DoOrDie(err)
	swaggerSpec2, err := ReadSwaggerSpecs(path2)
	utils.DoOrDie(err)

	typeNames := map[string]bool{}
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

	for typeName := range typeNames {
		fmt.Printf("inspecting type %s\n", typeName)

		resolved1 := swaggerSpec1.ResolveToJsonBlob(typeName)
		resolved2 := swaggerSpec2.ResolveToJsonBlob(typeName)
		//resolved1 := swaggerSpec1.AnalyzeType(typeName)
		//resolved2 := swaggerSpec2.AnalyzeType(typeName)

		for groupName1, type1 := range resolved1 {
			for groupName2, type2 := range resolved2 {
				fmt.Printf("comparing %s: %s@%s vs. %s@%s\n", typeName, args.Versions[0], groupName1, args.Versions[1], groupName2)
				for _, e := range utils.DiffJsonValues(utils.MustJsonRemarshal(type1), utils.MustJsonRemarshal(type2)).Elements {
					if len(e.Path) > 0 && e.Path[len(e.Path)-1] == "description" && args.SkipDescriptions {
						logrus.Debugf("skipping description at %+v", e.Path)
					} else {
						fmt.Printf("  %s at %+v\n", e.Type, e.Path)
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

func MakePathFromKubeVersion(version string) string {
	return fmt.Sprintf("./swagger-data/%s-swagger-spec.json", version)
}

type ParseArgs struct {
	Version string
}

func setupSwaggerParseCommand() *cobra.Command {
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
	path := MakePathFromKubeVersion(args.Version)
	spec, err := ReadSwaggerSpecs(path)
	utils.DoOrDie(err)

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
