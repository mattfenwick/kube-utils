package swagger

import (
	"fmt"
	"github.com/mattfenwick/collections/pkg/file"
	"github.com/mattfenwick/collections/pkg/json"
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

	//options := &json.MarshalOptions{EscapeHTML: false, Indent: true, Sort: true}
	//utils.DoOrDie(json.MarshalToFileOptions(spec, "my-spec-1.txt", options))
	//utils.DoOrDie(json.MarshalToFileOptions(spec2, "my-spec-2.txt", options))

	sortedSpecBytes, err := json.SortOptions(specBytes, false, true)
	utils.DoOrDie(err)
	utils.DoOrDie(file.Write("my-spec-1.txt", sortedSpecBytes, 0644))
	sortedSpecStringBytes, err := json.SortOptions([]byte(specString), false, true)
	utils.DoOrDie(err)
	utils.DoOrDie(file.Write("my-spec-2.txt", sortedSpecStringBytes, 0644))

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
