package main

import (
	"flag"

	"github.com/kszab0/revolver"
)

func main() {
	configPath := flag.String("c", "revolver.yml", "Path to config file")
	flag.Parse()

	config, err := revolver.ParseConfigFile(*configPath)
	if err != nil {
		panic(err)
	}

	revolver.Watch(*config)
}
