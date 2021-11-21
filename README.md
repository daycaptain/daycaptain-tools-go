## TDA

Helper command inspired in [daycaptain tools](https://github.com/daycaptain/tools/) implemented in golang using just 
stdlib, without any third party dependencies.

The goal is to support many architectures without the need to install dependencies.

It's in an early stage of development, so can have some bugs, and definitely some missing features.

### Usage

1. Build from source:
    ```sh
    $ go build ./...
    ```
2. Or download [the latest build](https://github.com/szaffarano/daycaptain-tools-go/releases)

```sh
$ ./tda -h
expected task name, got: []

Usage: tda [options] <task name>

Token can be either specified via the -token flag, via the $DC_API_TOKEN 
environment variable, or via the $DC_API_TOKEN_COMMAND environment variable.
The last option is useful when the token is stored in a command line tool, e.g.

export DC_API_TOKEN_COMMAND="pass some/key"

Options:
  -W	the task to this week
  -d string
    	Add the task to the DATE (formatted by ISO-8601, e.g. 2021-01-31)
  -date string
    	Add the task to the DATE (formatted by ISO-8601, e.g. 2021-01-31)
  -i	(Default)  Add the task to the backlog inbox (default true)
  -inbox
    	(Default)  Add the task to the backlog inbox (default true)
  -m	Add the task to tomorrow's tasks
  -t	Add the task to today's tasks
  -today
    	Add the task to today's tasks
  -token string
    	API key, default $DC_API_TOKEN or $DC_API_TOKEN_COMMAND
  -tomorrow
    	Add the task to tomorrow's tasks
  -version
    	Prints current version and exit
  -w string
    	Add the task to the WEEK (formatted by ISO-8601, e.g. 2021-W07)
  -week string
    	Add the task to the WEEK (formatted by ISO-8601, e.g. 2021-W07)
```

Instead of hard-code the API KEY, you can use the environment variable $DC_API_TOKEN_COMMAND to store the key in a 
command line tool.

```sh
export DC_API_TOKEN_COMMAND="pass some/key"
```

### TODO

- [ ] Organize the code (i.e. move the functions to a separate file).
- [ ] Add code coverage.
- [x] Add git workflow to publish relaases.
