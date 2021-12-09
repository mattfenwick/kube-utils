package main

import (
	"github.com/mattfenwick/kube-utils/go/pkg/kubernetes/apiversions"
)

func main() {
	parseJsonSpecs := true
	if parseJsonSpecs {
		apiversions.ParseJsonSpecs()
	} else {
		apiversions.ParseKindResults()
	}
}
