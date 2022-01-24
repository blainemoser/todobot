package api

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	logging "github.com/blainemoser/Logging"
	"github.com/blainemoser/todobot/testsuite"
)

const testSlackChallenge = `{
    "token": "Jhj5dZrVaK7ZwHHjRyZWjbDl",
    "challenge": "3eZbrw1aBm2rZgRNFdxV2595E9CY3gmdALWMmHkvFXO7tYXAYM8P",
    "type": "url_verification"
}`

var (
	port int = 9000
	a    *Api
	l    *logging.Log
	code int
)

func TestMain(m *testing.M) {
	var err error
	err = getAPI()
	if err != nil {
		panic(err)
	}
	code = m.Run()
}

func getAPI() (err error) {
	l, err = testsuite.TestLogger()
	if err != nil {
		return err
	}
	a = Boot(port, l)
	go func() {
		err = a.Run()
	}()
	return err
}

func TestPing(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()
	a.ping(w, req)
	res := w.Result()
	data, err := testsuite.GetBody(res)
	if err != nil {
		t.Fatal(err)
	}
	err = testsuite.EvaluateResult(data, map[string]interface{}{
		"error":   false,
		"message": "pong",
	})
	if err != nil {
		t.Error(err)
	}
}

func TestSlackEvent(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/slack-event", nil)
	w := httptest.NewRecorder()
	a.slackEvent(w, req)
	res := w.Result()
	_, err := testsuite.GetBody(res)
	if err != nil {
		t.Fatal(err)
	}
}

func TestSlackChallenge(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/slack-event", strings.NewReader(testSlackChallenge))
	w := httptest.NewRecorder()
	a.slackEvent(w, req)
	res := w.Result()
	data, err := testsuite.GetBody(res)
	if err != nil {
		t.Fatal(err)
	}
	err = testsuite.EvaluateResult(data, map[string]interface{}{
		"challenge": "3eZbrw1aBm2rZgRNFdxV2595E9CY3gmdALWMmHkvFXO7tYXAYM8P",
	})
	if err != nil {
		t.Error(err)
	}
}

func TestError(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/slack-event", nil)
	w := httptest.NewRecorder()
	a.slackEvent(w, req)
	res := w.Result()
	defer res.Body.Close()
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	err = testsuite.EvaluateResult(data, map[string]interface{}{
		"error":   true,
		"message": fmt.Sprintf("method '%s' is not allowed", http.MethodGet),
	})
	if err != nil {
		t.Error(err)
	}
}
