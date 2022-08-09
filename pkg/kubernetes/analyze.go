package kubernetes

import (
	"github.com/mattfenwick/collections/pkg/set"
	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
)

func getResourceName(o map[string]interface{}) string {
	return o["metadata"].(map[string]interface{})["name"].(string)
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
