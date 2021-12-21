package schema_json

import (
	"encoding/json"
	"github.com/pkg/errors"
	"io/ioutil"
)

type SwaggerProperty struct {
	AdditionalProperties     *SwaggerProperty  `json:"additionalProperties,omitempty"`
	Description              string            `json:"description,omitempty"`
	Items                    map[string]string `json:"items,omitempty"`
	Type                     string            `json:"type,omitempty"`
	Ref                      string            `json:"$ref,omitempty"`
	Format                   string            `json:"format,omitempty"`
	XKubernetesPatchMergeKey string            `json:"x-kubernetes-patch-merge-key,omitempty"`
	XKubernetesPatchStrategy string            `json:"x-kubernetes-patch-strategy,omitempty"`
	XKubernetesListType      string            `json:"x-kubernetes-list-type,omitempty"`
	XKubernetesListMapKeys   []string          `json:"x-kubernetes-list-map-keys,omitempty"`
}

type SwaggerSpec struct {
	Definitions map[string]struct {
		Description                 string                      `json:"description,omitempty"`
		Properties                  map[string]*SwaggerProperty `json:"properties,omitempty"`
		Format                      string                      `json:"format,omitempty"`
		Type                        string                      `json:"type"`
		Required                    []string                    `json:"required,omitempty"`
		XKubernetesGroupVersionKind []struct {
			Group   string `json:"group"`
			Kind    string `json:"kind"`
			Version string `json:"version"`
		} `json:"x-kubernetes-group-version-kind,omitempty"`
	} `json:"definitions"`
	Info struct {
		Title   string `json:"title"`
		Version string `json:"version"`
	} `json:"info"`
	//Paths map[string]interface{}
	//Security int
	//SecurityDefinitions int
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
