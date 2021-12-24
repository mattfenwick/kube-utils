package schema_json

import (
	"fmt"
	"github.com/mattfenwick/kube-utils/go/pkg/utils"
	"github.com/pkg/errors"
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

	command.PersistentFlags().StringVarP(&verbosity, "verbosity", "v", "info", "log level; one of [info, debug, trace, warn, error, fatal, panic]")

	command.AddCommand(setupSwaggerResolveCommand())

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
	command.Flags().StringSliceVar(&args.GroupVersions, "group-versions", []string{}, "group/versions to look for type under; looks under all if not specified")
	command.Flags().StringVar(&args.TypeName, "type", "", "kubernetes type")
	utils.DoOrDie(command.MarkFlagRequired("type"))
	command.Flags().StringVar(&args.Version, "version", "1.23.0", "kubernetes spec version")

	return command
}

func RunResolve(args *ResolveArgs) {
	// TODO either guarantee the data is present, or curl it
	path := fmt.Sprintf("./swagger-data/%s-swagger-spec.json", args.Version)
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
			fmt.Printf("%s.%s:\n%s\n", groupVersion, typeName, SwaggerAnalysisTypeTable(analysis))
		case "condensed":
			fmt.Printf("%s.%s:\n%s\n", groupVersion, typeName, strings.Join(SwaggerAnalysisTypeSummary(analysis), "\n"))
		default:
			panic(errors.Errorf("invalid output format: %s", args.Format))
		}
	}
}
