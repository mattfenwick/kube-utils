package kubernetes

import (
	"github.com/mattfenwick/collections/pkg/dict"
	"github.com/mattfenwick/collections/pkg/slice"
	"github.com/olekukonko/tablewriter"
	"golang.org/x/exp/maps"
	"strings"
)

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
