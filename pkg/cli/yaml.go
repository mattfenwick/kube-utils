package cli

import (
	"github.com/mattfenwick/kube-utils/pkg/kubernetes"
	"github.com/mattfenwick/kube-utils/pkg/utils"
	"github.com/spf13/cobra"
)

func SetupAnalyzeYamlCommand() *cobra.Command {
	args := &kubernetes.YamlAnalysisArgs{}

	command := &cobra.Command{
		Use:   "analyze-yaml",
		Short: "analyze yaml",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, as []string) {
			RunAnalyzeYaml(args)
		},
	}

	command.Flags().StringVar(&args.ChartPath, "chart-path", "", "path to yaml file")
	utils.DoOrDie(command.MarkFlagRequired("chart-path"))

	command.Flags().BoolVar(&args.PrintSkipped, "print-skipped", true, "if true, prints skipped resources")
	command.Flags().StringSliceVar(&args.Resources, "resources", []string{}, "pod-owning resources to print; if empty, all are printed")

	return command
}

func RunAnalyzeYaml(args *kubernetes.YamlAnalysisArgs) {
	kubernetes.RunYamlAnalysis(args)
}
