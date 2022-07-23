package swagger

import (
	"github.com/spf13/cobra"
)

func SetupResourceCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "resource",
		Short: "kube openapi resource schema tools",
		Args:  cobra.ExactArgs(0),
	}

	command.AddCommand(setupExplainResourceCommand())
	command.AddCommand(setupCompareResourceCommand())

	return command
}

func SetupGVKCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "gvk",
		Short: "kube openapi gvk (Group/Version/Kind) schema tools",
		Args:  cobra.ExactArgs(0),
	}

	command.AddCommand(setupExplainGvkCommand())
	command.AddCommand(setupCompareGvkCommand())

	return command
}

func setupExplainGvkCommand() *cobra.Command {
	//args := &ParseSwaggerArgs{}

	command := &cobra.Command{
		Use:   "explain",
		Short: "explain gvks from a swagger spec",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, as []string) {
			RunExplainGvks()
		},
	}

	return command
}

func RunExplainGvks() {
	panic("TODO")
}

func setupCompareGvkCommand() *cobra.Command {
	//args := &ParseSwaggerArgs{}

	command := &cobra.Command{
		Use:   "compare",
		Short: "compare gvks across kube versions",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, as []string) {
			RunCompareGvks()
		},
	}

	return command
}

func RunCompareGvks() {
	CompareJsonSpecsAcrossKubeVersions()
}

type ExplainArgs struct {
	Format        string
	GroupVersions []string
	TypeNames     []string
	Version       string
}

func setupExplainResourceCommand() *cobra.Command {
	args := &ExplainArgs{}

	command := &cobra.Command{
		Use:   "explain",
		Short: "explain types from a swagger spec",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, as []string) {
			RunExplainResource(args)
		},
	}

	command.Flags().StringVar(&args.Format, "format", "condensed", "output format")
	command.Flags().StringSliceVar(&args.GroupVersions, "group-version", []string{}, "group/versions to look for type under; looks under all if not specified")
	command.Flags().StringSliceVar(&args.TypeNames, "type", []string{}, "kubernetes types to explain")
	command.Flags().StringVar(&args.Version, "version", "1.23.0", "kubernetes spec version")

	return command
}

type CompareArgs struct {
	Versions []string
	//GroupVersions []string // TODO ?
	TypeNames        []string
	SkipDescriptions bool
	PrintValues      bool
}

func setupCompareResourceCommand() *cobra.Command {
	args := &CompareArgs{}

	command := &cobra.Command{
		Use:   "compare",
		Short: "compare types across kube versions",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, as []string) {
			RunCompareResource(args)
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
