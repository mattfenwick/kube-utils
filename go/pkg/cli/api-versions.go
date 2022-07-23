package cli

import (
	apiversions "github.com/mattfenwick/kube-utils/go/pkg/kubernetes/swagger/apiversions"
	"github.com/spf13/cobra"
)

//type KindArgs struct {}

func setupKindCommand() *cobra.Command {
	//args := &KindArgs{}

	command := &cobra.Command{
		Use:   "kind",
		Short: "compare types from across swagger specs",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, as []string) {
			RunKind()
		},
	}

	return command
}

func RunKind() {
	apiversions.ParseKindResults()
}
