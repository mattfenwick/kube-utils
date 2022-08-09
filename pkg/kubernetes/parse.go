package kubernetes

import (
	"bytes"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"io"
	"io/ioutil"
)

func ParseManyFromFile(path string) ([]interface{}, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to read file %s", path)
	}

	return ParseMany(data)
}

func ParseMany(data []byte) ([]interface{}, error) {
	decoder := yaml.NewDecoder(bytes.NewReader(data))

	var out []interface{}
	for {
		var next interface{}
		err := decoder.Decode(&next)
		if err == io.EOF {
			break
		} else if err != nil {
			return out, errors.Wrapf(err, "unable to decode")
		}
		out = append(out, next)
	}

	return out, nil
}
