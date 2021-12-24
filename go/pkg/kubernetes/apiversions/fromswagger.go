package apiversions

import (
	"fmt"
	"github.com/mattfenwick/kube-utils/go/pkg/swagger"
	"github.com/mattfenwick/kube-utils/go/pkg/utils"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
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

	err := os.MkdirAll(swagger.SpecsRootDirectory, 0777)
	utils.DoOrDie(errors.Wrapf(err, "unable to mkdir %s", swagger.SpecsRootDirectory))

	previousTable := &ResourcesTable{
		Version: "???",
		Kinds:   map[string][]string{},
	}

	for _, version := range swagger.LatestKubePatchVersions {
		path := fmt.Sprintf("%s/%s-swagger-spec.json", swagger.SpecsRootDirectory, version)
		err = utils.GetFileFromURL(swagger.BuildSwaggerSpecsURLFromKubeVersion(version), path)
		utils.DoOrDie(err)

		obj, err := swagger.ReadSwaggerSpecs(path)
		utils.DoOrDie(err)

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
			resourceDiff.Table(utils.Set(includeResources), utils.Set(excludeResources)))

		previousTable = resourcesTable
	}
}
