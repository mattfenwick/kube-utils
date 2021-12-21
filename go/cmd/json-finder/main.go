package main

import (
	"github.com/mattfenwick/kube-utils/go/pkg/schema-json"
	"github.com/sirupsen/logrus"
)

func main() {
	if true {
		logrus.SetLevel(logrus.DebugLevel)
		schema_json.RunFindByPath()
	} else {
		schema_json.RunFindByRegex()
	}
}
