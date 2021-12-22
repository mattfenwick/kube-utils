package schema_json

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
)

type SwaggerAdditionalProperty struct {
	Format string            `json:"format,omitempty"`
	Ref    string            `json:"$ref,omitempty"`
	Items  map[string]string `json:"items,omitempty"`
	Type   string            `json:"type,omitempty"`
}

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

func ReadSwaggerSpecs(path string) (*SwaggerSpec, error) {
	in, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to read file %s", path)
	}

	obj := &SwaggerSpec{}
	err = json.Unmarshal(in, obj)

	return obj, errors.Wrapf(err, "unable to unmarshal json")
}
