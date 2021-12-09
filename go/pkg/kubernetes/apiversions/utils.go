package apiversions

import (
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
)

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
