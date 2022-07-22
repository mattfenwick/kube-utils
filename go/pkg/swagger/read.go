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

func ReadSwaggerSpecFromFile[A any](path string) (*A, error) {
	return json.ParseFile[A](path)
}

func ReadSwaggerSpecFromGithub(version Version) (*Kube14OrNewerSpec, error) {
	path := MakePathFromKubeVersion(version)

	if !file.Exists(path) {
		logrus.Infof("file for version %s not found (path %s); downloading instead", version, path)

		err := os.MkdirAll(SpecsRootDirectory, 0777)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to mkdir %s", SpecsRootDirectory)
		}

		err = utils.GetFileFromURL(version.SwaggerSpecURL(), path)
		if err != nil {
			return nil, err
		}

		// get the keys sorted
		err = json.SortFileOptions(path, false, true)
		if err != nil {
			return nil, err
		}
	}

	spec, err := ReadSwaggerSpecFromFile[Kube14OrNewerSpec](path)
	utils.DoOrDie(err)

	return spec, nil
}

func MustReadSwaggerSpec(version Version) *Kube14OrNewerSpec {
	spec, err := ReadSwaggerSpecFromGithub(version)
	utils.DoOrDie(err)
	return spec
}

func DownloadSwaggerSpec(version Version) []byte {
	bytes, err := utils.GetURL(version.SwaggerSpecURL())
	utils.DoOrDie(err)
	return bytes
}

func MakePathFromKubeVersion(version Version) string {
	return fmt.Sprintf("%s/%s-swagger-spec.json", SpecsRootDirectory, version.ToString())
}
