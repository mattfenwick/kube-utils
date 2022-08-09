package kubernetes

import (
	"github.com/mattfenwick/collections/pkg/slice"
	"github.com/mattfenwick/collections/pkg/yaml"
	"github.com/mattfenwick/kube-utils/pkg/utils"
	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
)

func RunYamlAnalysis(path string) {
	objs, err := yaml.ParseManyFromFile[map[string]interface{}](path)
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

	model.Tables()
}
