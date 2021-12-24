package swagger

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
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
		"1.18.19",
		"1.19.11",
		"1.20.7",
		"1.21.2",
		"1.22.4",
		"1.23.0",
	}
}

func ReadSwaggerSpecs(path string) (*Spec, error) {
	in, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to read file %s", path)
	}

	obj := &Spec{}
	err = json.Unmarshal(in, obj)

	return obj, errors.Wrapf(err, "unable to unmarshal json")
}

func MakePathFromKubeVersion(version string) string {
	return fmt.Sprintf("%s/%s-swagger-spec.json", SpecsRootDirectory, version)
}

func BuildSwaggerSpecsURLFromKubeVersion(kubeVersion string) string {
	return fmt.Sprintf("https://raw.githubusercontent.com/kubernetes/kubernetes/v%s/api/openapi-spec/swagger.json", kubeVersion)
}
