package swagger

import (
	"fmt"
	"github.com/mattfenwick/kube-utils/go/pkg/utils"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"strings"
)

type AdditionalProperty struct {
	Format string            `json:"format,omitempty"`
	Ref    string            `json:"$ref,omitempty"`
	Items  map[string]string `json:"items,omitempty"`
	Type   string            `json:"type,omitempty"`
}

type Property struct {
	AdditionalProperties *AdditionalProperty `json:"additionalProperties,omitempty"`
	Description          string              `json:"description,omitempty"`
	Format               string              `json:"format,omitempty"`
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

func (p *Property) ResolveToJsonBlob(resolve func(string) (string, *Definition), path []string, inProgress map[string]bool) map[string]interface{} {
	logrus.Debugf("resolve property path: %+v", path)
	out := map[string]interface{}{}
	if p.Description != "" {
		out["description"] = p.Description
	}
	if p.Format != "" {
		out["format"] = p.Format
	}
	if p.Items != nil {
		items := map[string]interface{}{}
		if p.Items.Format != "" {
			items["format"] = p.Items.Format
		}
		if p.Items.Ref != "" {
			typeName, resolvedType := resolve(p.Items.Ref)
			if !inProgress[typeName] {
				items["$ref"] = resolvedType.ResolveToJsonBlob(resolve, append(path, "items", "$ref", p.Items.Ref), utils.AddKey(inProgress, typeName))
			} else {
				items["type"] = "(circular)"
			}
		}
		if p.Items.Type != "" {
			items["type"] = p.Items.Type
		}
		out["items"] = items
	}
	if p.Ref != "" {
		typeName, resolvedType := resolve(p.Ref)
		if !inProgress[typeName] {
			out["$ref"] = resolvedType.ResolveToJsonBlob(resolve, append(path, "$ref", p.Ref), utils.AddKey(inProgress, typeName))
		} else {
			out["type"] = "(circular)"
		}
	}
	if p.Type != "" {
		out["type"] = p.Type
	}
	return out
}

type GVK struct {
	Group   string `json:"group"`
	Version string `json:"version"`
	Kind    string `json:"kind"`
}

func (g *GVK) GroupVersion() string {
	if g.Group == "" {
		return g.Version
	}
	return fmt.Sprintf("%s.%s", g.Group, g.Version)
}

func (g *GVK) ToString() string {
	return fmt.Sprintf("%s.%s", g.GroupVersion(), g.Kind)
}

type Definition struct {
	Description                 string                   `json:"description,omitempty"`
	Format                      string                   `json:"format,omitempty"`
	Properties                  map[string]*Property     `json:"properties,omitempty"`
	Required                    []string                 `json:"required,omitempty"`
	Type                        string                   `json:"type,omitempty"`
	XKubernetesGroupVersionKind []*GVK                   `json:"x-kubernetes-group-version-kind,omitempty"`
	XKubernetesUnions           []map[string]interface{} `json:"x-kubernetes-unions,omitempty"`
}

func (d *Definition) ResolveToJsonBlob(resolve func(string) (string, *Definition), path []string, inProgress map[string]bool) map[string]interface{} {
	if len(path) > 100 {
		panic(errors.Errorf("TODO -- maximum depth exceeded, this probably indicates a bug: path %+v", path))
	}
	logrus.Debugf("resolve definition path: %+v", path)
	out := map[string]interface{}{}
	if d.Description != "" {
		out["description"] = d.Description
	}
	if d.Format != "" {
		out["format"] = d.Format
	}
	if len(d.Properties) > 0 {
		properties := map[string]interface{}{}
		for propName, property := range d.Properties {
			properties[propName] = property.ResolveToJsonBlob(resolve, append(path, "properties", propName), inProgress)
		}
		out["properties"] = properties
	}
	if d.Required != nil {
		out["required"] = d.Required
	}
	if d.Type != "" {
		out["type"] = d.Type
	}
	return out
}

// Kube14OrNewerSpec models kubernetes API specs for version 14 and later
//   Version 13 and earlier use a slightly different schema and
//   so should not be handled with this type.
type Kube14OrNewerSpec struct {
	Definitions map[string]*Definition `json:"definitions"`
	Info        struct {
		Title   string `json:"title"`
		Version string `json:"version"`
	} `json:"info"`
	definitionsByNameCache map[string]map[string]*Definition
	//Paths map[string]interface{}
	//Security int
	//SecurityDefinitions int
}

func (s *Kube14OrNewerSpec) ResolveRef(ref string) (string, *Definition) {
	typeName := ParseRef(ref)
	resolvedType, ok := s.Definitions[typeName]
	if !ok {
		panic(errors.Errorf("unable to resolve type %s", ref))
	}

	return typeName, resolvedType
}

func (s *Kube14OrNewerSpec) DefinitionsByNameByGroup() map[string]map[string]*Definition {
	if s.definitionsByNameCache == nil {
		s.definitionsByNameCache = map[string]map[string]*Definition{}
		for name, def := range s.Definitions {
			if len(def.XKubernetesGroupVersionKind) != 1 {
				logrus.Debugf("skipping type %s, has %d groupversionkinds", name, len(def.XKubernetesGroupVersionKind))
				continue
			}
			gvk := def.XKubernetesGroupVersionKind[0]
			if _, ok := s.definitionsByNameCache[gvk.Kind]; !ok {
				s.definitionsByNameCache[gvk.Kind] = map[string]*Definition{}
			}
			s.definitionsByNameCache[gvk.Kind][gvk.GroupVersion()] = def
		}
	}
	return s.definitionsByNameCache
}

func (s *Kube14OrNewerSpec) VersionKindLengths() []string {
	var lengths []string
	for name, def := range s.Definitions {
		lengths = append(lengths, fmt.Sprintf("%d: %s", len(def.XKubernetesGroupVersionKind), name))
	}
	return lengths
}

func (s *Kube14OrNewerSpec) ResolveAllToJsonBlob() map[string]interface{} {
	out := map[string]interface{}{}
	for name, def := range s.Definitions {
		out[name] = def.ResolveToJsonBlob(s.ResolveRef, []string{"definitions", name}, map[string]bool{})
	}
	return out
}

func (s *Kube14OrNewerSpec) ResolveToJsonBlob(name string) map[string]interface{} {
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

func ParseGVK(gvk string) *GVK {
	split := strings.Split(gvk, ".")
	if len(split) < 3 {
		panic(errors.Errorf("invalid gvk string: %s", gvk))
	}
	return &GVK{
		Group:   strings.Join(split[:len(split)-2], "."),
		Version: split[len(split)-2],
		Kind:    split[len(split)-1],
	}
}
