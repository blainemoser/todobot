package slackapi

import (
	"testing"

	logging "github.com/blainemoser/Logging"
	"github.com/blainemoser/todobot/testsuite"
	"github.com/blainemoser/todobot/user"
)

var (
	port      int = 9000
	l         *logging.Log
	suite     *testsuite.TestSuite
	timestamp int64 = 1643505910
)

func TestMain(m *testing.M) {
	var err error
	suite, err = testsuite.Initialize("slackapi")
	if err != nil {
		panic(err)
	}
	defer suite.TearDown()
	suite.ResultCode = m.Run()
}

func TestGetUser(t *testing.T) {
	result, err := Slack(suite.TestDatabase, suite.TestEnv["slackToken"]).GetUserDetails(suite.TestEnv["testUser"])
	if err != nil {
		t.Fatal(err)
	}
	ui, err := user.NewInit(suite.TestDatabase, result, suite.TestEnv["testUser"])
	if err != nil {
		t.Fatal(err)
	}
	if ui.TZOffset == 0 {
		t.Errorf("test user init does not have a timezone offset")
	}
}
