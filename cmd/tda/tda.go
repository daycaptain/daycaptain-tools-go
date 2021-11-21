package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

const (
	// DayCaptainURL is the base URL for the DayCaptain API
	DayCaptainURL = "https://daycaptain.com"

	// TokenEnvVar is the environment variable name for the token
	TokenEnvVar = "DC_API_TOKEN"

	// TokenCmdEnvVar environment variable that holds a command to run to get the token
	TokenCmdEnvVar = "DC_API_TOKEN_COMMAND"
)

var (
	token       string
	showVersion bool

	tdaConfig TdaConfig
)

var (
	// version is the version of the application
	version = "dev"
)

type Task struct {
	String string `json:"string"`
}

type TdaConfig struct {
	today    bool
	tomorrow bool
	date     string
	thisWeek bool
	week     string
	inbox    bool
}

type Request struct {
	url string
}

func (c *TdaConfig) GetCommand() (Request, error) {
	optionsSet := make([]Request, 0)
	fallback := Request{fmt.Sprintf("%s/backlog-items", DayCaptainURL)}

	if c.today {
		now := time.Now().Format("2006-01-02")
		optionsSet = append(optionsSet, Request{fmt.Sprintf("%s/%s/tasks", DayCaptainURL, now)})
	}

	if c.tomorrow {
		tomorrow := time.Now().AddDate(0, 0, 1).Format("2006-01-02")
		optionsSet = append(optionsSet, Request{fmt.Sprintf("%s/%s/tasks", DayCaptainURL, tomorrow)})
	}

	if c.date != "" {
		current, err := time.Parse("2006-01-02", c.date)
		if err != nil {
			return Request{}, err
		}
		optionsSet = append(optionsSet, Request{fmt.Sprintf("%s/%s/tasks", DayCaptainURL, current.Format("2006-01-02"))})
	}

	if c.thisWeek {
		year, isoWeek := time.Now().ISOWeek()
		optionsSet = append(optionsSet, Request{fmt.Sprintf("%s/%d-W%d/tasks", DayCaptainURL, year, isoWeek)})
	}

	if c.week != "" {
		r, err := regexp.Compile(`^\d{4}-W\d{2}$`)
		if err != nil {
			panic(err)
		}

		if !r.Match([]byte(c.week)) {
			return Request{}, fmt.Errorf("invalid week format: %s", c.week)
		}

		optionsSet = append(optionsSet, Request{fmt.Sprintf("%s/%s/tasks", DayCaptainURL, c.week)})
	}

	switch len(optionsSet) {
	case 0:
		return fallback, nil
	case 1:
		return optionsSet[0], nil
	default:
		return Request{}, fmt.Errorf("More than one option specified: %v", optionsSet)
	}
}

func init() {
	// initialize the flags
	flag.StringVar(&token, "token", "", "API key, default $DC_API_TOKEN or $DC_API_TOKEN_COMMAND")
	flag.BoolVar(&showVersion, "version", false, "Prints current version and exit")

	flag.BoolVar(&tdaConfig.today, "t", false, "Add the task to today's tasks")
	flag.BoolVar(&tdaConfig.today, "today", false, "Add the task to today's tasks")

	flag.BoolVar(&tdaConfig.tomorrow, "m", false, "Add the task to tomorrow's tasks")
	flag.BoolVar(&tdaConfig.tomorrow, "tomorrow", false, "Add the task to tomorrow's tasks")

	flag.StringVar(&tdaConfig.date, "d", "", "Add the task to the DATE (formatted by ISO-8601, e.g. 2021-01-31)")
	flag.StringVar(&tdaConfig.date, "date", "", "Add the task to the DATE (formatted by ISO-8601, e.g. 2021-01-31)")

	flag.BoolVar(&tdaConfig.thisWeek, "W", false, "the task to this week")

	flag.StringVar(&tdaConfig.week, "w", "", "Add the task to the WEEK (formatted by ISO-8601, e.g. 2021-W07)")
	flag.StringVar(&tdaConfig.week, "week", "", "Add the task to the WEEK (formatted by ISO-8601, e.g. 2021-W07)")

	flag.BoolVar(&tdaConfig.inbox, "i", true, "(Default)  Add the task to the backlog inbox")
	flag.BoolVar(&tdaConfig.inbox, "inbox", true, "(Default)  Add the task to the backlog inbox")
}

func main() {
	flag.Usage = help
	flag.Parse()

	if showVersion {
		printVersion()
		return
	}

	token = fromEnv(TokenEnvVar, TokenCmdEnvVar, token)
	if token == "" {
		usage("Token is mandatory")
		return
	}

	if len(flag.Args()) != 1 {
		usage(fmt.Sprintf("expected task name, got: %q", flag.Args()))
		return
	}

	cmd, err := tdaConfig.GetCommand()
	if err != nil {
		usage(err.Error())
		return
	}

	body, err := json.Marshal(Task{String: flag.Arg(0)})
	if err != nil {
		usage(err.Error())
		return
	}

	req, err := http.NewRequest("POST", cmd.url, bytes.NewReader(body))
	if err != nil {
		usage(err.Error())
		return
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Add("Content-Type", "application/json")

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		usage(err.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			usage(err.Error())
			return
		}

		usage(fmt.Sprintf("%s: %s", resp.Status, string(body)))
		return
	}

	fmt.Println("OK")
}

func printVersion() {
	fmt.Println(version)
}

func help() {
	usage("")
}

func fromEnv(name, nameCmd, value string) string {
	if value != "" {
		return value
	} else if envValue, ok := os.LookupEnv(name); ok {
		return envValue
	} else if cmd, ok := os.LookupEnv(nameCmd); ok {
		command := strings.Split(cmd, " ")
		if output, err := exec.Command(command[0], command[1:]...).Output(); err == nil {
			// remove trailing newline
			return strings.Replace(string(output), "\n", "", 1)
		}
	}
	return value
}

func usage(msg string) {
	if msg != "" {
		fmt.Fprintf(os.Stderr, msg)
		fmt.Fprintf(os.Stderr, "\n\n")
	}
	header := `Usage: tda [options] <task name>

Token can be either specified via the -token flag, via the $DC_API_TOKEN 
environment variable, or via the $DC_API_TOKEN_COMMAND environment variable.
The last option is useful when the token is stored in a command line tool, e.g.

export DC_API_TOKEN_COMMAND="pass some/key"

Options:
`

	fmt.Fprintf(os.Stderr, header)
	flag.PrintDefaults()
	os.Exit(2)
}
