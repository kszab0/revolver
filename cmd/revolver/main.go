package main

import (
	"os"

	"github.com/kszab0/revolver"
)

func main() {
	config, err := revolver.ParseFlags(os.Args)
	if err != nil {
		panic(err)
	}
	revolver.Watch(*config)
}
