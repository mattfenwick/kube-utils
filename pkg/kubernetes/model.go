package kubernetes

import (
	"fmt"
	"github.com/mattfenwick/kube-utils/pkg/graph"
)

type Container struct {
	IsInit     bool
	Name       string
	ConfigMaps []string
	Secrets    []string
}

type PodSpec struct {
	Containers       []*Container
	ServiceAccount   string
	ImagePullSecrets []string
	// TODO env vars
	// TODO image
}

type Model struct {
	Pods       map[string]map[string]*PodSpec
	Secrets    []string
	ConfigMaps []string
}

func (m *Model) AddPodWrapper(kind string, name string, spec *PodSpec) {
	if m.Pods == nil {
		m.Pods = map[string]map[string]*PodSpec{}
	}
	if _, ok := m.Pods[kind]; !ok {
		m.Pods[kind] = map[string]*PodSpec{}
	}
	m.Pods[kind][name] = spec
}

func (m *Model) GetUsedUnusedSecretsAndConfigMaps() (*KeySetComparison, *KeySetComparison) {
	createdSecrets := map[string]bool{}
	for _, s := range m.Secrets {
		createdSecrets[s] = true
	}
	createdConfigMaps := map[string]bool{}
	for _, cm := range m.ConfigMaps {
		createdConfigMaps[cm] = true
	}

	usedSecrets := map[string]bool{}
	usedConfigMaps := map[string]bool{}
	for _, podSpecs := range m.Pods {
		for _, podSpec := range podSpecs {
			for _, container := range podSpec.Containers {
				for _, usedSecret := range container.Secrets {
					usedSecrets[usedSecret] = true
				}
				for _, usedConfigMap := range container.ConfigMaps {
					usedConfigMaps[usedConfigMap] = true
				}
			}
		}
	}
	return CompareKeySets(createdSecrets, usedSecrets), CompareKeySets(createdConfigMaps, usedConfigMaps)
}

func (m *Model) Graph() *graph.Graph {
	yamlGraph := graph.NewGraph("", "")

	secretsComparison, configMapsComparison := m.GetUsedUnusedSecretsAndConfigMaps()

	secretsGraph := graph.NewGraph("secrets", "secrets")
	unusedSecretsGraph := graph.NewGraph("unused secrets", "unused secrets")
	unknownSourceSecretsGraph := graph.NewGraph("unknown source secrets", "unknown source secrets")
	for secret := range secretsComparison.JustA {
		unusedSecretsGraph.AddNode("secret: "+secret, fmt.Sprintf(`label="%s"`, secret))
	}
	for secret := range secretsComparison.Both {
		secretsGraph.AddNode("secret: "+secret, fmt.Sprintf(`label="%s"`, secret))
	}
	for secret := range secretsComparison.JustB {
		unknownSourceSecretsGraph.AddNode("secret: "+secret, fmt.Sprintf(`label="%s"`, secret))
	}
	yamlGraph.AddSubgraph(secretsGraph)
	yamlGraph.AddSubgraph(unusedSecretsGraph)
	yamlGraph.AddSubgraph(unknownSourceSecretsGraph)

	cmsGraph := graph.NewGraph("configmaps", "configmaps")
	unusedConfigMapsGraph := graph.NewGraph("unused configmaps", "unused configmaps")
	unknownSourceConfigMapsGraph := graph.NewGraph("unknown source configmaps", "unknown source configmaps")
	for secret := range configMapsComparison.JustA {
		unusedConfigMapsGraph.AddNode("configmap: "+secret, fmt.Sprintf(`label="%s"`, secret))
	}
	for secret := range configMapsComparison.Both {
		cmsGraph.AddNode("configmap: "+secret, fmt.Sprintf(`label="%s"`, secret))
	}
	for secret := range configMapsComparison.JustB {
		unknownSourceConfigMapsGraph.AddNode("configmap: "+secret, fmt.Sprintf(`label="%s"`, secret))
	}
	yamlGraph.AddSubgraph(cmsGraph)
	yamlGraph.AddSubgraph(unusedConfigMapsGraph)
	yamlGraph.AddSubgraph(unknownSourceConfigMapsGraph)

	for kind, objects := range m.Pods {
		for name, spec := range objects {
			resourceName := fmt.Sprintf("%s: %s", kind, name)
			yamlGraph.AddNode(resourceName)
			// TODO:
			//spec.ServiceAccount
			//spec.ImagePullSecrets
			for _, container := range spec.Containers {
				initPiece := ""
				if container.IsInit {
					initPiece = " (init)"
				}
				containerNodeName := fmt.Sprintf("%s: %s/%s%s", kind, name, container.Name, initPiece)
				yamlGraph.AddNode(containerNodeName)
				yamlGraph.AddEdge(resourceName, containerNodeName)
				for _, cm := range container.ConfigMaps {
					yamlGraph.AddEdge(containerNodeName, "configmap: "+cm)
				}
				for _, secret := range container.Secrets {
					yamlGraph.AddEdge(containerNodeName, "secret: "+secret)
				}
			}
		}
	}
	return yamlGraph
}
