package schema_json

import (
	"encoding/json"
	"fmt"
	"github.com/mattfenwick/kube-utils/go/pkg/utils"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"strings"
)

type SwaggerAdditionalProperty struct {
	Format string            `json:"format,omitempty"`
	Ref    string            `json:"$ref,omitempty"`
	Items  map[string]string `json:"items,omitempty"`
	Type   string            `json:"type,omitempty"`
}

type SwaggerProperty struct {
	AdditionalProperties *SwaggerAdditionalProperty `json:"additionalProperties,omitempty"`
	Description          string                     `json:"description,omitempty"`
	Format               string                     `json:"format,omitempty"`
	Items                *struct {
		Format string `json:"format,omitempty"`
		Ref    string `json:"$ref,omitempty"`
		Type   string `json:"type,omitempty"`
	} `json:"items,omitempty"`
	Ref                      string   `json:"$ref,omitempty"`
	Type                     string   `json:"type,omitempty"`
	XKubernetesListMapKeys   []string `json:"x-kubernetes-list-map-keys,omitempty"`
	XKubernetesListType      string   `json:"x-kubernetes-list-type,omitempty"`
	XKubernetesPatchMergeKey string   `json:"x-kubernetes-patch-merge-key,omitempty"`
	XKubernetesPatchStrategy string   `json:"x-kubernetes-patch-strategy,omitempty"`
}

type SwaggerDefinition struct {
	Description                 string                      `json:"description,omitempty"`
	Format                      string                      `json:"format,omitempty"`
	Properties                  map[string]*SwaggerProperty `json:"properties,omitempty"`
	Required                    []string                    `json:"required,omitempty"`
	Type                        string                      `json:"type,omitempty"`
	XKubernetesGroupVersionKind []struct {
		Group   string `json:"group"`
		Kind    string `json:"kind"`
		Version string `json:"version"`
	} `json:"x-kubernetes-group-version-kind,omitempty"`
	XKubernetesUnions []map[string]interface{} `json:"x-kubernetes-unions,omitempty"`
}

type SwaggerSpec struct {
	Definitions map[string]*SwaggerDefinition `json:"definitions"`
	Info        struct {
		Title   string `json:"title"`
		Version string `json:"version"`
	} `json:"info"`
	definitionsByNameCache map[string]map[string]*SwaggerDefinition
	//Paths map[string]interface{}
	//Security int
	//SecurityDefinitions int
}

func (s *SwaggerSpec) ResolveRef(ref string) *SwaggerDefinition {
	typeName := ParseRef(ref)
	resolvedType, ok := s.Definitions[typeName]
	if !ok {
		panic(errors.Errorf("unable to resolve type %s", ref))
	}
	return resolvedType
}

func (s *SwaggerSpec) DefinitionsByName() map[string]map[string]*SwaggerDefinition {
	if s.definitionsByNameCache == nil {
		s.definitionsByNameCache = map[string]map[string]*SwaggerDefinition{}
		for name, def := range s.Definitions {
			if len(def.XKubernetesGroupVersionKind) != 1 {
				logrus.Infof("skipping type %s, has multiple groupversionkind", name)
				continue
			}
			gvk := def.XKubernetesGroupVersionKind[0]
			if _, ok := s.definitionsByNameCache[gvk.Kind]; !ok {
				s.definitionsByNameCache[gvk.Kind] = map[string]*SwaggerDefinition{}
			}
			s.definitionsByNameCache[gvk.Kind][gvk.Group+"."+gvk.Version] = def
		}
	}
	return s.definitionsByNameCache
}

func (s *SwaggerSpec) VersionKindLengths() []string {
	var lengths []string
	for name, def := range s.Definitions {
		lengths = append(lengths, fmt.Sprintf("%d: %s", len(def.XKubernetesGroupVersionKind), name))
	}
	return lengths
}

func ParseRef(ref string) string {
	pieces := strings.Split(ref, "/")
	if len(pieces) != 3 {
		panic(errors.Errorf("unable to parse ref: expected 3 pieces, found %d (%s)", len(pieces), ref))
	}
	return pieces[2]
}

func (s *SwaggerSpec) ResolveAllRefsToJsonBlob() interface{} {
	jsonObject, err := utils.JsonRemarshal(s)
	utils.DoOrDie(err)

	return s.ResolveAllRefsToJsonBlobHelper(jsonObject, []string{})
}

func (s *SwaggerSpec) ResolveAllRefsToJsonBlobHelper(obj interface{}, path []string) interface{} {
	if obj == nil {
		return nil
	}
	switch o := obj.(type) {
	case map[string]interface{}:
		out := map[string]interface{}{}
		for k, v := range o {
			if k == "$ref" {
				switch refType := v.(type) {
				case string:
					out[k] = s.ResolveRef(refType)
				case map[string]interface{}:
					// TODO special case?  nothing to do for now
					out[k] = v
				default:
					panic(errors.Errorf("unable to handle type %T", v))
				}
			} else {
				out[k] = s.ResolveAllRefsToJsonBlobHelper(v, append(path, k))
			}
		}
		return out
	case []interface{}:
		var out []interface{}
		for i, v := range o {
			out = append(out, s.ResolveAllRefsToJsonBlobHelper(v, append(path, fmt.Sprintf("%d", i))))
		}
		return out
	case int:
		return o
	case string:
		return o
	case bool:
		return o
	//case types.Nil: // TODO is this necessary?
	default:
		panic(errors.Errorf("unrecognized type: %s, %T, %+v", path, o, o))
	}
}

func ReadSwaggerSpecs(path string) (*SwaggerSpec, error) {
	in, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to read file %s", path)
	}

	obj := &SwaggerSpec{}
	err = json.Unmarshal(in, obj)

	return obj, errors.Wrapf(err, "unable to unmarshal json")
}

//func (s *SwaggerSpec) RecursivelyResolveRef(ref string) *SwaggerDefinitionResolved {
//	typeName := ParseRef(ref)
//	resolvedType, ok := s.Definitions[typeName]
//	if !ok {
//		panic(errors.Errorf("unable to resolve type %s", ref))
//	}
//	return resolvedType
//}
//
//func (s *SwaggerSpec) ResolveType(ref string) *SwaggerDefinitionResolved {
//	typeName := ParseRef(ref)
//	resolvedType, ok := s.Definitions[typeName]
//	if !ok {
//		panic(errors.Errorf("unable to resolve type %s", ref))
//	}
//
//	properties := map[string]*SwaggerPropertyResolved{}
//	for propName, property := range resolvedType.Properties {
//		items := *(struct {
//			Format string                     `json:"format,omitempty"`
//			Ref    *SwaggerDefinitionResolved `json:"$ref,omitempty"`
//			Type   string                     `json:"type,omitempty"`
//		}){
//		Format: "",
//			Ref:    s.ResolveType(property.Items.Ref),
//				Type:   "",
//		}
//		if property.Items == nil {
//			items = nil
//		}
//		properties[propName] = &SwaggerPropertyResolved{
//			Description: property.Description,
//			Format:      property.Format,
//			Items: items,
//			Ref:  s.ResolveType(property.Ref),
//			Type: property.Type,
//		}
//	}
//
//	return &SwaggerDefinitionResolved{
//		Description:                 resolvedType.Description,
//		Format:                      resolvedType.Format,
//		Properties:                  nil,
//		Required:                    resolvedType.Required,
//		Type:                        resolvedType.Type,
//		XKubernetesGroupVersionKind: resolvedType.XKubernetesGroupVersionKind,
//	}
//}
//
//func (s *SwaggerSpec) ResolveTypes() *SwaggerSpecResolved {
//	definitions := map[string]*SwaggerDefinitionResolved{}
//	for name, def := range s.Definitions {
//		properties := map[string]*SwaggerPropertyResolved{}
//		for propName, property := range def.Properties {
//			properties[propName] = &SwaggerPropertyResolved{
//				Description: property.Description,
//				Format:      property.Format,
//				Items: struct {
//					Format string                     `json:"format,omitempty"`
//					Ref    *SwaggerDefinitionResolved `json:"$ref,omitempty"`
//					Type   string                     `json:"type,omitempty"`
//				}{
//					Format: "",
//					Ref:    s.ResolveRef(property.Items.Ref),
//					Type:   "",
//				},
//				Ref:  s.ResolveRef(property.Ref),
//				Type: property.Type,
//			}
//		}
//		definitions[name] = &SwaggerDefinitionResolved{
//			Description:                 def.Description,
//			Format:                      def.Format,
//			Properties:                  properties,
//			Required:                    def.Required,
//			Type:                        def.Type,
//			XKubernetesGroupVersionKind: def.XKubernetesGroupVersionKind,
//		}
//	}
//	return &SwaggerSpecResolved{Definitions: definitions}
//}
//
//type SwaggerPropertyResolved struct {
//	Description string `json:"description,omitempty"`
//	Format      string `json:"format,omitempty"`
//	Items       struct {
//		Format string                     `json:"format,omitempty"`
//		Ref    *SwaggerDefinitionResolved `json:"$ref,omitempty"`
//		Type   string                     `json:"type,omitempty"`
//	} `json:"items,omitempty"`
//	Ref  *SwaggerDefinitionResolved `json:"$ref,omitempty"`
//	Type string                     `json:"type,omitempty"`
//}
//
//type SwaggerDefinitionResolved struct {
//	Description                 string                              `json:"description,omitempty"`
//	Format                      string                              `json:"format,omitempty"`
//	Properties                  map[string]*SwaggerPropertyResolved `json:"properties,omitempty"`
//	Required                    []string                            `json:"required,omitempty"`
//	Type                        string                              `json:"type,omitempty"`
//	XKubernetesGroupVersionKind []struct {
//		Group   string `json:"group"`
//		Kind    string `json:"kind"`
//		Version string `json:"version"`
//	} `json:"x-kubernetes-group-version-kind,omitempty"`
//}
//
//type SwaggerSpecResolved struct {
//	Definitions map[string]*SwaggerDefinitionResolved `json:"definitions"`
//	//definitionsByNameCache map[string]map[string]*SwaggerDefinition
//}
