package cli

import (
	"github.com/mattfenwick/kube-utils/go/pkg/kubernetes"
	"github.com/mattfenwick/kube-utils/go/pkg/utils"
	"github.com/spf13/cobra"
)

type AnalyzeYamlArgs struct {
	Path string
}

func SetupAnalyzeYamlCommand() *cobra.Command {
	args := &AnalyzeYamlArgs{}

	command := &cobra.Command{
		Use:   "analyze-yaml",
		Short: "analyze yaml",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, as []string) {
			RunAnalyzeYaml(args)
		},
	}

	command.Flags().StringVar(&args.Path, "path", "", "path to yaml file")
	utils.DoOrDie(command.MarkFlagRequired("path"))

	return command
}

func RunAnalyzeYaml(args *AnalyzeYamlArgs) {
	kubernetes.RunAnalyzeExample(args.Path)
}
