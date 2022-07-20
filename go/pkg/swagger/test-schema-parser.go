package swagger

import (
	"fmt"
	"github.com/mattfenwick/kube-utils/go/pkg/utils"
	"github.com/spf13/cobra"
	"os/exec"
	"reflect"
)

func setupTestSchemaParserCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "test-schema-parser",
		Short: "make sure schema parser handles everything",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, as []string) {
			TestSchemaParser()
		},
	}

	return command
}

func TestSchemaParser() {
	version := "1.8.15"

	specBytes := DownloadSwaggerSpec(version)
	spec, err := utils.ParseJson[Spec](specBytes)
	utils.DoOrDie(err)

	specString := utils.JsonString(spec)
	spec2, err := utils.ParseJson[Spec]([]byte(specString))
	utils.DoOrDie(err)

	utils.DoOrDie(utils.WriteFile("my-spec-1.txt", utils.JsonUnmarshalMarshalFromBytes(specBytes), 0644))
	utils.DoOrDie(utils.WriteFile("my-spec-2.txt", utils.JsonUnmarshalMarshalFromBytes([]byte(specString)), 0644))

	diff, err := utils.CommandRun(exec.Command("git", "diff", "--no-index", "my-spec-1.txt", "my-spec-2.txt"))
	utils.DoOrDie(err)
	fmt.Printf("%s\n", diff)
	//spec := MustReadSwaggerSpec(version)
	//specString := utils.JsonString(spec)
	//
	//spec2, err := utils.ParseJson[Spec]([]byte(specString))
	//utils.DoOrDie(err)
	//specString2 := utils.JsonString(spec2)

	fmt.Printf("same? %t, %t\n", reflect.DeepEqual(spec, spec2), specString == string(specBytes))
}
