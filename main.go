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
	result, err := tda.Run(version, os.Args[1:])
	if pe, ok := err.(*tda.ParsingError); ok {
		fmt.Fprintln(os.Stderr, pe.Error())
		os.Exit(2)
	} else if err != nil {
		panic(err)
	}
	fmt.Println(result)
}
