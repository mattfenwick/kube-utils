package swagger

import (
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"sort"
)

func SetUpLogger(logLevelStr string) error {
	logLevel, err := logrus.ParseLevel(logLevelStr)
	if err != nil {
		return errors.Wrapf(err, "unable to parse the specified log level: '%s'", logLevel)
	}
	logrus.SetLevel(logLevel)
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	logrus.Infof("log level set to '%s'", logrus.GetLevel())
	return nil
}

func SortedKeys(dict map[string]interface{}) []string {
	var keys []string
	for k := range dict {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
