package main

import (
	"fmt"
	"os"

	"github.com/szaffarano/daycaptain-tools-go/cmd/tda"
)

var (
	// version is the version of the application
	version = "dev"
)

func main() {
	err := tda.Run(version, os.Args[1:])
	if pe, ok := err.(*tda.ParsingError); ok {
		fmt.Fprintln(os.Stderr, pe.Error())
	} else if err != nil {
		panic(err)
	}
}
