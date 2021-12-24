package schema_json

import (
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"sort"
	"strings"
)

func SwaggerAnalysisTypeTable(o interface{}) string {
	tableString := &strings.Builder{}
	table := tablewriter.NewWriter(tableString)
	table.SetAutoWrapText(false)
	table.SetRowLine(true)
	table.SetAutoMergeCells(true)
	table.SetColMinWidth(0, 100)
	table.SetHeader([]string{"Field", "Type"})
	for _, values := range SwaggerAnalysisGetTypes(o, []string{}) {
		table.Append([]string{values[0], values[1]})
	}
	table.Render()
	return tableString.String()
}

func SwaggerAnalysisTypeSummary(obj interface{}) []string {
	var lines []string
	for _, t := range SwaggerAnalysisGetTypes(obj, []string{}) {
		chunks := strings.Split(t[0], ".")
		lines = append(lines,
			fmt.Sprintf(
				"%s%s: %s",
				strings.Repeat("  ", len(chunks)-1),
				chunks[len(chunks)-1],
				t[1]))
	}
	return lines
}

func SwaggerAnalysisGetTypes(obj interface{}, pathContext []string) [][2]string {
	path := make([]string, len(pathContext))
	copy(path, pathContext)

	logrus.Debugf("path: %+v", path)

	var out [][2]string
	switch o := obj.(type) {
	case *Primitive:
		out = append(out, [2]string{strings.Join(path, "."), o.Type})
	case *Array:
		out = append(out, [2]string{strings.Join(path, "."), "array"})
		out = append(out, SwaggerAnalysisGetTypes(o.ElementType, append(path, "[]"))...)
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
			out = append(out, SwaggerAnalysisGetTypes(o.Fields[fieldName], append(path, fieldName))...)
		}
	default:
		panic(errors.Errorf("invalid type: %T", o))
	}
	return out
}

type Primitive struct {
	Type string
}

type Array struct {
	ElementType interface{}
}

type Dict struct {
	ElementType interface{} // TODO not sure if this is necessary, is it always just string?
}

type Object struct {
	Fields map[string]interface{}
}

func (s *SwaggerSpec) AnalyzeType(typeName string) map[string]interface{} {
	jsonBlob := s.ResolveToJsonBlob(typeName)
	out := map[string]interface{}{}
	for group, typeDef := range jsonBlob {
		out[group] = analyzeTypeHelper("ingress", typeDef.(map[string]interface{}), []string{typeName, group})
	}
	return out
}

func analyzeTypeHelper(name string, o map[string]interface{}, pathContext []string) interface{} {
	path := make([]string, len(pathContext))
	copy(path, pathContext)

	logrus.Debugf("path: %+v", path)

	if o["type"] == nil && o["$ref"] == nil {
		panic(errors.Errorf("unable to parse type: nil type and ref (%+v)", o))
	}
	if o["type"] != nil && o["$ref"] != nil {
		panic(errors.Errorf("unable to parse type: both type and ref non-nil (%+v)", o))
	}

	var t string
	if v, ok := o["type"]; ok {
		t = v.(string)
	}
	var r map[string]interface{}
	if v, ok := o["$ref"]; ok {
		r = v.(map[string]interface{})
	}
	if t != "" {
		switch t {
		case "string", "integer", "boolean":
			return &Primitive{Type: t}
		case "object":
			if v, ok := o["properties"]; ok {
				fields := map[string]interface{}{}
				for propName, property := range v.(map[string]interface{}) {
					fields[propName] = analyzeTypeHelper(propName, property.(map[string]interface{}), append(path, propName))
				}
				return &Object{Fields: fields}
			} else {
				// TODO not sure if this is right
				return &Dict{ElementType: &Primitive{Type: "string"}}
			}
		case "array":
			elementType := analyzeTypeHelper(name, o["items"].(map[string]interface{}), append(path, "items"))
			return &Array{ElementType: elementType}
		default:
			panic(errors.Errorf("TODO -- handle %s", o["type"]))
		}
	} else if r != nil {
		return analyzeTypeHelper(name, r, append(path, "$ref"))
	} else {
		panic(errors.Errorf("shouldn't happen"))
	}
}
