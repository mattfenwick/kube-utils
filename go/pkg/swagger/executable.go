package swagger

import (
	"github.com/mattfenwick/kube-utils/go/pkg/utils"
	"github.com/pkg/errors"
)

func Executable() {
	command := setupRootCommand()
	utils.DoOrDie(errors.Wrapf(command.Execute(), "run root command"))
}
