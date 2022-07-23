package swagger

import (
	"fmt"
	"github.com/mattfenwick/kube-utils/go/pkg/utils"
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"sort"
	"strings"
)

func RunExplain(args *ExplainArgs) {
	swaggerSpec := MustReadSwaggerSpec(MustVersion(args.Version))

	// no types specified?  use them all
	//   otherwise, filter down to just the ones requested
	var typeNames []string
	if len(args.TypeNames) == 0 {
		for name := range swaggerSpec.DefinitionsByNameByGroup() {
			typeNames = append(typeNames, name)
		}
		sort.Strings(typeNames)
	} else {
		typeNames = args.TypeNames // TODO should this be sorted, or respect the input order?
	}

	for _, typeName := range typeNames {
		logrus.Debugf("analysing type %s", typeName)
		analyses := swaggerSpec.AnalyzeType(typeName)

		// no group/versions specified?  use them all
		//   otherwise, filter down to just the ones requested
		if len(args.GroupVersions) > 0 {
			filteredAnalyses := map[string]interface{}{}
			for _, groupVersion := range args.GroupVersions {
				if analysis, ok := analyses[groupVersion]; ok {
					filteredAnalyses[groupVersion] = analysis
				} else {
					logrus.Debugf("type %s not found under group/version %s (%+v)", typeName, groupVersion, utils.SortedKeys(analyses))
				}
			}
			analyses = filteredAnalyses
		}

		gvks := utils.SortedKeys(analyses)
		if len(gvks) == 0 {
			logrus.Debugf("no group/versions found for %s", typeName)
			continue
		}
		for _, groupVersion := range gvks {
			analysis := analyses[groupVersion]
			switch args.Format {
			case "table":
				fmt.Printf("%s.%s:\n%s\n", groupVersion, typeName, ExplainTypeTable(analysis))
			case "condensed":
				fmt.Printf("%s.%s:\n%s\n", groupVersion, typeName, strings.Join(ExplainTypeSummary(analysis), "\n"))
			default:
				panic(errors.Errorf("invalid output format: %s", args.Format))
			}
			fmt.Println()
		}
		fmt.Println()
	}
}

func ExplainTypeTable(o interface{}) string {
	tableString := &strings.Builder{}
	table := tablewriter.NewWriter(tableString)
	table.SetAutoWrapText(false)
	table.SetRowLine(true)
	table.SetAutoMergeCells(true)
	table.SetColMinWidth(1, 100)
	table.SetHeader([]string{"Type", "Field"})
	for _, values := range ResolveExplainType(o, []string{}) {
		table.Append([]string{values[1], values[0]})
	}
	table.Render()
	return tableString.String()
}

func ExplainTypeSummary(obj interface{}) []string {
	var lines []string
	for _, t := range ResolveExplainType(obj, []string{}) {
		chunks := strings.Split(t[0], ".")
		prefix := strings.Repeat("  ", len(chunks)-1)
		typeString := fmt.Sprintf("%s%s", prefix, chunks[len(chunks)-1])
		line := fmt.Sprintf("%-60s    %s", typeString, t[1])
		lines = append(lines, line)
	}
	return lines
}

func ResolveExplainType(obj interface{}, pathContext []string) [][2]string {
	path := make([]string, len(pathContext))
	copy(path, pathContext)

	logrus.Debugf("path: %+v", path)

	var out [][2]string
	switch o := obj.(type) {
	case *Any:
		out = append(out, [2]string{strings.Join(path, "."), "(any)"})
	case *Circular:
		out = append(out, [2]string{strings.Join(path, "."), "(circular)"})
	case *Primitive:
		out = append(out, [2]string{strings.Join(path, "."), o.Type})
	case *Array:
		out = append(out, [2]string{strings.Join(path, "."), "array"})
		out = append(out, ResolveExplainType(o.ElementType, append(path, "[]"))...)
	case *Dict:
		out = append(out, [2]string{strings.Join(path, "."), "map[string]string"})
	case *Object:
		out = append(out, [2]string{strings.Join(path, "."), "object"})
		var sortedFields []string
		for fieldName := range o.Fields {
			sortedFields = append(sortedFields, fieldName)
		}
		sort.Strings(sortedFields)
		for _, fieldName := range sortedFields {
			out = append(out, ResolveExplainType(o.Fields[fieldName], append(path, fieldName))...)
		}
	default:
		panic(errors.Errorf("invalid type: %T", o))
	}
	return out
}
