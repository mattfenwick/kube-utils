package kubernetes

import (
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	goyaml "gopkg.in/yaml.v3"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"
)

func BounceMarshal(in interface{}, out interface{}) error {
	yamlBytes, err := goyaml.Marshal(in)
	if err != nil {
		return errors.Wrapf(err, "unable to marshal yaml")
	}
	err = yaml.UnmarshalStrict(yamlBytes, out)
	return nil
}

func RunAnalyzeExample(path string) {
	objs, err := ParseManyFromFile(path)
	DoOrDie(err)
	model := NewModel()
	for _, o := range objs {
		if o == nil {
			logrus.Infof("skipping nil\n")
			continue
		}
		m := o.(map[string]interface{})
		resourceName := m["metadata"].(map[string]interface{})["name"].(string)
		kind := m["kind"].(string)
		logrus.Infof("kind, name: %s, %s\n", kind, resourceName)
		switch kind {
		case "Deployment":
			var dep *appsv1.Deployment
			DoOrDie(BounceMarshal(o, &dep))
			model.AddPodWrapper("Deployment", dep.Name, AnalyzeDeployment(dep))
		case "StatefulSet":
			var sset *appsv1.StatefulSet
			DoOrDie(BounceMarshal(o, &sset))
			model.AddPodWrapper("StatefulSet", sset.Name, AnalyzeStatefulSet(sset))
		case "Job":
			var job *batchv1.Job
			DoOrDie(BounceMarshal(o, &job))
			model.AddPodWrapper("Job", job.Name, AnalyzeJob(job))
		case "CronJob":
			var job *batchv1.CronJob
			DoOrDie(BounceMarshal(o, &job))
			model.AddPodWrapper("CronJob", job.Name, AnalyzeCronJob(job))
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
		ConfigMaps: map[string]bool{},
		Secrets:    map[string]bool{},
	}
	for _, mount := range containerSpec.VolumeMounts {
		if configMapName, ok := configMaps[mount.Name]; ok {
			container.ConfigMaps[configMapName] = true
		} else if secretName, ok := secrets[mount.Name]; ok {
			container.Secrets[secretName] = true
		}
	}

	for _, envVar := range containerSpec.Env {
		logrus.Debugf("env var? %+v\n", envVar)
		if envVar.ValueFrom != nil {
			if envVar.ValueFrom.ConfigMapKeyRef != nil {
				container.ConfigMaps[envVar.ValueFrom.ConfigMapKeyRef.Name] = true
			} else if envVar.ValueFrom.SecretKeyRef != nil {
				container.Secrets[envVar.ValueFrom.SecretKeyRef.Name] = true
			}
		}
	}
	for _, envFrom := range containerSpec.EnvFrom {
		logrus.Debugf("env from: %+v\n", envFrom)
		if envFrom.ConfigMapRef != nil {
			container.ConfigMaps[envFrom.ConfigMapRef.Name] = true
		} else if envFrom.SecretRef != nil {
			container.Secrets[envFrom.SecretRef.Name] = true
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
