package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

const (
	// DayCaptainURL is the base URL for the DayCaptain API
	DayCaptainURL = "https://daycaptain.com/api/"

	// TokenEnvVar is the environment variable name for the token
	TokenEnvVar = "DC_API_TOKEN"
)

var (
	token   string
	version bool
	debug   bool
)

var (
	// Version is the version of the application
	Version = "dev"
)

func init() {
	// initialize the flags
	flag.StringVar(&token, "token", "", "token")
	flag.BoolVar(&version, "version", false, "version")
	flag.BoolVar(&debug, "debug", false, "debug")
}

func main() {
	flag.Usage = help
	flag.Parse()

	if version {
		printVersion()
		return
	}

	token = fromEnv(TokenEnvVar, token)
	if token == "" {
		usage("Token is mandatory")
		return
	}

	if len(flag.Args()) != 1 {
		usage(fmt.Sprintf("%v: expected one single command", strings.Join(flag.Args(), ", ")))
		return
	}

	fmt.Println(flag.Arg(0))
	fmt.Println(token)
}

func printVersion() {
	fmt.Println(Version)
}

func help() {
	usage("")
}

func fromEnv(name, fallback string) string {
	if value, ok := os.LookupEnv(name); ok {
		return value
	}
	return fallback
}

func usage(msg string) {
	if msg != "" {
		fmt.Fprintf(os.Stderr, msg)
		fmt.Fprintf(os.Stderr, "\n\n")
	}
	fmt.Fprintf(os.Stderr, "usage: tda [options] [COMMAND]\n")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, commands())
	os.Exit(2)
}

func commands() string {
	return `
  Commands:
  date      Add the task to the DATE (formatted by ISO-8601, e.g. 2021-01-31
  inbox     Add the task to the backlog inbox
  today     Add the task to today's tasks
  tomorrow  Add the task to tomorrow's tasks
  w         Add the task to this week
  week      Add the task to the WEEK
  `
}
