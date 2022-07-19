package utils

import (
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"os"
)

func DoOrDie(err error) {
	if err != nil {
		logrus.Fatalf("%+v", err)
	}
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

func FileExists(path string) (bool, error) {
	if _, err := os.Stat(path); err == nil {
		return true, nil
	} else if errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else {
		return false, errors.Wrapf(err, "unable to os.Stat path %s", path)
	}
}
