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

// ExplainGVKTable models a pivot table
type ExplainGVKTable struct {
	FirstColumnHeader string
	Rows              map[string]map[string][]string
	Columns           []string
}

func (e *ExplainGVKTable) Add(rowKey string, columnKey string, value string) {
	if _, ok := e.Rows[rowKey]; !ok {
		e.Rows[rowKey] = map[string][]string{}
	}
	e.Rows[rowKey][columnKey] = append(e.Rows[rowKey][columnKey], value)
}

func (e *ExplainGVKTable) FormattedTable() string {
	tableString := &strings.Builder{}
	table := tablewriter.NewWriter(tableString)
	table.SetAutoWrapText(false)
	table.SetRowLine(true)
	table.SetAutoMergeCells(true)
	table.SetHeader(append([]string{e.FirstColumnHeader}, e.Columns...))
	for _, rowKey := range slice.Sort(maps.Keys(e.Rows)) {
		row := []string{rowKey}
		for _, columnKey := range e.Columns {
			row = append(row, strings.Join(slice.Sort(e.Rows[rowKey][columnKey]), "\n"))
		}
		table.Append(row)
	}
	table.Render()
	return tableString.String()
}

type ExplainGVKGroupBy string

const (
	ExplainGVKGroupByResource   = "ExplainGVKGroupByResource"
	ExplainGVKGroupByApiVersion = "ExplainGVKGroupByApiVersion"
)

func (e ExplainGVKGroupBy) Header() string {
	switch e {
	case ExplainGVKGroupByResource:
		return "Resource"
	case ExplainGVKGroupByApiVersion:
		return "API version"
	default:
		panic(errors.Errorf("invalid groupBy: %s", e))
	}
}

func ExplainGvks(groupBy ExplainGVKGroupBy, versions []string, include func(string, string) bool) {
	table := &ExplainGVKTable{
		FirstColumnHeader: groupBy.Header(),
		Rows:              map[string]map[string][]string{},
		Columns:           versions,
	}
	for _, version := range versions {
		kubeVersion := MustVersion(version)
		logrus.Debugf("kube version: %s", version)

		spec := MustReadSwaggerSpec(kubeVersion)
		for name, def := range spec.Definitions {
			if len(def.XKubernetesGroupVersionKind) > 0 {
				logrus.Debugf("%s, %s, %+v\n", name, def.Type, def.XKubernetesGroupVersionKind)
			}
			for _, gvk := range def.XKubernetesGroupVersionKind {
				apiVersion := gvk.ApiVersion()
				if include(apiVersion, gvk.Kind) {
					logrus.Debugf("adding gvk: %s, %s", apiVersion, gvk.Kind)
					switch groupBy {
					case ExplainGVKGroupByResource:
						table.Add(gvk.Kind, kubeVersion.ToString(), apiVersion)
					case ExplainGVKGroupByApiVersion:
						table.Add(apiVersion, kubeVersion.ToString(), gvk.Kind)
					default:
						panic(errors.Errorf("invalid groupBy: %s", groupBy))
					}
				} else {
					logrus.Debugf("skipping gvk: %s, %s", apiVersion, gvk.Kind)
				}
			}
		}
	}

	fmt.Printf("\n%s\n\n", table.FormattedTable())
}
