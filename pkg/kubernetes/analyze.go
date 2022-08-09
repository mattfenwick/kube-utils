package kubernetes

import (
	"github.com/mattfenwick/collections/pkg/set"
	"github.com/mattfenwick/collections/pkg/slice"
	"github.com/mattfenwick/collections/pkg/yaml"
	"github.com/mattfenwick/kube-utils/pkg/utils"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	goyaml "gopkg.in/yaml.v3"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	k8syaml "sigs.k8s.io/yaml"
)

func BounceMarshalGeneric[A any](in interface{}) (*A, error) {
	yamlBytes, err := goyaml.Marshal(in)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to marshal yaml")
	}
	var out A
	err = k8syaml.UnmarshalStrict(yamlBytes, &out)
	return &out, errors.Wrapf(err, "unable to unmarshal k8s yaml")
}

func getResourceName(o map[string]interface{}) string {
	return o["metadata"].(map[string]interface{})["name"].(string)
}

func RunAnalyzeExample(path string) {
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

func analyzeVolumeMounts(isInitContainer bool, configMaps map[string]string, secrets map[string]string, containerSpec v1.Container) *Container {
	container := &Container{
		IsInit:     isInitContainer,
		Name:       containerSpec.Name,
		Image:      containerSpec.Image,
		ConfigMaps: set.FromSlice[string](nil),
		Secrets:    set.FromSlice[string](nil),
	}
	for _, mount := range containerSpec.VolumeMounts {
		if configMapName, ok := configMaps[mount.Name]; ok {
			container.ConfigMaps.Add(configMapName)
		} else if secretName, ok := secrets[mount.Name]; ok {
			container.Secrets.Add(secretName)
		}
	}

	for _, envVar := range containerSpec.Env {
		logrus.Debugf("env var? %+v\n", envVar)
		if envVar.ValueFrom != nil {
			if envVar.ValueFrom.ConfigMapKeyRef != nil {
				container.ConfigMaps.Add(envVar.ValueFrom.ConfigMapKeyRef.Name)
			} else if envVar.ValueFrom.SecretKeyRef != nil {
				container.Secrets.Add(envVar.ValueFrom.SecretKeyRef.Name)
			}
		}
	}
	for _, envFrom := range containerSpec.EnvFrom {
		logrus.Debugf("env from: %+v\n", envFrom)
		if envFrom.ConfigMapRef != nil {
			container.ConfigMaps.Add(envFrom.ConfigMapRef.Name)
		} else if envFrom.SecretRef != nil {
			container.Secrets.Add(envFrom.SecretRef.Name)
		}
	}

	return container
}

func AnalyzePodSpec(spec v1.PodSpec) *PodSpec {
	var containers []*Container
	configs := map[string]string{}
	secrets := map[string]string{}
	for _, volume := range spec.Volumes {
		if volume.ConfigMap != nil {
			configs[volume.Name] = volume.ConfigMap.LocalObjectReference.Name
		} else if volume.Secret != nil {
			secrets[volume.Name] = volume.Secret.SecretName
		}
	}
	for _, contSpec := range spec.Containers {
		containers = append(containers, analyzeVolumeMounts(false, configs, secrets, contSpec))
		// TODO:
		//contSpec.Env
		//contSpec.EnvFrom
	}
	for _, contSpec := range spec.InitContainers {
		containers = append(containers, analyzeVolumeMounts(true, configs, secrets, contSpec))
	}
	ips := make([]string, len(spec.ImagePullSecrets))
	for i, ref := range spec.ImagePullSecrets {
		ips[i] = ref.Name
	}
	return &PodSpec{
		Containers:       containers,
		ServiceAccount:   spec.ServiceAccountName,
		ImagePullSecrets: ips,
	}
}

func AnalyzeJob(job *batchv1.Job) *PodSpec {
	return AnalyzePodSpec(job.Spec.Template.Spec)
}

func AnalyzeCronJob(job *batchv1.CronJob) *PodSpec {
	return AnalyzePodSpec(job.Spec.JobTemplate.Spec.Template.Spec)
}

func AnalyzeStatefulSet(sset *appsv1.StatefulSet) *PodSpec {
	return AnalyzePodSpec(sset.Spec.Template.Spec)
}

func AnalyzeDeployment(dep *appsv1.Deployment) *PodSpec {
	return AnalyzePodSpec(dep.Spec.Template.Spec)
}
