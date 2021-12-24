package utils

import (
	"bytes"
	"encoding/json"
	"github.com/pkg/errors"
	"io/ioutil"
)

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

func JsonRemarshal(obj interface{}) (interface{}, error) {
	bs, err := MarshalIndent(obj, "", "  ")
	if err != nil {
		return "", err
	}

	var out interface{}
	err = json.Unmarshal(bs, &out)
	return out, errors.Wrapf(err, "unable to unmarshal json")
}

func MustJsonRemarshal(obj interface{}) interface{} {
	out, err := JsonRemarshal(obj)
	DoOrDie(err)
	return out
}

func WriteJson(path string, obj interface{}) error {
	bs, err := MarshalIndent(obj, "", "  ")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(path, bs, 0644)
	return errors.Wrapf(err, "unable to write file %s", path)
}

// JsonUnmarshalMarshal is used to get keys from a struct into a sorted order.
//   See https://stackoverflow.com/a/61887446/894284.
//   Apparently, golang's json library sorts keys from
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
