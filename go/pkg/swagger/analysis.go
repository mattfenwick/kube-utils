package swagger

import (
	"fmt"
	"github.com/mattfenwick/kube-utils/go/pkg/utils"
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/maps"
	"sort"
	"strings"
)

func AnalysisTypeTable(o interface{}) string {
	tableString := &strings.Builder{}
	table := tablewriter.NewWriter(tableString)
	table.SetAutoWrapText(false)
	table.SetRowLine(true)
	table.SetAutoMergeCells(true)
	table.SetColMinWidth(1, 100)
	table.SetHeader([]string{"Type", "Field"})
	for _, values := range AnalysisGetTypes(o, []string{}) {
		table.Append([]string{values[1], values[0]})
	}
	table.Render()
	return tableString.String()
}

func AnalysisTypeSummary(obj interface{}) []string {
	var lines []string
	for _, t := range AnalysisGetTypes(obj, []string{}) {
		chunks := strings.Split(t[0], ".")
		prefix := strings.Repeat("  ", len(chunks)-1)
		typeString := fmt.Sprintf("%s%s", prefix, chunks[len(chunks)-1])
		line := fmt.Sprintf("%-60s    %s", typeString, t[1])
		lines = append(lines, line)
	}
	return lines
}

func AnalysisGetTypes(obj interface{}, pathContext []string) [][2]string {
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
		out = append(out, AnalysisGetTypes(o.ElementType, append(path, "[]"))...)
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
			out = append(out, AnalysisGetTypes(o.Fields[fieldName], append(path, fieldName))...)
		}
	default:
		panic(errors.Errorf("invalid type: %T", o))
	}
	return out
}

func CompareAnalysisTypes(a interface{}, b interface{}) *utils.JsonDocumentDiffs {
	diffs := &utils.JsonDocumentDiffs{}
	CompareAnalysisTypesHelper(a, b, []string{}, diffs)
	return diffs
}

func CopySlice(s []string) []string {
	newCopy := make([]string, len(s))
	copy(newCopy, s)
	return newCopy
}

func CompareAnalysisTypesHelper(a interface{}, b interface{}, pathContext []string, diffs *utils.JsonDocumentDiffs) {
	// make a copy to avoid aliasing
	path := CopySlice(pathContext)

	logrus.Debugf("path: %+v", path)

	if a == nil && b != nil {
		diffs.Add(&utils.JDiff{Type: utils.DiffTypeAdd, Old: a, New: b, Path: path})
	} else if b == nil {
		diffs.Add(&utils.JDiff{Type: utils.DiffTypeRemove, Old: a, New: b, Path: path})
	} else {
		switch aVal := a.(type) {
		case *Any:
			switch bVal := b.(type) {
			case *Any:
				// nothing to do
			default:
				diffs.Add(&utils.JDiff{Type: utils.DiffTypeChange, Old: aVal, New: bVal, Path: path})
			}
		case *Circular:
			switch bVal := b.(type) {
			case *Circular:
				// nothing to do
			default:
				diffs.Add(&utils.JDiff{Type: utils.DiffTypeChange, Old: aVal, New: bVal, Path: path})
			}
		case *Dict:
			switch bVal := b.(type) {
			case *Dict:
				CompareAnalysisTypesHelper(aVal.ElementType, bVal.ElementType, append(path, "{}"), diffs)
			default:
				diffs.Add(&utils.JDiff{Type: utils.DiffTypeChange, Old: aVal, New: bVal, Path: path})
			}
		case *Object:
			switch bVal := b.(type) {
			case *Object:
				aKeys := maps.Keys(aVal.Fields)
				sort.Strings(aKeys)
				for _, k := range aKeys {
					CompareAnalysisTypesHelper(aVal.Fields[k], bVal.Fields[k], append(path, fmt.Sprintf(`%s`, k)), diffs)
				}
				bKeys := maps.Keys(bVal.Fields)
				sort.Strings(bKeys)
				for _, k := range bKeys {
					if _, ok := aVal.Fields[k]; !ok {
						diffs.Add(&utils.JDiff{Type: utils.DiffTypeAdd, New: bVal.Fields[k], Path: append(path, fmt.Sprintf(`%s`, k))})
					}
				}
				// compare 'required' fields:
				minLength := len(aVal.Required)
				if len(bVal.Required) < minLength {
					minLength = len(bVal.Required)
				}
				for i, aSub := range aVal.Required {
					newPath := append(CopySlice(path), "required", fmt.Sprintf("%d", i))
					if i >= len(aVal.Required) {
						diffs.Add(&utils.JDiff{Type: utils.DiffTypeAdd, New: bVal.Required[i], Path: newPath})
					} else if i >= len(bVal.Required) {
						diffs.Add(&utils.JDiff{Type: utils.DiffTypeRemove, Old: aSub, Path: newPath})
					} else if aSub != bVal.Required[i] {
						diffs.Add(&utils.JDiff{Type: utils.DiffTypeChange, Old: aSub, New: bVal.Required[i], Path: newPath})
					}
				}
			default:
				diffs.Add(&utils.JDiff{Type: utils.DiffTypeChange, Old: aVal, New: bVal, Path: path})
			}
		case *Array:
			switch bVal := b.(type) {
			case *Array:
				CompareAnalysisTypesHelper(aVal.ElementType, bVal.ElementType, append(path, "[]"), diffs)
			default:
				diffs.Add(&utils.JDiff{Type: utils.DiffTypeChange, Old: aVal, New: bVal, Path: path})
			}
		case *Primitive:
			switch bVal := b.(type) {
			case *Primitive:
				if aVal.Type != bVal.Type {
					diffs.Add(&utils.JDiff{Type: utils.DiffTypeChange, Old: aVal, New: bVal, Path: path})
				}
			default:
				diffs.Add(&utils.JDiff{Type: utils.DiffTypeChange, Old: aVal, New: bVal, Path: path})
			}
		default:
			panic(errors.Errorf("unrecognized type at path %s: %T (value %+v)", path, aVal, aVal))
		}
	}
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
	Fields   map[string]interface{}
	Required []string
}

type Any struct{}

type Circular struct{} // TODO at some point, add in a string to refer to type by name?

func (s *Kube14OrNewerSpec) AnalyzeType(typeName string) map[string]interface{} {
	jsonBlob := s.ResolveToJsonBlob(typeName)
	out := map[string]interface{}{}
	var sortedGroups []string
	for group := range jsonBlob {
		sortedGroups = append(sortedGroups, group)
	}
	sort.Strings(sortedGroups)
	for _, group := range sortedGroups {
		typeDef := jsonBlob[group]
		out[group] = analyzeTypeHelper("ingress", typeDef.(map[string]interface{}), []string{typeName, group})
	}
	return out
}

func analyzeTypeHelper(name string, o map[string]interface{}, pathContext []string) interface{} {
	path := make([]string, len(pathContext))
	copy(path, pathContext)

	logrus.Debugf("path: %+v", path)

	if o["type"] == nil && o["$ref"] == nil {
		logrus.Debugf("(this happens in older specs and is probably okay) both 'type' and '$ref' are nil at: %+v , %s", pathContext, name)
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
		if o["required"] != nil {
			if t != "object" {
				panic(errors.Errorf("'required' field found on non-object: %+v", o))
			} else {
				logrus.Debugf("sanity check: found object with 'required' field")
			}
		}
		switch t {
		case "string", "integer", "boolean", "number":
			return &Primitive{Type: t}
		case "(circular)":
			return &Circular{}
		case "object":
			return AnalyzeObject(o, path)
		case "array":
			elementType := analyzeTypeHelper(name, o["items"].(map[string]interface{}), append(path, "items"))
			return &Array{ElementType: elementType}
		default:
			panic(errors.Errorf("TODO -- handle %s", o["type"]))
		}
	} else if r != nil {
		return analyzeTypeHelper(name, r, append(path, "$ref"))
	} else {
		// assume it's an object
		return AnalyzeObject(o, path)
	}
}

func AnalyzeObject(o map[string]interface{}, path []string) interface{} {
	if v, ok := o["properties"]; ok {
		fields := map[string]interface{}{}
		for propName, property := range v.(map[string]interface{}) {
			fields[propName] = analyzeTypeHelper(propName, property.(map[string]interface{}), append(path, propName))
		}
		var required []string
		if rs, ok := o["required"]; ok {
			required = rs.([]string)
		}
		return &Object{Fields: fields, Required: required}
	} else {
		if o["required"] != nil {
			panic(errors.Errorf("'required' field found on dict: %+v", o))
		} else {
			logrus.Debugf("sanity check: Dict does not have 'required' field")
		}
		// TODO not sure if this is right: can we always assume dicts have elements of type string?
		return &Dict{ElementType: &Primitive{Type: "string"}}
	}
}
