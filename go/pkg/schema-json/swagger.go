package schema_json

import (
	"encoding/json"
	"fmt"
	"github.com/mattfenwick/kube-utils/go/pkg/utils"
	"github.com/pkg/errors"
	"io/ioutil"
)

type SwaggerAdditionalProperty struct {
	Format string            `json:"format,omitempty"`
	Ref    string            `json:"$ref,omitempty"`
	Items  map[string]string `json:"items,omitempty"`
	Type   string            `json:"type,omitempty"`
}

type Change struct {
	FieldName string
	Old       string
	New       string
	MapDiff   *utils.MapDiff
}

func (s *SwaggerAdditionalProperty) Compare(other *SwaggerAdditionalProperty) []*Change {
	var changes []*Change
	if s.Format != other.Format {
		changes = append(changes, &Change{FieldName: "Format", Old: s.Format, New: other.Format})
	}
	if s.Ref != other.Ref {
		changes = append(changes, &Change{FieldName: "Ref", Old: s.Ref, New: other.Ref})
	}
	if s.Type != other.Type {
		changes = append(changes, &Change{FieldName: "Type", Old: s.Type, New: other.Type})
	}
	itemsDiff := utils.CompareMaps(s.Items, other.Items)
	if !itemsDiff.IsSame() {
		changes = append(changes, &Change{FieldName: "Format", Old: s.Format, New: other.Format})
	}
	return changes
}

//type SwaggerAdditionalPropertyDiff struct {
//	Attributes *utils.MapDiff
//	Items      *utils.MapDiff
//}

//func (s *SwaggerAdditionalProperty) Dict() map[string]string {
//	return utils.MapFilterEmptyValues(map[string]string{
//		"format": s.Format,
//		"ref":    s.Ref,
//		"type":   s.Type,
//	})
//}

//func (s *SwaggerAdditionalProperty) Diff(other *SwaggerAdditionalProperty) *SwaggerAdditionalPropertyDiff {
//	// TODO handle nils in either side
//	return &SwaggerAdditionalPropertyDiff{
//		Items:      utils.CompareMaps(s.Items, other.Items),
//		Attributes: utils.CompareMaps(s.Dict(), other.Dict()),
//	}
//}

type SwaggerProperty struct {
	AdditionalProperties     *SwaggerAdditionalProperty `json:"additionalProperties,omitempty"`
	Description              string                     `json:"description,omitempty"`
	Format                   string                     `json:"format,omitempty"`
	Items                    map[string]string          `json:"items,omitempty"`
	Ref                      string                     `json:"$ref,omitempty"`
	Type                     string                     `json:"type,omitempty"`
	XKubernetesListMapKeys   []string                   `json:"x-kubernetes-list-map-keys,omitempty"`
	XKubernetesListType      string                     `json:"x-kubernetes-list-type,omitempty"`
	XKubernetesPatchMergeKey string                     `json:"x-kubernetes-patch-merge-key,omitempty"`
	XKubernetesPatchStrategy string                     `json:"x-kubernetes-patch-strategy,omitempty"`
}

//func (s *SwaggerProperty)Compare(other *SwaggerProperty) []*Change {
//	var changes []*Change
//
//	if s.AdditionalProperties == nil && other.AdditionalProperties == nil {
//
//	} else if s.AdditionalProperties == nil {
//
//	} else if other.AdditionalProperties == nil {
//
//	} else {
//
//	}
//
//	if s.Format != other.Format {
//		changes = append(changes, &Change{FieldName: "Format", Old: s.Format, New: other.Format})
//	}
//	if s.Ref != other.Ref {
//		changes = append(changes, &Change{FieldName: "Ref", Old: s.Ref, New: other.Ref})
//	}
//	if s.Type != other.Type {
//		changes = append(changes, &Change{FieldName: "Type", Old: s.Type, New: other.Type})
//	}
//	itemsDiff := utils.CompareMaps(s.Items, other.Items)
//	if !itemsDiff.IsSame() {
//		changes = append(changes, &Change{FieldName: "Format", Old: s.Format, New: other.Format})
//	}
//	return changes
//}

//type SwaggerPropertyDiff struct {
//	AdditionalProperties *SwaggerAdditionalPropertyDiff
//	Attributes           *utils.MapDiff
//	Items                *utils.MapDiff
//}

//func (s *SwaggerProperty) Dict() map[string]string {
//	return utils.MapFilterEmptyValues(map[string]string{
//		"description": s.Description,
//		"format":      s.Format,
//		"ref":         s.Ref,
//		"type":        s.Type,
//	})
//}

//func (s *SwaggerProperty) Diff(other *SwaggerProperty) *SwaggerPropertyDiff {
//	return &SwaggerPropertyDiff{
//		AdditionalProperties: s.AdditionalProperties.Diff(other.AdditionalProperties),
//		Items:                utils.CompareMaps(s.Items, other.Items),
//		Attributes:           utils.CompareMaps(s.Dict(), other.Dict()),
//	}
//}

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

//func (s *SwaggerDefinition) Dict() map[string]string {
//	return utils.MapFilterEmptyValues(map[string]string{
//		"description": s.Description,
//		"format":      s.Format,
//		"type":        s.Type,
//	})
//}
//
//type SwaggerDefinitionDiff struct {
//	Attributes *utils.MapDiff
//	Properties int // TODO what's the type here? *utils.MapDiff
//}
//
//func (s *SwaggerDefinition) Diff(other *SwaggerDefinition) *SwaggerDefinitionDiff {
//	return &SwaggerDefinitionDiff{
//		Attributes: utils.CompareMaps(s.Dict(), other.Dict()),
//		Properties: 0,
//	}
//}

type SwaggerSpec struct {
	Definitions map[string]*SwaggerDefinition `json:"definitions"`
	Info        struct {
		Title   string `json:"title"`
		Version string `json:"version"`
	} `json:"info"`
	//Paths map[string]interface{}
	//Security int
	//SecurityDefinitions int
}

func (s *SwaggerSpec) VersionKindLengths() []string {
	var lengths []string
	for name, def := range s.Definitions {
		lengths = append(lengths, fmt.Sprintf("%d: %s", len(def.XKubernetesGroupVersionKind), name))
	}
	return lengths
}

//func (s *SwaggerSpec) Diff(other *SwaggerSpec) error {
//	panic("TODO")
//}

func ReadSwaggerSpecs(path string) (*SwaggerSpec, error) {
	in, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to read file %s", path)
	}

	obj := &SwaggerSpec{}
	err = json.Unmarshal(in, obj)

	return obj, errors.Wrapf(err, "unable to unmarshal json")
}
