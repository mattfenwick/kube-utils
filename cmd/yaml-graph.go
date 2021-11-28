package main

import "github.com/mattfenwick/kube-utils/pkg/kubernetes"

func main() {
	path := "yaml-example-source.yaml"

	//kubernetes.Run(path)
	kubernetes.RunAnalyzeExample(path)
}
