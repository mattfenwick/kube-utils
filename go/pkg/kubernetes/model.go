package kubernetes

import (
	"fmt"
	"github.com/mattfenwick/kube-utils/go/pkg/graph"
	"github.com/olekukonko/tablewriter"
	"github.com/sirupsen/logrus"
	"sort"
	"strings"
)

type Container struct {
	IsInit     bool
	Name       string
	ConfigMaps map[string]bool
	Secrets    map[string]bool
}

func (c *Container) SecretsSlice() []string {
	var secrets []string
	for secret := range c.Secrets {
		secrets = append(secrets, secret)
	}
	sort.Strings(secrets)
	return secrets
}

func (c *Container) ConfigMapsSlice() []string {
	var cms []string
	for cm := range c.ConfigMaps {
		cms = append(cms, cm)
	}
	sort.Strings(cms)
	return cms
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
	Skipped    map[string][]string
}

func NewModel() *Model {
	return &Model{
		Pods:       map[string]map[string]*PodSpec{},
		Secrets:    nil,
		ConfigMaps: nil,
		Skipped:    map[string][]string{},
	}
}

func (m *Model) AddSkippedResource(kind string, name string) {
	if _, ok := m.Skipped[kind]; !ok {
		m.Skipped[kind] = []string{}
	}
	m.Skipped[kind] = append(m.Skipped[kind], name)
}

func (m *Model) AddPodWrapper(kind string, name string, spec *PodSpec) {
	if _, ok := m.Pods[kind]; !ok {
		m.Pods[kind] = map[string]*PodSpec{}
	}
	m.Pods[kind][name] = spec
}

func (m *Model) SecretConfigMapsUsages() (map[string][]string, map[string][]string) {
	usedSecrets := map[string][]string{}
	usedConfigMaps := map[string][]string{}
	for kind, podSpecs := range m.Pods {
		for resourceName, podSpec := range podSpecs {
			for _, container := range podSpec.Containers {
				for usedSecret := range container.Secrets {
					logrus.Infof("usage of secret %s by %s/%s/%s", usedSecret, kind, resourceName, container.Name)
					usedSecrets[usedSecret] = append(usedSecrets[usedSecret], fmt.Sprintf("%s/%s: %s", kind, resourceName, container.Name))
				}
				for usedConfigMap := range container.ConfigMaps {
					logrus.Infof("usage of configmap %s by %s/%s/%s", usedConfigMap, kind, resourceName, container.Name)
					usedConfigMaps[usedConfigMap] = append(usedConfigMaps[usedConfigMap], fmt.Sprintf("%s/%s: %s", kind, resourceName, container.Name))
				}
			}
		}
	}
	return usedSecrets, usedConfigMaps
}

func (m *Model) SecretUsages(name string) []string {
	secretUsages, _ := m.SecretConfigMapsUsages()
	return secretUsages[name]
}

func (m *Model) ConfigMapUsages(name string) []string {
	_, configMapUsages := m.SecretConfigMapsUsages()
	return configMapUsages[name]
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
				for usedSecret := range container.Secrets {
					usedSecrets[usedSecret] = true
				}
				for usedConfigMap := range container.ConfigMaps {
					usedConfigMaps[usedConfigMap] = true
				}
			}
		}
	}
	return CompareKeySets(createdSecrets, usedSecrets), CompareKeySets(createdConfigMaps, usedConfigMaps)
}

//type Table struct {
//	Title string
//
//}

func (m *Model) Tables() { //[]*Table {
	fmt.Println("\nskipped resources:")
	m.SkippedResourcesTable()

	fmt.Println("secrets:")
	m.SecretsTable()

	fmt.Println("\nconfig maps:")
	m.ConfigMapsTable()

	for kind, resources := range m.Pods {
		fmt.Printf("\nkind: %s\n", kind)
		m.PodsTable(resources)
	}
}

func (m *Model) SkippedResourcesTable() {
	tableString := &strings.Builder{}
	table := tablewriter.NewWriter(tableString)
	table.SetAutoWrapText(false)
	table.SetRowLine(true)
	table.SetAutoMergeCells(true)
	table.SetHeader([]string{"Kind", "Name"})
	for kind, names := range m.Skipped {
		for _, name := range names {
			table.Append([]string{kind, name})
		}
	}
	table.Render()
	fmt.Printf("%s\n", tableString)
}

func (m *Model) PodsTable(resources map[string]*PodSpec) {
	tableString := &strings.Builder{}
	table := tablewriter.NewWriter(tableString)
	table.SetAutoWrapText(false)
	table.SetRowLine(true)
	table.SetAutoMergeCells(true)
	table.SetHeader([]string{"Resource", "Container", "Secrets", "ConfigMaps", "Init"})
	for resourceName, podSpec := range resources {
		for _, container := range podSpec.Containers {
			initString := "init"
			if !container.IsInit {
				initString = ""
			}
			table.Append([]string{resourceName, container.Name, strings.Join(container.SecretsSlice(), "\n"), strings.Join(container.ConfigMapsSlice(), "\n"), initString})
		}
	}
	table.Render()
	fmt.Printf("%s\n", tableString)
}

func (m *Model) SecretsTable() {
	tableString := &strings.Builder{}
	table := tablewriter.NewWriter(tableString)
	table.SetAutoWrapText(false)
	table.SetRowLine(true)
	table.SetAutoMergeCells(true)
	table.SetHeader([]string{"Name", "Source", "Usages"})

	secretsComparison, _ := m.GetUsedUnusedSecretsAndConfigMaps()
	for secret := range secretsComparison.JustA {
		table.Append([]string{secret, "chart", "(none)"})
	}
	for secret := range secretsComparison.Both {
		table.Append([]string{secret, "chart", strings.Join(m.SecretUsages(secret), "\n")})
	}
	for secret := range secretsComparison.JustB {
		table.Append([]string{secret, "unknown", strings.Join(m.SecretUsages(secret), "\n")})
	}
	table.Render()
	fmt.Printf("%s\n", tableString)
}

func (m *Model) ConfigMapsTable() {
	tableString := &strings.Builder{}
	table := tablewriter.NewWriter(tableString)
	table.SetAutoWrapText(false)
	table.SetRowLine(true)
	table.SetAutoMergeCells(true)
	table.SetHeader([]string{"Name", "Source", "Usages"})

	_, configMapsComparison := m.GetUsedUnusedSecretsAndConfigMaps()
	for configMap := range configMapsComparison.JustA {
		table.Append([]string{configMap, "chart", "(none)"})
	}
	for configMap := range configMapsComparison.Both {
		table.Append([]string{configMap, "chart", strings.Join(m.ConfigMapUsages(configMap), "\n")})
	}
	for configMap := range configMapsComparison.JustB {
		table.Append([]string{configMap, "unknown", strings.Join(m.ConfigMapUsages(configMap), "\n")})
	}
	table.Render()
	fmt.Printf("%s\n", tableString)
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
				yamlGraph.AddNode(containerNodeName, fmt.Sprintf(`label="%s%s"`, container.Name, initPiece))
				yamlGraph.AddEdge(resourceName, containerNodeName)
				for cm := range container.ConfigMaps {
					yamlGraph.AddEdge(containerNodeName, "configmap: "+cm)
				}
				for secret := range container.Secrets {
					yamlGraph.AddEdge(containerNodeName, "secret: "+secret)
				}
			}
		}
	}
	return yamlGraph
}
