package swagger

import (
	"fmt"
	"github.com/mattfenwick/collections/pkg/file"
	"github.com/mattfenwick/collections/pkg/json"
	"github.com/mattfenwick/kube-utils/go/pkg/utils"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"os"
)

const (
	SpecsRootDirectory = "./swagger-data"
)

var (
	LatestKubePatchVersions []string
)

func init() {
	// these version numbers come from https://github.com/kubernetes/kubernetes/tree/master/CHANGELOG
	LatestKubePatchVersions = []string{
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
		"1.18.20",
		"1.19.16",
		"1.20.15",
		"1.21.14",
		"1.22.12",
		"1.23.9",
		"1.24.3",
		"1.25.0-alpha.2",
	}
}

func ReadSwaggerSpecFromFile[A any](path string) (*A, error) {
	return json.ParseFile[A](path)
}

func ReadSwaggerSpecFromGithub(version string) (*Spec, error) {
	path := MakePathFromKubeVersion(version)

	if !file.Exists(path) {
		logrus.Infof("file for version %s not found (path %s); downloading instead", version, path)

		err := os.MkdirAll(SpecsRootDirectory, 0777)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to mkdir %s", SpecsRootDirectory)
		}

		err = utils.GetFileFromURL(BuildSwaggerSpecsURLFromKubeVersion(version), path)
		if err != nil {
			return nil, err
		}

		// get the keys sorted
		err = json.SortFileOptions(path, false, true)
		if err != nil {
			return nil, err
		}
	}

	spec, err := ReadSwaggerSpecFromFile[Spec](path)
	utils.DoOrDie(err)

	return spec, nil
}

func MustReadSwaggerSpec(version string) *Spec {
	spec, err := ReadSwaggerSpecFromGithub(version)
	utils.DoOrDie(err)
	return spec
}

func DownloadSwaggerSpec(version string) []byte {
	bytes, err := utils.GetURL(BuildSwaggerSpecsURLFromKubeVersion(version))
	utils.DoOrDie(err)
	return bytes
}

func MakePathFromKubeVersion(version string) string {
	return fmt.Sprintf("%s/%s-swagger-spec.json", SpecsRootDirectory, version)
}

func BuildSwaggerSpecsURLFromKubeVersion(kubeVersion string) string {
	return fmt.Sprintf("https://raw.githubusercontent.com/kubernetes/kubernetes/v%s/api/openapi-spec/swagger.json", kubeVersion)
}
