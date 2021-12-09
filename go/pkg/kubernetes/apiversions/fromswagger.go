package apiversions

import (
	"encoding/json"
	"fmt"
	"github.com/mattfenwick/kube-utils/go/pkg/simulator"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
)

type SwaggerSpec struct {
	Definitions map[string]struct {
		Description string
		// Properties
		Type                        string
		XKubernetesGroupVersionKind []struct {
			Group   string
			Kind    string
			Version string
		} `json:"x-kubernetes-group-version-kind"`
	}
	Info struct {
		Title   string
		Version string
	}
	//Paths map[string]interface{}
	//Security int
	//SecurityDefinitions int
}

func ParseJsonSpecs() {
	resourceKinds := []string{
		"Service",
		"ClusterRole",
		"ClusterRoleBinding",
		"ConfigMap",
		"CronJob",
		"CustomResourceDefinition",
		"Deployment",
		"Ingress",
		"Job",
		"Role",
		"RoleBinding",
		"Secret",
		"ServiceAccount",
		"StatefulSet",
	}

	previousTable := &ResourcesTable{
		Version: "???",
		Kinds:   map[string][]string{},
	}
	for _, version := range []string{
		"1.16.15",
		"1.17.17",
		"1.18.19",
		"1.19.11",
		"1.20.7",
		"1.21.2",
		"1.22.4",
		"1.23.0",
	} {
		err := os.MkdirAll("./swagger-data", 0777)
		simulator.DoOrDie(err)

		path := fmt.Sprintf("./swagger-data/%s-swagger-spec.json", version)
		err = GetFileFromURL(BuildSwaggerSpecsURL(version), path)
		simulator.DoOrDie(err)

		in, err := ioutil.ReadFile(path)
		simulator.DoOrDie(errors.Wrapf(err, "unable to read file %s", path))
		obj := &SwaggerSpec{}
		err = json.Unmarshal(in, obj)
		simulator.DoOrDie(errors.Wrapf(err, "unable to unmarshal json"))
		resourcesTable := &ResourcesTable{
			Version: version,
			Kinds:   map[string][]string{},
		}
		for a, b := range obj.Definitions {
			if len(b.XKubernetesGroupVersionKind) > 0 {
				logrus.Debugf("%s, %s, %+v\n", a, b.Type, b.XKubernetesGroupVersionKind)
			}
			for _, gvk := range b.XKubernetesGroupVersionKind {
				apiVersion := ""
				if gvk.Group != "" {
					apiVersion = gvk.Group + "."
				}
				apiVersion += gvk.Version
				resourcesTable.Kinds[gvk.Kind] = append(resourcesTable.Kinds[gvk.Kind], apiVersion)
			}
		}
		//fmt.Printf("simple table:\n%s\n", resourcesTable.SimpleTable())

		resourceDiff := previousTable.Diff(resourcesTable)
		fmt.Printf("comparing %s to %s\n%s\n",
			previousTable.Version,
			resourcesTable.Version,
			resourceDiff.Table(Set(resourceKinds), Set([]string{})))

		previousTable = resourcesTable
	}
}

func BuildSwaggerSpecsURL(kubeVersion string) string {
	return fmt.Sprintf("https://raw.githubusercontent.com/kubernetes/kubernetes/v%s/api/openapi-spec/swagger.json", kubeVersion)
}
