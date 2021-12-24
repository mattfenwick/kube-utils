package utils

import (
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
)

func DoOrDie(err error) {
	if err != nil {
		logrus.Fatalf("%+v", err)
	}
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

func Pointer(s string) *string {
	return &s
}

func AddKey(dict map[string]bool, key string) map[string]bool {
	out := map[string]bool{}
	for k, v := range dict {
		out[k] = v
	}
	out[key] = true
	return out
}

func StringPrefix(s string, chars int) string {
	if len(s) <= chars {
		return s
	}
	return s[:chars]
}

func Set(xs []string) map[string]bool {
	out := map[string]bool{}
	for _, x := range xs {
		out[x] = true
	}
	return out
}

func GetFileFromURL(url string, path string) error {
	response, err := http.Get(url)
	if err != nil {
		return errors.Wrapf(err, "unable to GET %s", url)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode > 299 {
		return errors.Errorf("GET request to %s failed with status code %d", url, response.StatusCode)
	}
	bytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return errors.Wrapf(err, "unable to read body from GET to %s", url)
	}

	return errors.Wrapf(ioutil.WriteFile(path, bytes, 0777), "unable to write file %s", path)
}
