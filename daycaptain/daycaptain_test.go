package daycaptain

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type request struct {
	body string
	url  string
	err  error
}

func TestNewTask(t *testing.T) {
	reqCh := make(chan request, 1)
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if strings.Contains(string(body), "500") {
			w.WriteHeader(500)
		}
		w.Write([]byte("OK"))
		reqCh <- request{body: string(body), err: err, url: r.URL.String()}
	}))
	defer mockServer.Close()

	dc := NewClient(mockServer.URL, "test-token")

	jsonTask := `{"string":"test task"}`
	jsonTask500 := `{"string":"test task 500"}`
	testCases := []struct {
		name     string
		task     Task
		when     string
		expected request
	}{
		{"new backlog task", Task{"test task"}, "", request{jsonTask, "/backlog-items", nil}},
		{"day task", Task{"test task"}, "2021-01-01", request{jsonTask, "/2021-01-01/tasks", nil}},
		{"week task", Task{"test task"}, "2021-W11", request{jsonTask, "/2021-W11/tasks", nil}},
		{"internal server error", Task{"test task 500"}, "2021-W11", request{jsonTask500, "/2021-W11/tasks", fmt.Errorf("500 Internal Server Error: OK")}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := dc.NewTask(tc.task, tc.when)
			select {
			case <-time.After(time.Millisecond * 250):
				t.Error("Timeout")
			case r := <-reqCh:
				assert.Nil(t, r.err)

				assert.Equal(t, tc.expected.body, r.body)
				assert.Equal(t, tc.expected.url, r.url, "/backlog-items")
				assert.Equal(t, tc.expected.body, r.body, `{"string":"test task"}`)
				if tc.expected.err != nil {
					if assert.NotNil(t, err) {
						assert.Equal(t, tc.expected.err.Error(), err.Error())
					}
				}
				assert.Equal(t, tc.expected.err, err)
			}
		})
	}

	t.Run("test format date", func(t *testing.T) {
		d, err := time.Parse("2006-01-02", "2021-01-01")
		assert.NoError(t, err)
		assert.Equal(t, "2021-01-01", FormatDate(d))
	})

	t.Run("test format week", func(t *testing.T) {
		d, err := time.Parse("2006-01-02", "2021-01-05")
		assert.NoError(t, err)
		assert.Equal(t, "2021-W1", FormatWeek(d))
	})

	t.Run("parse valid date", func(t *testing.T) {
		d, err := ParseDate("2021-01-05")
		assert.NoError(t, err)
		assert.Equal(t, "2021-01-05", d)
	})

	t.Run("parse invalid valid date", func(t *testing.T) {
		d, err := ParseDate("A-2021-01-05")
		assert.Error(t, err)
		assert.Equal(t, "", d)
	})

	t.Run("parse valid week", func(t *testing.T) {
		d, err := ParseWeek("2021-W1")
		assert.NoError(t, err)
		assert.Equal(t, "2021-W1", d)
	})

	t.Run("parse invalid year", func(t *testing.T) {
		d, err := ParseWeek("1921-W02")
		assert.Error(t, err)
		assert.Equal(t, "", d)
	})

	t.Run("parse week format", func(t *testing.T) {
		d, err := ParseWeek("a-b-c")
		assert.Error(t, err)
		assert.Equal(t, "", d)
	})

	t.Run("parse invalid week", func(t *testing.T) {
		d, err := ParseWeek("2021-W88")
		assert.Error(t, err)
		assert.Equal(t, "", d)
	})
}
