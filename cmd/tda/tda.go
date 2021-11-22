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

	ISO8601 = "2006-01-02"
)

var (
	token       string
	showVersion bool

	options tdaOptions
)

var (
	// version is the version of the application
	version      = "dev"
	isoWeekRegex *regexp.Regexp
)

type Task struct {
	String string `json:"string"`
}

type tdaOptions struct {
	today    bool
	tomorrow bool
	date     string
	thisWeek bool
	week     string
	inbox    bool
}

type request struct {
	url    string
	method string
}

func post(path string) *request {
	return &request{fmt.Sprintf("%s/%s", DayCaptainURL, path), "POST"}
}

func (c *tdaOptions) getCommand() (*request, error) {
	options := make([]*request, 0)
	fallback := post(fmt.Sprintf("%s/backlog-items", DayCaptainURL))

	if c.today {
		today := time.Now().Format(ISO8601)
		options = append(options, post(fmt.Sprintf("%s/tasks", today)))
	}

	if c.tomorrow {
		tomorrow := time.Now().AddDate(0, 0, 1).Format(ISO8601)
		options = append(options, post(fmt.Sprintf("%s/tasks", tomorrow)))
	}

	if c.date != "" {
		current, err := time.Parse("2006-01-02", c.date)
		if err != nil {
			return nil, err
		}
		options = append(options, post(fmt.Sprintf("%s/tasks", current.Format(ISO8601))))
	}

	if c.thisWeek {
		year, week := time.Now().ISOWeek()
		options = append(options, post(fmt.Sprintf("%d-W%d/tasks", year, week)))
	}

	if c.week != "" {
		if !isoWeekRegex.Match([]byte(c.week)) {
			return nil, fmt.Errorf("invalid ISO week format: %s", c.week)
		}

		options = append(options, post(fmt.Sprintf("%s/%s/tasks", DayCaptainURL, c.week)))
	}

	switch len(options) {
	case 0:
		return fallback, nil
	case 1:
		return options[0], nil
	default:
		return nil, fmt.Errorf("More than one option specified: %v", options)
	}
}

func init() {
	var err error

	// initialize the flags
	flag.StringVar(&token, "token", "", "API key, default $DC_API_TOKEN or $DC_API_TOKEN_COMMAND")
	flag.BoolVar(&showVersion, "version", false, "Prints current version and exit")

	flag.BoolVar(&options.today, "t", false, "Add the task to today's tasks")
	flag.BoolVar(&options.today, "today", false, "Add the task to today's tasks")

	flag.BoolVar(&options.tomorrow, "m", false, "Add the task to tomorrow's tasks")
	flag.BoolVar(&options.tomorrow, "tomorrow", false, "Add the task to tomorrow's tasks")

	flag.StringVar(&options.date, "d", "", "Add the task to the DATE (formatted by ISO-8601, e.g. 2021-01-31)")
	flag.StringVar(&options.date, "date", "", "Add the task to the DATE (formatted by ISO-8601, e.g. 2021-01-31)")

	flag.BoolVar(&options.thisWeek, "W", false, "the task to this week")

	flag.StringVar(&options.week, "w", "", "Add the task to the WEEK (formatted by ISO-8601, e.g. 2021-W07)")
	flag.StringVar(&options.week, "week", "", "Add the task to the WEEK (formatted by ISO-8601, e.g. 2021-W07)")

	flag.BoolVar(&options.inbox, "i", true, "(Default)  Add the task to the backlog inbox")
	flag.BoolVar(&options.inbox, "inbox", true, "(Default)  Add the task to the backlog inbox")

	isoWeekRegex, err = regexp.Compile(`^\d{4}-W\d{2}$`)
	if err != nil {
		panic(err)
	}
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

	cmd, err := options.getCommand()
	if err != nil {
		usage(err.Error())
		return
	}

	task := Task{flag.Args()[0]}
	body, err := json.Marshal(task)
	if err != nil {
		usage(err.Error())
		return
	}

	req, err := http.NewRequest(cmd.method, cmd.url, bytes.NewReader(body))
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
		fmt.Fprint(os.Stderr, msg)
		fmt.Fprint(os.Stderr, "\n\n")
	}
	header := `Usage: tda [options] <task name>

Token can be either specified via the -token flag, via the $DC_API_TOKEN 
environment variable, or via the $DC_API_TOKEN_COMMAND environment variable.
The last option is useful when the token is stored in a command line tool, e.g.

export DC_API_TOKEN_COMMAND="pass some/key"

Options:
`

	fmt.Fprint(os.Stderr, header)
	flag.PrintDefaults()
	os.Exit(2)
}
