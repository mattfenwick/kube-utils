package main

import (
	"fmt"
	"github.com/mattfenwick/kube-utils/go/pkg/simulator"
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"reflect"
	"regexp"
	"strings"
)

func main() {
	previousTable := &ResourcesTable{
		Version: "",
		Kinds:   map[string][]string{},
		Headers: nil,
		Rows:    nil,
	}
	for _, version := range []string{
		"1.18.19",
		"1.19.11",
		"1.20.7",
		"1.21.2",
		"1.22.4",
		"1.23.0",
	} {
		headers, rows, err := ReadCSV(fmt.Sprintf("../kube/data/v%s-api-resources.txt", version))
		simulator.DoOrDie(err)
		rsTable, err := NewResourcesTable(version, headers, rows)
		simulator.DoOrDie(err)

		fmt.Printf("%s\n", rsTable.Table())

		//for kind, apiVersions := range rsTable.Kinds {
		//	fmt.Printf("%s, %+v\n", kind, apiVersions)
		//}

		fmt.Printf("comparing %s to %s\n", previousTable.Version, rsTable.Version)
		resourceDiff := previousTable.Diff(rsTable)
		fmt.Printf("added: %+v\n", resourceDiff.Added)
		fmt.Printf("removed: %+v\n", resourceDiff.Removed)
		fmt.Printf("changed:\n%s\n", resourceDiff.Table())
		//for kind, change := range resourceDiff.Changed {
		//	if len(change.Added) != 0 || len(change.Removed) != 0 {
		//		fmt.Printf("kind %s; added: %+v, removed: %+v, same: %+v\n", kind, change.Added, change.Removed, change.Same)
		//	}
		//}

		previousTable = rsTable
	}
}

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

func (r *ResourceDiff) Table() string {
	tableString := &strings.Builder{}
	table := tablewriter.NewWriter(tableString)
	table.SetAutoWrapText(false)
	table.SetRowLine(true)
	table.SetAutoMergeCells(true)
	table.SetHeader([]string{"Kind", "Added", "Removed", "Same"})
	for kind, change := range r.Changed {
		if len(change.Added) != 0 || len(change.Removed) != 0 {
			table.Append([]string{
				kind,
				strings.Join(change.Added, "\n"),
				strings.Join(change.Removed, "\n"),
				strings.Join(change.Same, "\n"),
			})
			fmt.Printf("kind %s; added: %+v, removed: %+v, same: %+v\n", kind, change.Added, change.Removed, change.Same)
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

func Set(xs []string) map[string]bool {
	out := map[string]bool{}
	for _, x := range xs {
		out[x] = true
	}
	return out
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

func (r *ResourcesTable) Table() string {
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

func ReadCSV(path string) ([]string, [][]string, error) {
	in, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "unable to read file %s", path)
	}
	lines := strings.Split(string(in), "\n")
	headings, startIndexes := FindFieldStarts(lines[0])
	logrus.Debugf("headings: %+v\n%+v\n", headings, startIndexes)
	var rows [][]string
	for ix, line := range lines[1:] {
		logrus.Debugf("line: <%s>\n", line)
		if ix == len(lines)-2 && line == "" {
			break
		}
		var fields []string
		for i, start := range startIndexes {
			var stop int
			if i == len(startIndexes)-1 {
				stop = len(line)
			} else {
				stop = startIndexes[i+1]
			}
			trimmed := strings.TrimRight(line[start:stop], " ")
			fields = append(fields, trimmed)
			logrus.Debugf("trimmed? %+v\n", trimmed)
		}
		rows = append(rows, fields)
	}
	return headings, rows, nil
}

func FindFieldStarts(line string) ([]string, []int) {
	regex := regexp.MustCompile(`\S+`)
	nums := regex.FindAllStringIndex(line, -1)
	var headings []string
	var startIndexes []int
	for _, ns := range nums {
		logrus.Debugf("%+v\n", ns)
		start, stop := ns[0], ns[1]
		headings = append(headings, line[start:stop])
		startIndexes = append(startIndexes, start)
	}
	return headings, startIndexes
}
