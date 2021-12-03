package daycaptain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

const (

	// ISO8601 is the date format supported by DayCaptain
	ISO8601 = "2006-01-02"
)

var (
	isoWeekRegex *regexp.Regexp
)

// Task is the DayCaptain task
type Task struct {
	String string `json:"string"`
}

// DayCaptain is the client for the DayCaptain API
type DayCaptain struct {
	url   string
	token string
}

func init() {
	var err error
	isoWeekRegex, err = regexp.Compile(`^(\d{4})-W(\d{1,2})$`)
	if err != nil {
		panic(err)
	}
}

// NewClient returns a new DayCaptain client
func NewClient(url string, token string) *DayCaptain {
	return &DayCaptain{url: url, token: token}
}

// NewTask creates a new task
func (dc *DayCaptain) NewTask(task Task, when string) error {
	body, err := json.Marshal(task)
	if err != nil {
		return err
	}

	var url string
	if when == "" {
		url = fmt.Sprintf("%s/backlog-items", dc.url)
	} else {
		url = fmt.Sprintf("%s/%s/tasks", dc.url, when)
	}
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))

	if err != nil {
		return err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", dc.token))
	req.Header.Add("Content-Type", "application/json")

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		return fmt.Errorf(fmt.Sprintf("%s: %s", resp.Status, string(body)))
	}

	return nil
}

// FormatDate formats a given date in the ISO-8601 format, i.e. 2021-01-01
func FormatDate(t time.Time) string {
	return t.Format(ISO8601)
}

// FormatWeek formats a given date in the ISO-8601 week format, i.e. 2021-W33
func FormatWeek(t time.Time) string {
	year, week := t.ISOWeek()
	return fmt.Sprintf("%d-W%d", year, week)
}

// ParseDate parses a date in the ISO-8601 format, i.e. 2021-01-01
func ParseDate(d string) (string, error) {
	parsed, err := time.Parse("2006-01-02", d)
	if err != nil {
		return "", err
	}
	return FormatDate(parsed), nil
}

// ParseWeek parses a week number in the ISO-8601 format, i.e. 2021-W33
func ParseWeek(w string) (string, error) {
	parsed := isoWeekRegex.FindStringSubmatch(w)
	// if !isoWeekRegex.Match([]byte(w)) {
	if len(parsed) != 3 {
		return "", fmt.Errorf("invalid ISO week format: %s", w)
	}

	year, err := strconv.Atoi(parsed[1])
	if err != nil {
		return "", fmt.Errorf("unexpected error: %v", err)
	}
	week, err := strconv.Atoi(parsed[2])
	if err != nil {
		return "", fmt.Errorf("unexpected error: %v", err)
	}

	if year < 2020 {
		return "", fmt.Errorf("year must be >= 2020")
	}

	if week < 1 || week > 53 {
		return "", fmt.Errorf("week must be between 1 and 53")
	}

	return w, nil

}
