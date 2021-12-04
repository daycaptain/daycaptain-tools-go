package tda

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/szaffarano/daycaptain-tools-go/daycaptain"
)

const (
	// DayCaptainURLEnvVar is the environment variable name for the DayCaptain URL
	DayCaptainURLEnvVar = "DC_API_URL"

	// TokenEnvVar is the environment variable name for the token
	TokenEnvVar = "DC_API_TOKEN"

	// TokenCmdEnvVar environment variable that holds a command to run to get the token
	TokenCmdEnvVar = "DC_API_TOKEN_COMMAND"

	// DayCaptainURL is the base URL for the DayCaptain API
	DayCaptainURL = "https://daycaptain.com"
)

var (
	token       string
	showVersion bool

	options tdaOptions
)

type tdaOptions struct {
	today    bool
	tomorrow bool
	date     string
	thisWeek bool
	week     string
	inbox    bool
}

// ParsingError is returned by Run when some arguments-related error occurs
type ParsingError struct {
	Message string
}

func (p *ParsingError) Error() string {
	return p.Message
}

func initFlags(args []string) (*flag.FlagSet, error) {
	tda := flag.NewFlagSet("tda", flag.ContinueOnError)

	// initialize the flags
	tda.StringVar(&token, "token", "", "API key, default $DC_API_TOKEN or $DC_API_TOKEN_COMMAND")
	tda.BoolVar(&showVersion, "version", false, "Prints current version and exit")

	tda.BoolVar(&options.today, "t", false, "Add the task to today's tasks")
	tda.BoolVar(&options.today, "today", false, "Add the task to today's tasks")

	tda.BoolVar(&options.tomorrow, "m", false, "Add the task to tomorrow's tasks")
	tda.BoolVar(&options.tomorrow, "tomorrow", false, "Add the task to tomorrow's tasks")

	tda.StringVar(&options.date, "d", "", "Add the task to the DATE (formatted by ISO-8601, e.g. 2021-01-31)")
	tda.StringVar(&options.date, "date", "", "Add the task to the DATE (formatted by ISO-8601, e.g. 2021-01-31)")

	tda.BoolVar(&options.thisWeek, "W", false, "the task to this week")

	tda.StringVar(&options.week, "w", "", "Add the task to the WEEK (formatted by ISO-8601, e.g. 2021-W07)")
	tda.StringVar(&options.week, "week", "", "Add the task to the WEEK (formatted by ISO-8601, e.g. 2021-W07)")

	tda.BoolVar(&options.inbox, "i", true, "(Default)  Add the task to the backlog inbox")
	tda.BoolVar(&options.inbox, "inbox", true, "(Default)  Add the task to the backlog inbox")

	tda.Usage = func() {
		usage(tda, "")
	}

	var output bytes.Buffer
	tda.SetOutput(&output)
	if err := tda.Parse(args); err != nil {
		if !errors.Is(err, flag.ErrHelp) {
			return nil, err
		}
	}

	if len(output.String()) > 0 {
		return tda, &ParsingError{Message: output.String()}
	}

	return tda, nil
}

// Run executes the tda command, returning an error if any occurs or a string
// with the output of the command.
func Run(version string, args []string) (string, error) {
	tda, err := initFlags(args)
	var output bytes.Buffer

	if err != nil {
		return "", err
	}

	tda.SetOutput(&output)

	if showVersion {
		return version, nil
	}

	token = fromEnv(TokenEnvVar, TokenCmdEnvVar, token)
	if token == "" {
		usage(tda, "Token is mandatory")
		return "", &ParsingError{Message: output.String()}
	}

	if len(tda.Args()) != 1 {
		usage(tda, fmt.Sprintf("expected task name, got: %q", tda.Args()))
		return "", &ParsingError{Message: output.String()}
	}

	when, err := options.when()
	if err != nil {
		usage(tda, err.Error())
		return "", &ParsingError{Message: output.String()}
	}

	task := daycaptain.Task{String: tda.Args()[0]}

	url := DayCaptainURL
	if value, ok := os.LookupEnv(DayCaptainURLEnvVar); ok {
		url = value
	}
	client := daycaptain.NewClient(url, token)

	if err := client.NewTask(task, when); err != nil {
		return "", err
	}

	return "OK", nil
}

func (c *tdaOptions) when() (string, error) {
	var when string
	var err error

	if c.today {
		when = daycaptain.FormatDate(time.Now())
	} else if c.tomorrow {
		when = daycaptain.FormatDate(time.Now().AddDate(0, 0, 1))
	} else if c.date != "" {
		when, err = daycaptain.ParseDate(c.date)
		if err != nil {
			return "", err
		}
	} else if c.thisWeek {
		when = daycaptain.FormatWeek(time.Now())
	} else if c.week != "" {
		when, err = daycaptain.ParseWeek(c.week)
		if err != nil {
			return "", err
		}
	}

	return when, nil
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

func usage(tda *flag.FlagSet, msg string) {
	if msg != "" {
		fmt.Fprint(tda.Output(), msg)
		fmt.Fprint(tda.Output(), "\n\n")
	}
	header := `Usage: tda [options] <task name>

Token can be either specified via the -token flag, via the $DC_API_TOKEN 
environment variable, or via the $DC_API_TOKEN_COMMAND environment variable.
The last option is useful when the token is stored in a command line tool, e.g.

export DC_API_TOKEN_COMMAND="pass some/key"

Options:
`

	fmt.Fprint(tda.Output(), header)
	tda.PrintDefaults()
}
