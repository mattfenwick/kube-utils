package swagger

import (
	"fmt"
	"github.com/mattfenwick/collections/pkg/slice"
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/maps"
	"strings"
)

type ExplainGVKFormat string

const (
	ExplainGVKFormatByResource   = "ExplainGVKFormatByResource"
	ExplainGVKFormatByApiVersion = "ExplainGVKFormatByApiVersion"
)

func ExplainGvks(format ExplainGVKFormat, versions []string) {
	for _, version := range versions {
		kubeVersion := MustVersion(version)
		fmt.Printf("kube version: %s\n", version)
		switch format {
		case ExplainGVKFormatByResource:
			ExplainGvksByResource(kubeVersion)
		case ExplainGVKFormatByApiVersion:
			ExplainGvksByApiVersion(kubeVersion)
		default:
			panic(errors.Errorf("invalid format: %s", format))
		}
	}
}

func ExplainGvksByApiVersion(kubeVersion KubeVersion) {
	spec := MustReadSwaggerSpec(kubeVersion)

	gvksByApiVersion := map[string][]string{}
	for name, def := range spec.Definitions {
		if len(def.XKubernetesGroupVersionKind) > 0 {
			logrus.Debugf("%s, %s, %+v\n", name, def.Type, def.XKubernetesGroupVersionKind)
		}
		for _, gvk := range def.XKubernetesGroupVersionKind {
			apiVersion := ""
			if gvk.Group != "" {
				apiVersion = gvk.Group + "."
			}
			apiVersion += gvk.Version
			gvksByApiVersion[apiVersion] = append(gvksByApiVersion[apiVersion], gvk.Kind)
		}
	}

	fmt.Printf("\n%s\n\n", ExplainGvksResourceTable(gvksByApiVersion, []string{"API version", "Resources"}))
}

func ExplainGvksByResource(kubeVersion KubeVersion) {
	spec := MustReadSwaggerSpec(kubeVersion)

	gvksByResource := map[string][]string{}
	for name, def := range spec.Definitions {
		if len(def.XKubernetesGroupVersionKind) > 0 {
			logrus.Debugf("%s, %s, %+v\n", name, def.Type, def.XKubernetesGroupVersionKind)
		}
		for _, gvk := range def.XKubernetesGroupVersionKind {
			apiVersion := ""
			if gvk.Group != "" {
				apiVersion = gvk.Group + "."
			}
			apiVersion += gvk.Version
			gvksByResource[gvk.Kind] = append(gvksByResource[gvk.Kind], apiVersion)
		}
	}

	fmt.Printf("\n%s\n\n", ExplainGvksResourceTable(gvksByResource, []string{"Resource", "API versions"}))
}

func ExplainGvksResourceTable(rows map[string][]string, headers []string) string {
	tableString := &strings.Builder{}
	table := tablewriter.NewWriter(tableString)
	table.SetAutoWrapText(false)
	table.SetRowLine(true)
	table.SetAutoMergeCells(true)
	table.SetHeader(headers)
	for _, resource := range slice.Sort(maps.Keys(rows)) {
		apiVersions := rows[resource]
		table.Append([]string{resource, strings.Join(slice.Sort(apiVersions), "\n")})
	}
	table.Render()
	return tableString.String()
}
