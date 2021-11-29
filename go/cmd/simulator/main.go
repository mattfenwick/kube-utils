package main

import (
	"github.com/mattfenwick/kube-utils/go/pkg/simulator"
	"os"
)

func main() {
	if os.Args[1] == "server" {
		server()
	} else {
		client()
	}
}

func server() {
	simulator.RunServer()
}

func client() {
	serverAddress := "http://localhost:19999"
	if len(os.Args) >= 3 {
		serverAddress = os.Args[2]
	}
	simulator.RunClient(serverAddress)
}
