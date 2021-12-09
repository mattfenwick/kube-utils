package apiversions

import (
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"reflect"
	"sort"
	"strings"
)

type MapDiff struct {
	Added   []string
	Removed []string
	Same    []string
}

type ResourceDiff struct {
	Added   []string
	Removed []string
	Changed map[string]*MapDiff
}

func (r *ResourceDiff) SortedChangedKeys() []string {
	var keys []string
	for key := range r.Changed {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
	return keys
}

func (r *ResourceDiff) Table(includes map[string]bool, skips map[string]bool) string {
	tableString := &strings.Builder{}
	table := tablewriter.NewWriter(tableString)
	table.SetAutoWrapText(false)
	table.SetRowLine(true)
	table.SetAutoMergeCells(true)
	table.SetHeader([]string{"Kind", "Added", "Removed", "Same"})
	for _, kind := range r.SortedChangedKeys() {
		change := r.Changed[kind]
		if (len(includes) == 0 || includes[kind]) && !skips[kind] && (len(change.Added) != 0 || len(change.Removed) != 0) {
			table.Append([]string{
				kind,
				strings.Join(change.Added, "\n"),
				strings.Join(change.Removed, "\n"),
				strings.Join(change.Same, "\n"),
			})
			logrus.Debugf("kind %s; added: %+v, removed: %+v, same: %+v\n", kind, change.Added, change.Removed, change.Same)
		}
	}
	table.Render()
	return tableString.String()
}

type ResourcesTable struct {
	Version string
	Kinds   map[string][]string
	Headers []string
	Rows    [][]string
}

func (r *ResourcesTable) SortedKinds() []string {
	var kinds []string
	for kind := range r.Kinds {
		kinds = append(kinds, kind)
	}
	sort.Slice(kinds, func(i, j int) bool {
		return kinds[i] < kinds[j]
	})
	return kinds
}

func (r *ResourcesTable) SimpleTable() string {
	tableString := &strings.Builder{}
	table := tablewriter.NewWriter(tableString)
	table.SetAutoWrapText(false)
	table.SetRowLine(true)
	table.SetAutoMergeCells(true)
	table.SetHeader([]string{"Kind", "API Version"})
	for _, kind := range r.SortedKinds() {
		versions := r.Kinds[kind]
		sort.Slice(versions, func(i, j int) bool {
			return versions[i] < versions[j]
		})
		for _, apiVersion := range versions {
			table.Append([]string{kind, apiVersion})
		}
	}
	table.Render()
	return tableString.String()
}

func NewResourcesTable(version string, headers []string, rows [][]string) (*ResourcesTable, error) {
	if !reflect.DeepEqual(headers, []string{"NAME", "SHORTNAMES", "APIVERSION", "NAMESPACED", "KIND", "VERBS"}) {
		return nil, errors.Errorf("invalid headers: %+v", headers)
	}
	table := &ResourcesTable{Version: version, Kinds: map[string][]string{}, Headers: headers, Rows: rows}
	for _, row := range rows {
		table.Kinds[row[4]] = append(table.Kinds[row[4]], row[2])
	}
	return table, nil
}

func (r *ResourcesTable) Diff(other *ResourcesTable) *ResourceDiff {
	var added, removed []string
	changed := map[string]*MapDiff{}
	for ak, av := range r.Kinds {
		bv, ok := other.Kinds[ak]
		if ok {
			changed[ak] = SliceDiff(av, bv)
		} else {
			removed = append(removed, ak)
		}
	}
	for bk := range other.Kinds {
		if _, ok := r.Kinds[bk]; !ok {
			added = append(added, bk)
		}
	}
	return &ResourceDiff{
		Added:   added,
		Removed: removed,
		Changed: changed,
	}
}

func (r *ResourcesTable) KindResourcesTable() string {
	tableString := &strings.Builder{}
	table := tablewriter.NewWriter(tableString)
	table.SetAutoWrapText(false)
	table.SetRowLine(true)
	table.SetAutoMergeCells(true)
	table.SetHeader(r.Headers)
	table.AppendBulk(r.Rows)
	table.Render()
	return tableString.String()
}

func SliceDiff(as []string, bs []string) *MapDiff {
	aSet, bSet := Set(as), Set(bs)
	var added, removed, same []string
	for key := range aSet {
		if bSet[key] {
			same = append(same, key)
		} else {
			removed = append(removed, key)
		}
	}
	for key := range bSet {
		if !aSet[key] {
			added = append(added, key)
		}
	}
	return &MapDiff{
		Added:   added,
		Removed: removed,
		Same:    same,
	}
}
