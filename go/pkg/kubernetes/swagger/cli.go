package swagger

import (
	"github.com/mattfenwick/collections/pkg/set"
	"github.com/mattfenwick/collections/pkg/slice"
	"github.com/pkg/errors"
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

	defaultExcludeApiVersions = []string{}
	defaultIncludeApiVersions = []string{
		"v1",
		"apps.v1",
		"batch.v1",
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

type ExplainGVKArgs struct {
	GroupBy            string
	KubeVersions       []string
	IncludeApiVersions []string
	ExcludeApiVersions []string
	IncludeResources   []string
	ExcludeResources   []string
	IncludeAll         bool
	Diff               bool
	// TODO add flag to verify parsing?  by serializing/deserializing to check if it matches input?
}

func (e *ExplainGVKArgs) GetGroupBy() ExplainGVKGroupBy {
	switch e.GroupBy {
	case "resource":
		return ExplainGVKGroupByResource
	case "apiversion", "api-version":
		return ExplainGVKGroupByApiVersion
	default:
		panic(errors.Errorf("invalid group by value: %s", e.GroupBy))
	}
}

func setupExplainGvkCommand() *cobra.Command {
	args := &ExplainGVKArgs{}

	command := &cobra.Command{
		Use:   "explain",
		Short: "explain gvks from a swagger spec",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, as []string) {
			RunExplainGvks(args)
		},
	}

	command.Flags().BoolVar(&args.Diff, "diff", false, "if true, calculate a diff from kube version to kube version.  if true, simply print resources")

	command.Flags().BoolVar(&args.IncludeAll, "include-all", false, "if true, includes all apiversions and resources regardless of includes/excludes.  This is useful for debugging")

	command.Flags().StringVar(&args.GroupBy, "group-by", "resource", "what to group by: valid values are 'resource' and 'api-version'")
	command.Flags().StringSliceVar(&args.KubeVersions, "kube-version", defaultKubeVersions, "kube versions to explain")

	command.Flags().StringSliceVar(&args.ExcludeResources, "resource-exclude", []string{}, "resources to exclude")
	command.Flags().StringSliceVar(&args.IncludeResources, "resource", []string{}, "resources to include")

	command.Flags().StringSliceVar(&args.ExcludeApiVersions, "apiversion-exclude", []string{}, "api versions to exclude")
	command.Flags().StringSliceVar(&args.IncludeApiVersions, "apiversion", []string{}, "api versions to include")

	return command
}

func shouldAllow(s string, allows *set.Set[string], forbids *set.Set[string]) bool {
	return (allows.Len() == 0 || allows.Contains(s)) && !forbids.Contains(s)
}

func RunExplainGvks(args *ExplainGVKArgs) {
	var include func(apiVersion string, resource string) bool
	if args.IncludeAll {
		include = func(apiVersion string, resource string) bool {
			return true
		}
	} else {
		includeResources := set.NewSet(args.IncludeResources)
		excludeResources := set.NewSet(args.ExcludeResources)
		includeApiVersions := set.NewSet(args.IncludeApiVersions)
		excludeApiVersions := set.NewSet(args.ExcludeApiVersions)

		include = func(apiVersion string, resource string) bool {
			includeApiVersion := shouldAllow(apiVersion, includeApiVersions, excludeApiVersions)
			includeResource := shouldAllow(resource, includeResources, excludeResources)
			return includeApiVersion && includeResource
		}
	}
	ExplainGvks(args.GetGroupBy(), args.KubeVersions, include, args.Diff)
}

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
