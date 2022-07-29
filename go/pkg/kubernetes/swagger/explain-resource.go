package swagger

import (
	"fmt"
	"github.com/mattfenwick/collections/pkg/set"
	"github.com/mattfenwick/collections/pkg/slice"
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/maps"
	"strings"
)

func RunExplainResource(args *ExplainResourceArgs) {
	spec := MustReadSwaggerSpecFromGithub(MustVersion(args.Version))
	gvks := spec.ResolveStructure()

	resources := set.NewSet(args.TypeNames)
	fmt.Printf("\n\n\n\n")
	for _, name := range slice.Sort(maps.Keys(gvks)) {
		if len(args.TypeNames) > 0 && !resources.Contains(name) {
			continue
		}
		switch args.Format {
		case "debug":
			fmt.Printf("%s:\n", name)
			for gv, kind := range gvks[name] {
				fmt.Printf("gv: %s:\n", gv)
				for _, path := range kind.Paths([]string{name}) {
					if args.Depth == 0 || len(path.Fst) <= args.Depth {
						fmt.Printf("  %+v: %s\n", path.Fst, path.Snd)
					}
				}
			}
			//json.Print(resolved[name])
			fmt.Printf("\n\n")
		case "table":
			for gv, kind := range gvks[name] {
				fmt.Printf("%s %s:\n", gv, name)
				fmt.Printf("%s\n\n", TableResource(kind, args.Depth))
			}
		case "condensed":
			fmt.Printf("%s:\n", name)
			for gv, kind := range gvks[name] {
				fmt.Printf("%s\n\n", CondensedResource(gv, name, kind, args.Depth))
			}
		default:
			panic(errors.Errorf("invalid output format: %s", args.Format))
		}
	}
}

func TableResource(kind *ResolvedType, depth int) string {
	tableString := &strings.Builder{}
	table := tablewriter.NewWriter(tableString)
	table.SetAutoWrapText(false)
	table.SetRowLine(true)
	table.SetAutoMergeCells(true)
	table.SetColMinWidth(1, 100)
	table.SetHeader([]string{"Type", "Field"})
	for _, values := range ExplainResource(kind, []string{}, 0, depth) {
		table.Append([]string{values[1], values[0]})
	}
	table.Render()
	return tableString.String()
}

func ExplainResource(obj *ResolvedType, pathContext []string, depth int, maxDepth int) [][2]string {
	logrus.Debugf("path: %+v", pathContext)
	if maxDepth > 0 && depth >= maxDepth {
		return nil
	}

	path := make([]string, len(pathContext))
	copy(path, pathContext)

	var out [][2]string
	if obj.Circular != "" {
		out = append(out, [2]string{strings.Join(path, "."), obj.Circular})
	} else if obj.Primitive != "" {
		out = append(out, [2]string{strings.Join(path, "."), obj.Primitive})
	} else if obj.Array != nil {
		out = append(out, [2]string{strings.Join(path, "."), "array"})
		out = append(out, ExplainResource(obj.Array, append(path, "[]"), depth+1, maxDepth)...)
	} else if obj.Object != nil {
		out = append(out, [2]string{strings.Join(path, "."), "object"})
		for _, fieldName := range slice.Sort(maps.Keys(obj.Object.Properties)) {
			out = append(out, ExplainResource(obj.Object.Properties[fieldName], append(path, fieldName), depth+1, maxDepth)...)
		}
		if obj.Object.AdditionalProperties != nil {
			out = append(out, ExplainResource(obj.Object.AdditionalProperties, append(path, "additionalProperties"), depth+1, maxDepth)...)
		}
	} else if obj.Empty {

	} else {
		panic(errors.Errorf("invalid ResolvedType: %+v", obj))
	}
	return out
}

func CondensedResource(gv string, name string, kind *ResolvedType, depth int) string {
	lines := []string{gv + ":"}
	for _, path := range kind.Paths([]string{name}) {
		if depth == 0 || len(path.Fst) <= depth {
			prefix := strings.Repeat("  ", len(path.Fst)-1)
			typeString := fmt.Sprintf("%s%s", prefix, path.Fst[len(path.Fst)-1])
			lines = append(lines, fmt.Sprintf("%-60s    %s", typeString, path.Snd))
		}
	}
	return strings.Join(lines, "\n")
}
