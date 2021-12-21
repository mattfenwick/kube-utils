package utils

import (
	"bytes"
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
	bs, err := ioutil.ReadFile(path)
	if err != nil {
		return errors.Wrapf(err, "unable to read file %s", path)
	}

	err = json.Unmarshal(bs, obj)
	return errors.Wrapf(err, "unable to unmarshal json")
}

// Marshal is a stand-in for json.Marshal which *does not escape HTML*
func Marshal(t interface{}) ([]byte, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(t)
	return buffer.Bytes(), errors.Wrapf(err, "unable to encode json")
}

// MarshalIndent is a stand-in for json.MarshalIndent which *does not escape HTML*
func MarshalIndent(v interface{}, prefix, indent string) ([]byte, error) {
	bs, err := Marshal(v)
	if err != nil {
		return nil, err
	}
	var destinationBuffer bytes.Buffer
	err = json.Indent(&destinationBuffer, bs, prefix, indent)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to indent json")
	}
	return destinationBuffer.Bytes(), nil
}

func WriteJson(path string, obj interface{}) error {
	bs, err := MarshalIndent(obj, "", "  ")
	if err != nil {
		return errors.Wrapf(err, "unable to marshal json")
	}

	err = ioutil.WriteFile(path, bs, 0644)
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

func MapFilterEmptyValues(dict map[string]string) map[string]string {
	out := map[string]string{}
	for k, v := range dict {
		if v != "" {
			out[k] = v
		}
	}
	return out
}
