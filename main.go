package main

import "github.com/getkaze/helm/cmd"

var version = "dev"

func main() {
	cmd.Version = version
	cmd.Execute()
}
