package cli

import (
	"github.com/mattfenwick/kube-utils/pkg/utils"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func RunRootCommand() {
	command := SetupRootCommand()
	utils.DoOrDie(errors.Wrapf(command.Execute(), "run root command"))
}

type RootFlags struct {
	Verbosity string
}

func SetupRootCommand() *cobra.Command {
	flags := &RootFlags{}
	command := &cobra.Command{
		Use:   "utils",                      // TODO need unique name
		Short: "utilities for kube hacking", // TODO need focus
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return utils.SetUpLogger(flags.Verbosity)
		},
	}

	command.PersistentFlags().StringVarP(&flags.Verbosity, "verbosity", "v", "info", "log level; one of [info, debug, trace, warn, error, fatal, panic]")

	command.AddCommand(SetupVersionCommand())
	command.AddCommand(SetupAnalyzeYamlCommand())

	return command
}
