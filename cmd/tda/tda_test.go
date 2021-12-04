package tda

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	dc "github.com/szaffarano/daycaptain-tools-go/daycaptain"
)

type request struct {
	url        string
	authHeader string
}

func TestNewTask(t *testing.T) {
	req := make(chan request, 1)

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req <- request{r.URL.String(), r.Header.Get("Authorization")}
		w.WriteHeader(http.StatusCreated)
	}))
	defer mockServer.Close()

	assert.NoError(t, os.Setenv(DayCaptainURLEnvVar, mockServer.URL))
	assert.NoError(t, os.Setenv(TokenEnvVar, "token"))

	testCases := []struct {
		name string
		args []string

		wantOutput     string
		wantURL        string
		wantAuthHeader string
		errContains    string
	}{
		{
			"backlog task with default flag",
			[]string{"hello world"},
			"OK",
			"/backlog-items",
			"Bearer token",
			"",
		},
		{
			"backlog task",
			[]string{"-i", "hello world"},
			"OK",
			"/backlog-items",
			"Bearer token",
			"",
		},
		{
			"today task",
			[]string{"-t", "hello world"},
			"OK",
			fmt.Sprintf("/%s/tasks", dc.FormatDate(time.Now())),
			"Bearer token",
			"",
		},
		{
			"tomorrow task",
			[]string{"-tomorrow", "hello world"},
			"OK",
			fmt.Sprintf("/%s/tasks", dc.FormatDate(time.Now().Add(time.Hour*24))),
			"Bearer token",
			"",
		},
		{
			"this week task",
			[]string{"-W", "hello world"},
			"OK",
			fmt.Sprintf("/%s/tasks", dc.FormatWeek(time.Now())),
			"Bearer token",
			"",
		},
		{
			"other week task",
			[]string{"-w", "2021-W1", "hello world"},
			"OK",
			"/2021-W1/tasks",
			"Bearer token",
			"",
		},
		{
			"more than one flag",
			[]string{"-tomorrow", "-date", "2021-01-10", "hello world"},
			"",
			"",
			"",
			"Only one of the following flags can be specified: date, tomorrow",
		},
		{
			"extra flags",
			[]string{"-tomorrow", "something", "-date", "2021-01-10", "hello world"},
			"",
			"",
			"",
			"expected task name, got",
		},

		{
			"other day task custom token",
			[]string{"-token", "newToken", "-date", "2021-01-10", "hello world"},
			"OK",
			"/2021-01-10/tasks",
			"Bearer newToken",
			"",
		},
		{
			"other day task",
			[]string{"-date", "2021-01-10", "hello world"},
			"OK",
			"/2021-01-10/tasks",
			"Bearer token",
			"",
		},
		{
			"help",
			[]string{"--help"},
			"",
			"",
			"",
			"Usage",
		},
		{
			"version",
			[]string{"-version"},
			"1.0.0",
			"",
			"",
			"",
		},
		{
			"invalid arguments",
			[]string{"-unknown"},
			"",
			"",
			"",
			"flag provided but not defined: -unknown",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Run("1.0.0", tc.args)

			var actual request
			select {
			case actual = <-req:
				assert.Equal(t, actual.url, tc.wantURL)
			case <-time.After(time.Millisecond * 100):
				assert.Equal(t, "", tc.wantURL)
			}

			assert.Equal(t, tc.wantOutput, got)
			assert.Equal(t, tc.wantAuthHeader, actual.authHeader)
			if tc.errContains != "" && assert.Error(t, err) {
				assert.Contains(t, err.Error(), tc.errContains)
			}
		})
	}
}
