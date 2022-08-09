package kubernetes

import (
	"fmt"
	"github.com/mattfenwick/collections/pkg/set"
	"github.com/mattfenwick/collections/pkg/slice"
	"github.com/mattfenwick/collections/pkg/yaml"
	"github.com/mattfenwick/kube-utils/pkg/utils"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/maps"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
)

type YamlAnalysisArgs struct {
	ChartPath    string
	PrintSkipped bool
	Resources    []string
}

func RunYamlAnalysis(args *YamlAnalysisArgs) {
	objs, err := yaml.ParseManyFromFile[map[string]interface{}](args.ChartPath)
	utils.DoOrDie(err)
	model := NewModel()
	for _, m := range slice.SortOn(getResourceName, objs) {
		if m == nil {
			logrus.Debugf("skipping nil\n")
			continue
		}
		resourceName := getResourceName(m)
		kind := m["kind"].(string)
		logrus.Debugf("kind, name: %s, %s\n", kind, resourceName)
		switch kind {
		case "Deployment":
			dep, err := BounceMarshalGeneric[appsv1.Deployment](m)
			utils.DoOrDie(err)
			model.AddPodWrapper("Deployment", dep.Name, AnalyzeDeployment(dep))
		case "StatefulSet":
			sset, err := BounceMarshalGeneric[appsv1.StatefulSet](m)
			utils.DoOrDie(err)
			model.AddPodWrapper("StatefulSet", sset.Name, AnalyzeStatefulSet(sset))
		case "Job":
			job, err := BounceMarshalGeneric[batchv1.Job](m)
			utils.DoOrDie(err)
			model.AddPodWrapper("Job", job.Name, AnalyzeJob(job))
		case "CronJob":
			cj, err := BounceMarshalGeneric[batchv1.CronJob](m)
			utils.DoOrDie(err)
			model.AddPodWrapper("CronJob", cj.Name, AnalyzeCronJob(cj))
		case "Secret":
			model.Secrets = append(model.Secrets, resourceName)
		case "ConfigMap":
			model.ConfigMaps = append(model.ConfigMaps, resourceName)
		default:
			model.AddSkippedResource(kind, resourceName)
		}
	}

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
