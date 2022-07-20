package utils

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io/fs"
	"io/ioutil"
	"os"
	"sigs.k8s.io/yaml"
)

func JsonStringNoIndent(obj interface{}) string {
	bytes, err := json.Marshal(obj)
	DoOrDie(errors.Wrapf(err, "unable to marshal json"))
	return string(bytes)
}

func JsonString(obj interface{}) string {
	bytes, err := json.MarshalIndent(obj, "", "  ")
	DoOrDie(errors.Wrapf(err, "unable to marshal json"))
	return string(bytes)
}

func ParseJson[T any](bs []byte) (*T, error) {
	var t T
	if err := json.Unmarshal(bs, &t); err != nil {
		return nil, errors.Wrapf(err, "unable to unmarshal json")
	}
	return &t, nil
}

func ParseJsonFromFile[T any](path string) (*T, error) {
	bytes, err := ReadFileBytes(path)
	if err != nil {
		return nil, err
	}
	return ParseJson[T](bytes)
}

func ParseYaml[T any](bs []byte) (*T, error) {
	var t T
	if err := yaml.Unmarshal(bs, &t); err != nil {
		return nil, errors.Wrapf(err, "unable to unmarshal yaml")
	}
	return &t, nil
}

func ParseYamlStrict[T any](bs []byte) (*T, error) {
	var t T
	if err := yaml.UnmarshalStrict(bs, &t); err != nil {
		return nil, errors.Wrapf(err, "unable to unmarshal yaml")
	}
	return &t, nil
}

func ParseYamlFromFile[T any](path string) (*T, error) {
	bytes, err := ReadFileBytes(path)
	if err != nil {
		return nil, err
	}
	return ParseYaml[T](bytes)
}

func ParseYamlFromFileStrict[T any](path string) (*T, error) {
	bytes, err := ReadFileBytes(path)
	if err != nil {
		return nil, err
	}
	return ParseYamlStrict[T](bytes)
}

func YamlString(obj interface{}) string {
	bytes, err := yaml.Marshal(obj)
	DoOrDie(errors.Wrapf(err, "unable to marshal yaml"))
	return string(bytes)
}

func PrintJson(obj interface{}) {
	fmt.Printf("%s\n", JsonString(obj))
}

func WriteJsonToFile(obj interface{}, path string) error {
	content := JsonString(obj)
	return WriteFile(path, content, 0644)
}

func DoesFileExist(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	} else if errors.Is(err, os.ErrNotExist) {
		return false
	} else {
		panic(errors.Wrapf(err, "unable to determine if file %s exists", path))
	}
}

// WriteFile wraps calls to ioutil.WriteFile, ensuring that errors are wrapped in a stack trace
func WriteFile(filename string, contents string, perm fs.FileMode) error {
	return errors.Wrapf(ioutil.WriteFile(filename, []byte(contents), perm), "unable to write file %s", filename)
}

// WriteFileBytes wraps calls to ioutil.WriteFile, ensuring that errors are wrapped in a stack trace
func WriteFileBytes(filename string, bytes []byte, perm fs.FileMode) error {
	return errors.Wrapf(ioutil.WriteFile(filename, bytes, perm), "unable to write file %s", filename)
}

// ReadFile wraps calls to ioutil.ReadFile, ensuring that errors are wrapped in a stack trace
func ReadFile(filename string) (string, error) {
	bytes, err := ioutil.ReadFile(filename)
	return string(bytes), errors.Wrapf(err, "unable to read file %s", filename)
}

// ReadFileBytes wraps calls to ioutil.ReadFile, ensuring that errors are wrapped in a stack trace
func ReadFileBytes(filename string) ([]byte, error) {
	bytes, err := ioutil.ReadFile(filename)
	return bytes, errors.Wrapf(err, "unable to read file %s", filename)
}

func FileExists(path string) (bool, error) {
	if _, err := os.Stat(path); err == nil {
		return true, nil
	} else if errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else {
		return false, errors.Wrapf(err, "unable to os.Stat path %s", path)
	}
}
