package main

import (
	"github.com/mattfenwick/kube-utils/go/pkg/kubernetes"
	"os"
)

func main() {
	path := os.Args[1]

	//kubernetes.Run(path)
	kubernetes.RunAnalyzeExample(path)
}