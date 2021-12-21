package utils

import (
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io/ioutil"
)

func DoOrDie(err error) {
	if err != nil {
		logrus.Fatalf("%+v", err)
	}
}

func ReadJson(path string, obj interface{}) error {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return errors.Wrapf(err, "unable to read file %s", path)
	}

	err = json.Unmarshal(bytes, obj)
	return errors.Wrapf(err, "unable to unmarshal json")
}

func WriteJson(path string, obj interface{}) error {
	bytes, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return errors.Wrapf(err, "unable to marshal json")
	}

	err = ioutil.WriteFile(path, bytes, 0644)
	return errors.Wrapf(err, "unable to write file %s", path)
}

// JsonUnmarshalMarshal is used to get keys from a struct into a sorted order.
//   See JsonUnmarshalMarshal.  Apparently, golang's json library sorts keys from
//   maps, but NOT from structs.  So this function works by reading json into a
//   generic structure of maps, then marshaling back into completely sorted json.
func JsonUnmarshalMarshal(path string) error {
	obj := map[string]interface{}{}
	err := ReadJson(path, &obj)
	if err != nil {
		return err
	}
	return WriteJson(path, obj)
}

func MapKeys(dict map[string]interface{}) []string {
	var keys []string
	for key := range dict {
		keys = append(keys, key)
	}
	return keys
}
