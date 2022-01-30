package api

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	logging "github.com/blainemoser/Logging"
	"github.com/blainemoser/todobot/event"
	"github.com/blainemoser/todobot/tests"
	"github.com/blainemoser/todobot/testsuite"
)

const testSlackChallenge = `{
    "token": "Jhj5dZrVaK7ZwHHjRyZWjbDl",
    "challenge": "3eZbrw1aBm2rZgRNFdxV2595E9CY3gmdALWMmHkvFXO7tYXAYM8P",
    "type": "url_verification"
}`

const testChannelName = "CH0000001"

var (
	port  int = 9000
	a     *Api
	l     *logging.Log
	suite *testsuite.TestSuite
)

func TestMain(m *testing.M) {
	var err error
	suite, err = testsuite.Initialize("api")
	if err != nil {
		panic(err)
	}
	defer suite.TearDown()
	err = getAPI()
	if err != nil {
		panic(err)
	}
	suite.ResultCode = m.Run()
}

func getAPI() (err error) {
	l, err = testsuite.TestLogger()
	if err != nil {
		return err
	}
	a = Boot(port, suite.TestEnv["slackURL"], suite.TestDatabase, l)
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

func modifiedTestPayload() *strings.Reader {
	return strings.NewReader(
		strings.Replace(tests.TestEventPayload, "C02NLG80TEH", testChannelName, 1),
	)
}

func TestSlackEvent(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/slack-event", modifiedTestPayload())
	w := httptest.NewRecorder()
	a.slackEvent(w, req)
	res := w.Result()
	_, err := testsuite.GetBody(res)
	if err != nil {
		t.Fatal(err)
	}
	if event.Queue == nil {
		t.Fatalf("event queue is empty")
	}
	e := event.Queue.Back()
	if e == nil || e.Value == nil {
		t.Fatalf("api queue has no events")
	}
	event, ok := e.Value.(*event.Event)
	if !ok {
		t.Fatalf("event is not of the event type; type assertion failed")
	}
	if event.Channel != testChannelName {
		t.Fatalf("expected event channel to be '%s', got '%s'", testChannelName, event.Channel)
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
