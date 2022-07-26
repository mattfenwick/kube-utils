package swagger

import (
	"fmt"
	"github.com/mattfenwick/collections/pkg/slice"
	"github.com/mattfenwick/kube-utils/go/pkg/utils"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/maps"
)

func CompareResolvedResources(a interface{}, b interface{}) *utils.JsonDocumentDiffs {
	diffs := &utils.JsonDocumentDiffs{}
	CompareResolvedResourcesHelper(a, b, []string{}, diffs)
	return diffs
}

func CompareResolvedResourcesHelper(a interface{}, b interface{}, pathContext []string, diffs *utils.JsonDocumentDiffs) {
	// make a copy to avoid aliasing
	path := utils.CopySlice(pathContext)

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
				CompareResolvedResourcesHelper(aVal.ElementType, bVal.ElementType, append(path, "{}"), diffs)
			default:
				diffs.Add(&utils.JDiff{Type: utils.DiffTypeChange, Old: aVal, New: bVal, Path: path})
			}
		case *Object:
			switch bVal := b.(type) {
			case *Object:
				for _, k := range slice.Sort(maps.Keys(aVal.Fields)) {
					CompareResolvedResourcesHelper(aVal.Fields[k], bVal.Fields[k], append(path, fmt.Sprintf(`%s`, k)), diffs)
				}
				for _, k := range slice.Sort(maps.Keys(bVal.Fields)) {
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
					newPath := append(utils.CopySlice(path), "required", fmt.Sprintf("%d", i))
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
				CompareResolvedResourcesHelper(aVal.ElementType, bVal.ElementType, append(path, "[]"), diffs)
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

func ResolveResource(s *Kube14OrNewerSpec, typeName string) map[string]interface{} {
	jsonBlob := s.ResolveToJsonBlob(typeName)
	out := map[string]interface{}{}

	for _, group := range slice.Sort(maps.Keys(jsonBlob)) {
		typeDef := jsonBlob[group]
		out[group] = resolveResourceHelper(typeDef.(map[string]interface{}), []string{typeName, group})
	}
	return out
}

func resolveResourceHelper(o map[string]interface{}, pathContext []string) interface{} {
	path := make([]string, len(pathContext))
	copy(path, pathContext)

	logrus.Debugf("path: %+v", path)

	if o["type"] == nil && o["$ref"] == nil {
		logrus.Debugf("(this happens in older specs and is probably okay) both 'type' and '$ref' are nil at: %+v", pathContext)
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
			return ResolveObject(o, path)
		case "array":
			elementType := resolveResourceHelper(o["items"].(map[string]interface{}), append(path, "items"))
			return &Array{ElementType: elementType}
		default:
			panic(errors.Errorf("TODO -- handle %s", o["type"]))
		}
	} else if r != nil {
		return resolveResourceHelper(r, append(path, "$ref"))
	} else {
		// assume it's an object
		return ResolveObject(o, path)
	}
}

func ResolveObject(o map[string]interface{}, path []string) interface{} {
	if v, ok := o["properties"]; ok {
		fields := map[string]interface{}{}
		for propName, property := range v.(map[string]interface{}) {
			fields[propName] = resolveResourceHelper(property.(map[string]interface{}), append(path, propName))
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
