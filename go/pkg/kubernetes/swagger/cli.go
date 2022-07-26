package swagger

import (
	"github.com/mattfenwick/collections/pkg/slice"
	"github.com/sirupsen/logrus"
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

var (
	defaultExcludeResources = []string{"WatchEvent", "DeleteOptions"}
	defaultIncludeResources = []string{
		"Service",
		"ClusterRole",
		"ClusterRoleBinding",
		"ConfigMap",
		"CronJob",
		"CustomResourceDefinition",
		"Deployment",
		"Ingress",
		"Job",
		"Role",
		"RoleBinding",
		"Secret",
		"ServiceAccount",
		"StatefulSet",
	}
	defaultKubeVersions = []string{
		"1.18.20",
		"1.19.16",
		"1.20.15",
		"1.21.14",
		"1.22.12",
		"1.23.9",
		"1.24.3",
		"1.25.0-alpha.3",
	}
)

type CompareGVKArgs struct {
	ExcludeResources []string
	IncludeResources []string
	KubeVersions     []string
}

func (c *CompareGVKArgs) GetKubeVersions() []KubeVersion {
	return slice.Map(MustVersion, c.KubeVersions)
}

func setupCompareGvkCommand() *cobra.Command {
	args := &CompareGVKArgs{}

	command := &cobra.Command{
		Use:   "compare",
		Short: "compare gvks across kube versions",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, as []string) {
			RunCompareGvks(args)
		},
	}

	command.Flags().StringSliceVar(&args.ExcludeResources, "exclude", defaultExcludeResources, "resources to exclude")
	command.Flags().StringSliceVar(&args.IncludeResources, "include", defaultIncludeResources, "resources to include")
	command.Flags().StringSliceVar(&args.KubeVersions, "kube-version", defaultKubeVersions, "kube versions to compare across")

	return command
}

func RunCompareGvks(args *CompareGVKArgs) {
	logrus.Debugf("running 'compare gvk' with: exclude %+v, include %+v, kube versions %+v", args.ExcludeResources, args.IncludeResources, args.KubeVersions)
	CompareGVKs(args.ExcludeResources, args.IncludeResources, args.GetKubeVersions())
}

type ExplainResourceArgs struct {
	Format        string
	GroupVersions []string
	TypeNames     []string
	Version       string
}

func setupExplainResourceCommand() *cobra.Command {
	args := &ExplainResourceArgs{}

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

type CompareResourceArgs struct {
	Versions []string
	//GroupVersions []string // TODO ?
	TypeNames        []string
	SkipDescriptions bool
	PrintValues      bool
}

func setupCompareResourceCommand() *cobra.Command {
	args := &CompareResourceArgs{}

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
