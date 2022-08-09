package kubernetes

import (
	"fmt"
	"github.com/mattfenwick/collections/pkg/dict"
	"github.com/mattfenwick/collections/pkg/set"
	"github.com/mattfenwick/collections/pkg/slice"
	"github.com/mattfenwick/kube-utils/pkg/graph"
	"github.com/olekukonko/tablewriter"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/maps"
	"strings"
)

type Container struct {
	IsInit     bool
	Name       string
	ConfigMaps *set.Set[string]
	Secrets    *set.Set[string]
	Image      string
}

func (c *Container) SecretsSlice() []string {
	return slice.Sort(c.Secrets.ToSlice())
}

func (c *Container) ConfigMapsSlice() []string {
	return slice.Sort(c.ConfigMaps.ToSlice())
}

type PodSpec struct {
	Containers       []*Container
	ServiceAccount   string
	ImagePullSecrets []string
	// TODO env vars
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
				for _, usedSecret := range container.Secrets.ToSlice() {
					logrus.Debugf("usage of secret %s by %s/%s/%s", usedSecret, kind, resourceName, container.Name)
					usedSecrets[usedSecret] = append(usedSecrets[usedSecret], fmt.Sprintf("%s/%s: %s", kind, resourceName, container.Name))
				}
				for _, usedConfigMap := range container.ConfigMaps.ToSlice() {
					logrus.Debugf("usage of configmap %s by %s/%s/%s", usedConfigMap, kind, resourceName, container.Name)
					usedConfigMaps[usedConfigMap] = append(usedConfigMaps[usedConfigMap], fmt.Sprintf("%s/%s: %s", kind, resourceName, container.Name))
				}
			}
		}
	}
	return usedSecrets, usedConfigMaps
}

func (m *Model) SecretUsages(name string) []string {
	secretUsages, _ := m.SecretConfigMapsUsages()
	return slice.Sort(secretUsages[name])
}

func (m *Model) ConfigMapUsages(name string) []string {
	_, configMapUsages := m.SecretConfigMapsUsages()
	return slice.Sort(configMapUsages[name])
}

func (m *Model) GetUsedUnusedSecretsAndConfigMaps() (*KeySetComparison, *KeySetComparison) {
	usedSecrets := set.FromSlice[string](nil)
	usedConfigMaps := set.FromSlice[string](nil)
	for _, podSpecs := range m.Pods {
		for _, podSpec := range podSpecs {
			for _, container := range podSpec.Containers {
				for _, usedSecret := range container.Secrets.ToSlice() {
					usedSecrets.Add(usedSecret)
				}
				for _, usedConfigMap := range container.ConfigMaps.ToSlice() {
					usedConfigMaps.Add(usedConfigMap)
				}
			}
		}
	}
	return CompareKeySets(set.FromSlice(m.Secrets), usedSecrets), CompareKeySets(set.FromSlice(m.ConfigMaps), usedConfigMaps)
}

func (m *Model) GetImageUsages() map[string][]string {
	usages := map[string][]string{}

	for kind, podSpecs := range m.Pods {
		for resourceName, podSpec := range podSpecs {
			for _, container := range podSpec.Containers {
				usages[container.Image] = append(usages[container.Image], fmt.Sprintf("%s/%s: %s", kind, resourceName, container.Name))
			}
		}
	}
	return usages
}

func (m *Model) BuildTables() (string, string, string, string, map[string]string) {
	makePodsTable := func(a map[string]*PodSpec) string { return m.PodsTable(a) }
	return m.SkippedResourcesTable(), m.SecretsTable(), m.ConfigMapsTable(), m.ImagesTable(), dict.Map(makePodsTable, m.Pods)
}

func (m *Model) SkippedResourcesTable() string {
	tableString := &strings.Builder{}
	table := tablewriter.NewWriter(tableString)
	table.SetAutoWrapText(false)
	table.SetRowLine(true)
	table.SetAutoMergeCells(true)
	table.SetHeader([]string{"Kind", "Name"})
	for _, kind := range slice.Sort(maps.Keys(m.Skipped)) {
		names := m.Skipped[kind]
		for _, name := range slice.Sort(names) {
			table.Append([]string{kind, name})
		}
	}
	table.Render()
	return tableString.String()
}

func (m *Model) PodsTable(resources map[string]*PodSpec) string {
	tableString := &strings.Builder{}
	table := tablewriter.NewWriter(tableString)
	table.SetAutoWrapText(false)
	table.SetRowLine(true)
	table.SetAutoMergeCells(true)
	table.SetHeader([]string{"Resource", "Container", "Secrets", "ConfigMaps", "Init"})
	for _, resourceName := range slice.Sort(maps.Keys(resources)) {
		podSpec := resources[resourceName]
		for _, container := range slice.SortOn(func(c *Container) string { return c.Name }, podSpec.Containers) {
			initString := "Y"
			if !container.IsInit {
				initString = "N"
			}
			table.Append([]string{resourceName, container.Name, strings.Join(container.SecretsSlice(), "\n"), strings.Join(container.ConfigMapsSlice(), "\n"), initString})
		}
	}
	table.Render()
	return tableString.String()
}

func (m *Model) SecretsTable() string {
	tableString := &strings.Builder{}
	table := tablewriter.NewWriter(tableString)
	table.SetAutoWrapText(false)
	table.SetRowLine(true)
	table.SetAutoMergeCells(true)
	table.SetHeader([]string{"Name", "Source", "Usages"})

	secretsComparison, _ := m.GetUsedUnusedSecretsAndConfigMaps()
	for _, secret := range secretsComparison.JustA {
		table.Append([]string{secret, "chart", "(none)"})
	}
	for _, secret := range secretsComparison.Both {
		table.Append([]string{secret, "chart", strings.Join(m.SecretUsages(secret), "\n")})
	}
	for _, secret := range secretsComparison.JustB {
		table.Append([]string{secret, "unknown", strings.Join(m.SecretUsages(secret), "\n")})
	}
	table.Render()
	return tableString.String()
}

func (m *Model) ConfigMapsTable() string {
	tableString := &strings.Builder{}
	table := tablewriter.NewWriter(tableString)
	table.SetAutoWrapText(false)
	table.SetRowLine(true)
	table.SetAutoMergeCells(true)
	table.SetHeader([]string{"Name", "Source", "Usages"})

	_, configMapsComparison := m.GetUsedUnusedSecretsAndConfigMaps()
	for _, configMap := range configMapsComparison.JustA {
		table.Append([]string{configMap, "chart", "(none)"})
	}
	for _, configMap := range configMapsComparison.Both {
		table.Append([]string{configMap, "chart", strings.Join(m.ConfigMapUsages(configMap), "\n")})
	}
	for _, configMap := range configMapsComparison.JustB {
		table.Append([]string{configMap, "unknown", strings.Join(m.ConfigMapUsages(configMap), "\n")})
	}
	table.Render()
	return tableString.String()
}

func (m *Model) ImagesTable() string {
	tableString := &strings.Builder{}
	table := tablewriter.NewWriter(tableString)
	table.SetAutoWrapText(false)
	table.SetRowLine(true)
	table.SetAutoMergeCells(true)
	table.SetHeader([]string{"Image", "Source"})

	imageUsages := m.GetImageUsages()
	for _, image := range slice.Sort(maps.Keys(imageUsages)) {
		usages := slice.Sort(imageUsages[image])
		table.Append([]string{image, strings.Join(usages, "\n")})
	}

	table.Render()
	return tableString.String()
}

func (m *Model) Graph() *graph.Graph {
	yamlGraph := graph.NewGraph("", "")

	secretsComparison, configMapsComparison := m.GetUsedUnusedSecretsAndConfigMaps()

	secretsGraph := graph.NewGraph("secrets", "secrets")
	unusedSecretsGraph := graph.NewGraph("unused secrets", "unused secrets")
	unknownSourceSecretsGraph := graph.NewGraph("unknown source secrets", "unknown source secrets")
	for _, secret := range secretsComparison.JustA {
		unusedSecretsGraph.AddNode("secret: "+secret, fmt.Sprintf(`label="%s"`, secret))
	}
	for _, secret := range secretsComparison.Both {
		secretsGraph.AddNode("secret: "+secret, fmt.Sprintf(`label="%s"`, secret))
	}
	for _, secret := range secretsComparison.JustB {
		unknownSourceSecretsGraph.AddNode("secret: "+secret, fmt.Sprintf(`label="%s"`, secret))
	}
	yamlGraph.AddSubgraph(secretsGraph)
	yamlGraph.AddSubgraph(unusedSecretsGraph)
	yamlGraph.AddSubgraph(unknownSourceSecretsGraph)

	cmsGraph := graph.NewGraph("configmaps", "configmaps")
	unusedConfigMapsGraph := graph.NewGraph("unused configmaps", "unused configmaps")
	unknownSourceConfigMapsGraph := graph.NewGraph("unknown source configmaps", "unknown source configmaps")
	for _, secret := range configMapsComparison.JustA {
		unusedConfigMapsGraph.AddNode("configmap: "+secret, fmt.Sprintf(`label="%s"`, secret))
	}
	for _, secret := range configMapsComparison.Both {
		cmsGraph.AddNode("configmap: "+secret, fmt.Sprintf(`label="%s"`, secret))
	}
	for _, secret := range configMapsComparison.JustB {
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
				for _, cm := range container.ConfigMaps.ToSlice() {
					yamlGraph.AddEdge(containerNodeName, "configmap: "+cm)
				}
				for _, secret := range container.Secrets.ToSlice() {
					yamlGraph.AddEdge(containerNodeName, "secret: "+secret)
				}
			}
		}
	}
	return yamlGraph
}
