package swagger

import (
	"encoding/json"
	"fmt"
	"github.com/mattfenwick/kube-utils/go/pkg/utils"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io/ioutil"
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

func ReadSwaggerSpec[A any](path string) (*A, error) {
	in, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to read file %s", path)
	}

	var obj A
	err = json.Unmarshal(in, &obj)

	return &obj, errors.Wrapf(err, "unable to unmarshal json")
}

func MustReadSwaggerSpec(version string) *Spec {
	path := fmt.Sprintf("%s/%s-swagger-spec.json", SpecsRootDirectory, version)

	exists, err := utils.FileExists(path)
	utils.DoOrDie(err)

	if !exists {
		logrus.Infof("file for version %s not found (path %s); downloading instead", version, path)

		err = os.MkdirAll(SpecsRootDirectory, 0777)
		utils.DoOrDie(errors.Wrapf(err, "unable to mkdir %s", SpecsRootDirectory))

		utils.DoOrDie(utils.GetFileFromURL(BuildSwaggerSpecsURLFromKubeVersion(version), path))

		// get the keys sorted
		utils.DoOrDie(utils.JsonUnmarshalMarshal(path))
	}

	spec, err := ReadSwaggerSpec[Spec](path)
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
