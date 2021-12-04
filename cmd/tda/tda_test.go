package tda

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTask(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.Header)
		w.WriteHeader(http.StatusCreated)
	}))
	defer mockServer.Close()

	assert.NoError(t, os.Setenv(DayCaptainURLEnvVar, mockServer.URL))
	assert.NoError(t, os.Setenv(TokenEnvVar, "token"))

	testCases := []struct {
		name        string
		args        []string
		want        string
		errContains string
	}{
		{"backlog task", []string{"hello world"}, "OK", ""},
		{"today task", []string{"-t", "hello world"}, "OK", ""},
		{"tomorrow task", []string{"-tomorrow", "hello world"}, "OK", ""},
		{"this week task", []string{"-W", "hello world"}, "OK", ""},
		{"other wee task", []string{"-w", "2021-W1", "hello world"}, "OK", ""},
		{"other day task", []string{"-date", "2021-01-10", "hello world"}, "OK", ""},
		{"help", []string{"--help"}, "", "Usage"},
		{"version", []string{"-version"}, "1.0.0", ""},
		{"invalid arguments", []string{"-unknown"}, "", "flag provided but not defined: -unknown"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Run("1.0.0", tc.args)

			assert.Equal(t, tc.want, got)
			if tc.errContains != "" && assert.Error(t, err) {
				assert.Contains(t, err.Error(), tc.errContains)
			}
		})
	}
}
