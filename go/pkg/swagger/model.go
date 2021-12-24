package swagger

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

func (s *SwaggerProperty) ResolveToJsonBlob(resolve func(string) (string, *SwaggerDefinition), path []string, inProgress map[string]bool) map[string]interface{} {
	logrus.Debugf("resolve property path: %+v", path)
	out := map[string]interface{}{}
	if s.Description != "" {
		out["description"] = s.Description
	}
	if s.Format != "" {
		out["format"] = s.Format
	}
	if s.Items != nil {
		items := map[string]interface{}{}
		if s.Items.Format != "" {
			items["format"] = s.Items.Format
		}
		if s.Items.Ref != "" {
			typeName, resolvedType := resolve(s.Items.Ref)
			if !inProgress[typeName] {
				items["$ref"] = resolvedType.ResolveToJsonBlob(resolve, append(path, "items", "$ref", s.Items.Ref), utils.AddKey(inProgress, typeName))
			} else {
				items["$ref"] = "(circular)"
			}
		}
		if s.Items.Type != "" {
			items["type"] = s.Items.Type
		}
		out["items"] = items
	}
	if s.Ref != "" {
		typeName, resolvedType := resolve(s.Ref)
		if !inProgress[typeName] {
			out["$ref"] = resolvedType.ResolveToJsonBlob(resolve, append(path, "$ref", s.Ref), utils.AddKey(inProgress, typeName))
		} else {
			out["$ref"] = "(circular)"
		}
	}
	if s.Type != "" {
		out["type"] = s.Type
	}
	return out
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

func (s *SwaggerDefinition) ResolveToJsonBlob(resolve func(string) (string, *SwaggerDefinition), path []string, inProgress map[string]bool) map[string]interface{} {
	if len(path) > 100 {
		panic("TODO")
	}
	logrus.Debugf("resolve definition path: %+v", path)
	out := map[string]interface{}{}
	if s.Description != "" {
		out["description"] = s.Description
	}
	if s.Format != "" {
		out["format"] = s.Format
	}
	if len(s.Properties) > 0 {
		properties := map[string]interface{}{}
		for propName, property := range s.Properties {
			properties[propName] = property.ResolveToJsonBlob(resolve, append(path, "properties", propName), inProgress)
		}
		out["properties"] = properties
	}
	if s.Required != nil {
		out["required"] = s.Required
	}
	if s.Type != "" {
		out["type"] = s.Type
	}
	return out
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

func (s *SwaggerSpec) ResolveRef(ref string) (string, *SwaggerDefinition) {
	typeName := ParseRef(ref)
	resolvedType, ok := s.Definitions[typeName]
	if !ok {
		panic(errors.Errorf("unable to resolve type %s", ref))
	}

	return typeName, resolvedType
}

func (s *SwaggerSpec) DefinitionsByNameByGroup() map[string]map[string]*SwaggerDefinition {
	if s.definitionsByNameCache == nil {
		s.definitionsByNameCache = map[string]map[string]*SwaggerDefinition{}
		for name, def := range s.Definitions {
			if len(def.XKubernetesGroupVersionKind) != 1 {
				logrus.Debugf("skipping type %s, has %d groupversionkinds", name, len(def.XKubernetesGroupVersionKind))
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

func (s *SwaggerSpec) ResolveAllToJsonBlob() map[string]interface{} {
	out := map[string]interface{}{}
	for name, def := range s.Definitions {
		out[name] = def.ResolveToJsonBlob(s.ResolveRef, []string{"definitions", name}, map[string]bool{})
	}
	return out
}

func (s *SwaggerSpec) ResolveToJsonBlob(name string) map[string]interface{} {
	out := map[string]interface{}{}
	for groupName, def := range s.DefinitionsByNameByGroup()[name] {
		out[groupName] = def.ResolveToJsonBlob(s.ResolveRef, []string{groupName, name}, map[string]bool{})
	}
	return out
}

func ParseRef(ref string) string {
	pieces := strings.Split(ref, "/")
	if len(pieces) != 3 {
		panic(errors.Errorf("unable to parse ref: expected 3 pieces, found %d (%s)", len(pieces), ref))
	}
	return pieces[2]
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
