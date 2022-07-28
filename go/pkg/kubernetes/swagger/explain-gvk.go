package swagger

import (
	"fmt"
	"github.com/mattfenwick/collections/pkg/slice"
	"github.com/mattfenwick/kube-utils/go/pkg/kubernetes/swagger/apiversions"
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

func formatCell(items []string) string {
	return strings.Join(slice.Sort(items), "\n")
}

func (e *ExplainGVKTable) FormattedTable(calculateDiff bool) string {
	tableString := &strings.Builder{}
	table := tablewriter.NewWriter(tableString)
	table.SetAutoWrapText(false)
	table.SetRowLine(true)
	table.SetAutoMergeCells(true)
	table.SetHeader(append([]string{e.FirstColumnHeader}, e.Columns...))
	if calculateDiff {
		if len(e.Columns) == 0 {
			panic(errors.Errorf("unable to calculate diff for 0 versions"))
		}
		for _, rowKey := range slice.Sort(maps.Keys(e.Rows)) {
			prev := e.Rows[rowKey][e.Columns[0]]
			row := []string{rowKey, formatCell(prev)}

			for _, columnKey := range e.Columns[1:] {
				curr := e.Rows[rowKey][columnKey]

				diff := apiversions.SliceDiff(prev, curr)
				diff.Sort()

				var add, remove string
				if len(diff.Added) > 0 {
					add = fmt.Sprintf("add:\n  %s\n\n", strings.Join(slice.Sort(diff.Added), "\n  "))
				}
				if len(diff.Removed) > 0 {
					add = fmt.Sprintf("remove:\n  %s\n\n", strings.Join(slice.Sort(diff.Removed), "\n  "))
				}
				row = append(row, fmt.Sprintf("%s%s", add, remove))

				prev = curr
			}
			table.Append(row)
		}
	} else {
		for _, rowKey := range slice.Sort(maps.Keys(e.Rows)) {
			row := []string{rowKey}
			for _, columnKey := range e.Columns {
				row = append(row, formatCell(e.Rows[rowKey][columnKey]))
			}
			table.Append(row)
		}
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

func ExplainGvks(groupBy ExplainGVKGroupBy, versions []string, include func(string, string) bool, calculateDiff bool) string {
	table := &ExplainGVKTable{
		FirstColumnHeader: groupBy.Header(),
		Rows:              map[string]map[string][]string{},
		Columns:           versions,
	}
	for _, version := range versions {
		kubeVersion := MustVersion(version)
		logrus.Debugf("kube version: %s", version)

		spec := MustReadSwaggerSpecFromGithub(kubeVersion)
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

	return table.FormattedTable(calculateDiff)
}
