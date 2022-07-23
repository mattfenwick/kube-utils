package swagger

import (
	"fmt"
	"github.com/mattfenwick/collections/pkg/base"
	"github.com/mattfenwick/collections/pkg/slice"
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

type Definition struct {
	Description                 string               `json:"description,omitempty"`
	Format                      string               `json:"format,omitempty"`
	Properties                  map[string]*Property `json:"properties,omitempty"`
	Required                    []string             `json:"required,omitempty"`
	Type                        string               `json:"type,omitempty"`
	XKubernetesGroupVersionKind []struct {
		Group   string `json:"group"`
		Kind    string `json:"kind"`
		Version string `json:"version"`
	} `json:"x-kubernetes-group-version-kind,omitempty"`
	XKubernetesUnions []map[string]interface{} `json:"x-kubernetes-unions,omitempty"`
}

func (d *Definition) ResolveToJsonBlob(resolve func(string) (string, *Definition), path []string, inProgress map[string]bool) map[string]interface{} {
	if len(path) > 100 {
		panic("TODO")
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
			gv := gvk.Version
			if gvk.Group != "" {
				gv = gvk.Group + "." + gv
			}
			s.definitionsByNameCache[gvk.Kind][gv] = def
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

// KubeVersion

type KubeVersion []string

var (
	CompareKubeVersion = slice.CompareSlicePairwise[string]()
)

func NewVersion(v string) (KubeVersion, error) {
	pieces := strings.Split(v, ".")
	if len(pieces) < 3 {
		return nil, errors.Errorf("expected at least 3 pieces, found [%+v]", pieces)
	}
	return pieces, nil
}

func MustVersion(v string) KubeVersion {
	version, err := NewVersion(v)
	utils.DoOrDie(err)
	return version
}

func (v KubeVersion) Compare(b KubeVersion) base.Ordering {
	return CompareKubeVersion(v, b)
}

func (v KubeVersion) ToString() string {
	return strings.Join(v, ".")
}

func (v KubeVersion) SwaggerSpecURL() string {
	return fmt.Sprintf(GithubOpenapiURLTemplate, v.ToString())
}

var (
	GithubOpenapiURLTemplate = "https://raw.githubusercontent.com/kubernetes/kubernetes/v%s/api/openapi-spec/swagger.json"

	// LatestKubePatchVersionStrings records the latest known patch versions for each minor version
	//   these version numbers come from https://github.com/kubernetes/kubernetes/tree/master/CHANGELOG
	LatestKubePatchVersionStrings = []string{
		// there's nothing listed for 1.1
		//"1.2.7", // for some reason, these don't show up on the openapi github specs
		//"1.3.10",
		//"1.4.12",
		"1.5.8",
		"1.6.13",
		"1.7.16",
		"1.8.15",
		"1.9.11",
		"1.10.13",
		"1.11.10",
		"1.12.10",
		"1.13.12",
		"1.14.10",
		"1.15.12",
		"1.16.15",
		"1.17.17",
		"1.18.20",
		"1.19.16",
		"1.20.15",
		"1.21.14",
		"1.22.12",
		"1.23.9",
		"1.24.3",
		"1.25.0-alpha.2",
	}

	LatestKubePatchVersions = slice.Map(MustVersion, LatestKubePatchVersionStrings)
)
