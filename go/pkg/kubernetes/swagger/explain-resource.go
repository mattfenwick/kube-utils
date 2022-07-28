package swagger

import (
	"fmt"
	"github.com/mattfenwick/collections/pkg/slice"
	"github.com/mattfenwick/kube-utils/go/pkg/utils"
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/maps"
	"strings"
)

func RunExplainResource(args *ExplainResourceArgs) {
	if args.Hack {
		HackExplainResource(args)
		return
	}

	swaggerSpec := MustReadSwaggerSpecFromGithub(MustVersion(args.Version))

	// no types specified?  use them all
	//   otherwise, filter down to just the ones requested
	var typeNames []string
	if len(args.TypeNames) == 0 {
		typeNames = slice.Sort(maps.Keys(swaggerSpec.DefinitionsByNameByGroup()))
	} else {
		typeNames = args.TypeNames // TODO should this be sorted, or respect the input order?
	}

	for _, typeName := range typeNames {
		logrus.Debugf("analysing type %s", typeName)
		resources := ResolveResource(swaggerSpec, typeName)

		// no group/versions specified?  use them all
		//   otherwise, filter down to just the ones requested
		if len(args.GroupVersions) > 0 {
			filteredResources := map[string]interface{}{}
			for _, groupVersion := range args.GroupVersions {
				if analysis, ok := resources[groupVersion]; ok {
					filteredResources[groupVersion] = analysis
				} else {
					logrus.Debugf("type %s not found under group/version %s (%+v)", typeName, groupVersion, utils.SortedKeys(resources))
				}
			}
			resources = filteredResources
		}

		gvks := utils.SortedKeys(resources)
		if len(gvks) == 0 {
			logrus.Debugf("no group/versions found for %s", typeName)
			continue
		}
		for _, groupVersion := range gvks {
			resource := resources[groupVersion]
			switch args.Format {
			case "table":
				fmt.Printf("%s.%s:\n%s\n", groupVersion, typeName, ExplainResourceTable(resource))
			case "condensed":
				fmt.Printf("%s.%s:\n%s\n", groupVersion, typeName, strings.Join(ExplainResourceSummary(resource), "\n"))
			default:
				panic(errors.Errorf("invalid output format: %s", args.Format))
			}
			fmt.Println()
		}
		fmt.Println()
	}
}

func ExplainResourceTable(o interface{}) string {
	tableString := &strings.Builder{}
	table := tablewriter.NewWriter(tableString)
	table.SetAutoWrapText(false)
	table.SetRowLine(true)
	table.SetAutoMergeCells(true)
	table.SetColMinWidth(1, 100)
	table.SetHeader([]string{"Type", "Field"})
	for _, values := range ResolveExplainResource(o, []string{}) {
		table.Append([]string{values[1], values[0]})
	}
	table.Render()
	return tableString.String()
}

func ExplainResourceSummary(obj interface{}) []string {
	var lines []string
	for _, t := range ResolveExplainResource(obj, []string{}) {
		chunks := strings.Split(t[0], ".")
		prefix := strings.Repeat("  ", len(chunks)-1)
		typeString := fmt.Sprintf("%s%s", prefix, chunks[len(chunks)-1])
		line := fmt.Sprintf("%-60s    %s", typeString, t[1])
		lines = append(lines, line)
	}
	return lines
}

func ResolveExplainResource(obj interface{}, pathContext []string) [][2]string {
	path := make([]string, len(pathContext))
	copy(path, pathContext)

	logrus.Debugf("path: %+v", path)

	var out [][2]string
	switch o := obj.(type) {
	case *Circular:
		out = append(out, [2]string{strings.Join(path, "."), "(circular)"})
	case *Primitive:
		out = append(out, [2]string{strings.Join(path, "."), o.Type})
	case *Array:
		out = append(out, [2]string{strings.Join(path, "."), "array"})
		out = append(out, ResolveExplainResource(o.ElementType, append(path, "[]"))...)
	case *Dict:
		out = append(out, [2]string{strings.Join(path, "."), "map[string]string"})
	case *Object:
		out = append(out, [2]string{strings.Join(path, "."), "object"})
		for _, fieldName := range slice.Sort(maps.Keys(o.Fields)) {
			out = append(out, ResolveExplainResource(o.Fields[fieldName], append(path, fieldName))...)
		}
	default:
		panic(errors.Errorf("invalid type: %T", o))
	}
	return out
}
