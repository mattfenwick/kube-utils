package kubernetes

import (
	"fmt"
	"github.com/mattfenwick/collections/pkg/set"
	"github.com/mattfenwick/collections/pkg/slice"
	"github.com/mattfenwick/collections/pkg/yaml"
	"github.com/mattfenwick/kube-utils/pkg/utils"
	"golang.org/x/exp/maps"
)

type YamlAnalysisArgs struct {
	ChartPath    string
	PrintSkipped bool
	Resources    []string
}

func RunYamlAnalysis(args *YamlAnalysisArgs) {
	objs, err := yaml.ParseManyFromFile[map[string]interface{}](args.ChartPath)
	utils.DoOrDie(err)
	model := NewModelFromYaml(objs)

	//fmt.Printf("%s\n", model.Graph().RenderAsDot())
	skipped, secret, configmaps, images, pods := model.BuildTables()

	if args.PrintSkipped {
		fmt.Printf("skipped resources:\n%s\n\n", skipped)
	}

	resourcesToPrint := set.FromSlice(args.Resources)
	allow := func(resource string) bool {
		return len(args.Resources) == 0 || resourcesToPrint.Contains(resource)
	}

	if allow("secret") {
		fmt.Printf("secrets:\n%s\n\n", secret)
	}
	if allow("configMap") {
		fmt.Printf("config maps:\n%s\n\n", configmaps)
	}
	if allow("image") {
		fmt.Printf("images:\n%s\n\n", images)
	}

	for _, kind := range slice.Sort(maps.Keys(pods)) {
		if allow(kind) {
			fmt.Printf("\nkind: %s\n%s\n", kind, pods[kind])
		}
	}
}
