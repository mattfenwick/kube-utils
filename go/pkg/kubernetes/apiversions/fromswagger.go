package apiversions

import (
	"encoding/json"
	"fmt"
	schema_json "github.com/mattfenwick/kube-utils/go/pkg/schema-json"
	"github.com/mattfenwick/kube-utils/go/pkg/utils"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
)

func ParseJsonSpecs() {
	excludeResources := []string{"WatchEvent", "DeleteOptions"}
	includeResources := []string{
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
	// these version numbers come from https://github.com/kubernetes/kubernetes/tree/master/CHANGELOG
	for _, version := range []string{
		// for some reason, there's nothing listed for 1.1
		//"1.2.7", // for some reason, these don't show up
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
		"1.18.19",
		"1.19.11",
		"1.20.7",
		"1.21.2",
		"1.22.4",
		"1.23.0",
	} {
		err := os.MkdirAll("./swagger-data", 0777)
		utils.DoOrDie(err)

		path := fmt.Sprintf("./swagger-data/%s-swagger-spec.json", version)
		//err = GetFileFromURL(BuildSwaggerSpecsURL(version), path)
		utils.DoOrDie(err)

		in, err := ioutil.ReadFile(path)
		utils.DoOrDie(errors.Wrapf(err, "unable to read file %s", path))
		obj := &schema_json.SwaggerSpec{}
		err = json.Unmarshal(in, obj)
		utils.DoOrDie(errors.Wrapf(err, "unable to unmarshal json"))
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
			resourceDiff.Table(Set(includeResources), Set(excludeResources)))

		previousTable = resourcesTable
	}
}

func BuildSwaggerSpecsURL(kubeVersion string) string {
	return fmt.Sprintf("https://raw.githubusercontent.com/kubernetes/kubernetes/v%s/api/openapi-spec/swagger.json", kubeVersion)
}
