package swagger

import (
	"fmt"
	"github.com/mattfenwick/kube-utils/go/pkg/kubernetes/swagger/apiversions"
	"github.com/mattfenwick/kube-utils/go/pkg/utils"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"os"
)

func CompareGVKs(excludeResources []string, includeResources []string, kubeVersions []KubeVersion) {

	err := os.MkdirAll(SpecsRootDirectory, 0777)
	utils.DoOrDie(errors.Wrapf(err, "unable to mkdir %s", SpecsRootDirectory))

	previousTable := &apiversions.ResourcesTable{
		Version: "???",
		Kinds:   map[string][]string{},
	}

	for _, version := range kubeVersions {
		obj := MustReadSwaggerSpec(version)

		resourcesTable := &apiversions.ResourcesTable{
			Version: version.ToString(),
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
