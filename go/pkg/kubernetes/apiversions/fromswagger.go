package apiversions

import (
	"encoding/json"
	"fmt"
	"github.com/mattfenwick/kube-utils/go/pkg/simulator"
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"sort"
	"strings"
)

type SwaggerSpec struct {
	Definitions map[string]struct {
		Description string
		// Properties
		Type                        string
		XKubernetesGroupVersionKind []struct {
			Group   string
			Kind    string
			Version string
		} `json:"x-kubernetes-group-version-kind"`
	}
	Info struct {
		Title   string
		Version string
	}
	//Paths map[string]interface{}
	//Security int
	//SecurityDefinitions int
}

func ParseJsonSpecs() {
	previousTable := &ResourcesTable{
		Version: "???",
		Kinds:   map[string][]string{},
	}
	for _, version := range []string{
		"1.16.15",
		"1.17.17",
		"1.18.19",
		"1.19.11",
		"1.20.7",
		"1.21.2",
		"1.22.4",
		"1.23.0",
	} {
		err := os.MkdirAll("./swagger-data", 0777)
		simulator.DoOrDie(err)

		path := fmt.Sprintf("./swagger-data/%s-swagger-spec.json", version)
		err = GetFileFromURL(BuildSwaggerSpecsURL(version), path)
		simulator.DoOrDie(err)

		in, err := ioutil.ReadFile(path)
		simulator.DoOrDie(errors.Wrapf(err, "unable to read file %s", path))
		obj := &SwaggerSpec{}
		err = json.Unmarshal(in, obj)
		simulator.DoOrDie(errors.Wrapf(err, "unable to unmarshal json"))
		resourcesTable := &ResourcesTable{
			Version: version,
			Kinds:   map[string][]string{},
		}
		for a, b := range obj.Definitions {
			if len(b.XKubernetesGroupVersionKind) > 0 {
				logrus.Debugf("%s, %s, %+v\n", a, b.Type, b.XKubernetesGroupVersionKind)
			}
			for _, gvk := range b.XKubernetesGroupVersionKind {
				apiVersion := ""
				if gvk.Group != "" {
					apiVersion = gvk.Group + "."
				}
				apiVersion += gvk.Version
				resourcesTable.Kinds[gvk.Kind] = append(resourcesTable.Kinds[gvk.Kind], apiVersion)
			}
		}
		//fmt.Printf("simple table:\n%s\n", resourcesTable.SimpleTable())

		fmt.Printf("comparing %s to %s\n", previousTable.Version, resourcesTable.Version)
		resourceDiff := previousTable.Diff(resourcesTable)
		fmt.Printf("added: %+v\n", resourceDiff.Added)
		fmt.Printf("removed: %+v\n", resourceDiff.Removed)
		fmt.Printf("changed:\n%s\n", resourceDiff.Table(Set([]string{
			"Service",
			"ClusterRole",
			"ClusterRoleBinding",
			"ConfigMap",
			"CronJob",
			"CustomResourceDefinition",
			"Deployment",
			"Ingress",
			"Job",
			"Role",
			"RoleBinding",
			"Secret",
			"ServiceAccount",
			"StatefulSet",
		}), Set([]string{"WatchEvent", "DeleteOptions"})))

		previousTable = resourcesTable
	}
}

func BuildSwaggerSpecsURL(kubeVersion string) string {
	return fmt.Sprintf("https://raw.githubusercontent.com/kubernetes/kubernetes/v%s/api/openapi-spec/swagger.json", kubeVersion)
}

func GetFileFromURL(url string, path string) error {
	response, err := http.Get(url)
	if err != nil {
		return errors.Wrapf(err, "unable to GET %s", url)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode > 299 {
		return errors.Errorf("GET request to %s failed with status code %d", url, response.StatusCode)
	}
	bytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return errors.Wrapf(err, "unable to read body from GET to %s", url)
	}

	return errors.Wrapf(ioutil.WriteFile(path, bytes, 0777), "unable to write file %s", path)
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
