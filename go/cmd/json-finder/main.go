package main

import (
	"github.com/mattfenwick/kube-utils/go/pkg/json"
	"github.com/sirupsen/logrus"
)

func main() {
	if true {
		logrus.SetLevel(logrus.DebugLevel)
		json.RunFindByPath()
	} else {
		json.RunFindByRegex()
	}
}
