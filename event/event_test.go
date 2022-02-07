package event

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	jsonextract "github.com/blainemoser/JsonExtract"
	"github.com/blainemoser/todobot/tests"
	"github.com/blainemoser/todobot/testsuite"
	"github.com/blainemoser/todobot/user"
)

var (
	suite        *testsuite.TestSuite
	eventExtract jsonextract.JSONExtract
	testUser     map[string]interface{}
	newTUser     *user.User
	gmt          time.Location = *time.Now().UTC().Location()
)

func TestMain(m *testing.M) {
	testingMode = true
	var err error
	suite, err = testsuite.Initialize("event")
	if err != nil {
		panic(err)
	}
	defer suite.TearDown()
	err = getTestUser()
	if err != nil {
		panic(err)
	}
	suite.ResultCode = m.Run()
}

func TestCreateNew(t *testing.T) {
	ClearQueue()
	e, err := Create(testReminderOne(), suite.TestDatabase, suite.TestEnv["slackToken"])
	if err != nil {
		t.Fatal(err)
	}
	expects := testingNow + (2 * int64(hour))
	if e.Next != expects {
		t.Fatalf("expected next to be %d, got %d", expects, e.Next)
	}
}

func TestUpdate(t *testing.T) {
	ClearQueue()
	e, err := Create(testReminderTwo(), suite.TestDatabase, suite.TestEnv["slackToken"])
	if err != nil {
		t.Fatal(err)
	}
	expects := testingNow + int64(hour)
	if e.Next != expects {
		t.Fatalf("expected next on event one to be %d, got %d", expects, e.Next)
	}
	eUpdate, err := Create(testReminderTwoUpdate(), suite.TestDatabase, suite.TestEnv["slackToken"])
	if err != nil {
		t.Fatal(err)
	}
	if e.ID != eUpdate.ID {
		t.Errorf("expected events to have the same id, got %d and %d", e.ID, eUpdate.ID)
	}
	expects = testingNow + int64(3*hour) // the update asks to change it to three hours
	if e.Next != expects {
		t.Fatalf("expected next on event update to be %d, got %d", expects, eUpdate.Next)
	}
}

func TestRemoval(t *testing.T) {
	ClearQueue()
	e, err := Create(testReminderThree(), suite.TestDatabase, suite.TestEnv["slackToken"])
	if err != nil {
		t.Fatal(err)
	}
	expects := testingNow + int64(2*hour)
	if e.Next != expects {
		t.Errorf("expected next on event one to be %d, got %d", expects, e.Next)
	}
	eUpdate, err := Create(testReminderThreeRemoval(), suite.TestDatabase, suite.TestEnv["slackToken"])
	if err != nil {
		t.Fatal(err)
	}
	if e.ID != eUpdate.ID {
		t.Errorf("expected events to have the same id, got %d and %d", e.ID, eUpdate.ID)
	}
	err = checkEventRemoved(eUpdate)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUnparseable(t *testing.T) {
	ClearQueue()
	e, err := Create(testUnparseable(), suite.TestDatabase, suite.TestEnv["slackToken"])
	if err != nil {
		t.Fatal(err)
	}
	if e.Next > 0 {
		t.Errorf("expected next on uparseable event to be 0, got %d", e.Next)
	}
	err = checkEventNotInQueue(e)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(e.Message(), "couldn't understand") {
		t.Errorf("expected unparseable event to contain '%s' in the message, got '%s'", "couldn't understand", e.Message())
	}
}

func TestProcessQueue(t *testing.T) {
	ClearQueue()
	e, err := Create(testReminderOne(), suite.TestDatabase, suite.TestEnv["slackToken"])
	if err != nil {
		t.Fatal(err)
	}
	_, err = Create(testReminderThree(), suite.TestDatabase, suite.TestEnv["slackToken"])
	if err != nil {
		t.Fatal(err)
	}
	c := make(chan []map[string]string, 1)
	go ProcessQueue(c)
	result := <-c
	err = checkQueueProcessResult(result, e)
	if err != nil {
		t.Error(err)
	}
}

func TestList(t *testing.T) {
	ClearQueue()
	_, err := Create(testReminderOne(), suite.TestDatabase, suite.TestEnv["slackToken"])
	if err != nil {
		t.Fatal(err)
	}
	_, err = Create(testReminderThree(), suite.TestDatabase, suite.TestEnv["slackToken"])
	if err != nil {
		t.Fatal(err)
	}
	e, err := Create(testList(), suite.TestDatabase, suite.TestEnv["slackToken"])
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.ToLower(e.Emessage), "your todo list") {
		t.Fatalf("expected list event message to contain a todo list, got %s", e.Emessage)
	}
}

func TestSpecificTime(t *testing.T) {
	ClearQueue()
	e, err := Create(testReminderFour(), suite.TestDatabase, suite.TestEnv["slackToken"])
	if err != nil {
		t.Fatal(err)
	}
	checkTime(t, e.Next, "2022-01-30 11:00")
	e, err = Create(testReminderSix(), suite.TestDatabase, suite.TestEnv["slackToken"])
	if err != nil {
		t.Fatal(err)
	}
	checkTime(t, e.Next, "2022-01-30 23:00")
	e, err = Create(testReminderSeven(), suite.TestDatabase, suite.TestEnv["slackToken"])
	if err != nil {
		t.Fatal(err)
	}
	checkTime(t, e.Next, "2022-01-30 18:17")
	e, err = Create(testReminderEight(), suite.TestDatabase, suite.TestEnv["slackToken"])
	if err != nil {
		t.Fatal(err)
	}
	checkTime(t, e.Next, "2022-01-30 11:34")
	e, err = Create(testReminderNine(), suite.TestDatabase, suite.TestEnv["slackToken"])
	if err != nil {
		t.Fatal(err)
	}
	checkTime(t, e.Next, "2022-01-30 02:34")
}

func checkTime(t *testing.T, next int64, expects string) {
	expectsTime, err := time.ParseInLocation("2006-01-02 15:04", expects, &gmt)
	if err != nil {
		t.Fatal(err)
	}
	nextTime := time.Unix(next, 0)
	if nextTime.Unix() != expectsTime.Unix() {
		t.Errorf(
			"expected event next time to be %s, got %s",
			expectsTime.Format(time.RFC1123),
			nextTime.Format(time.RFC1123),
		)
	}
}

// Be sure to run these last two tests last!
func TestBootstrapQueue(t *testing.T) {
	ClearQueue()
	_, err := suite.TestDatabase.Exec("delete from users", nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = suite.TestDatabase.Exec("delete from events", nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = Create(testReminderOne(), suite.TestDatabase, suite.TestEnv["slackToken"])
	if err != nil {
		t.Fatal(err)
	}
	_, err = Create(testReminderThree(), suite.TestDatabase, suite.TestEnv["slackToken"])
	if err != nil {
		t.Fatal(err)
	}
	ClearQueue()
	err = BootQueue(suite.TestDatabase)
	if err != nil {
		t.Fatal(err)
	}
	if Queue.Len() < 2 {
		t.Errorf("expected queue length to be at least 2, but it is %d", Queue.Len())
	}
	c := make(chan []map[string]string, 1)
	go ProcessQueue(c)
	result := <-c
	expectsOne := "Remind me to pick up the laundry every two hours"
	expectsTwo := "Remind me to do tax forms every two hours"
	for _, m := range result {
		if !strings.Contains(m["message"], expectsOne) && !strings.Contains(m["message"], expectsTwo) {
			t.Errorf("expected event message to be either %s or %s, got %s", expectsOne, expectsTwo, m["message"])
		}
	}
}

func TestCheckQueueUsers(t *testing.T) {
	ClearQueue()
	_, err := suite.TestDatabase.Exec("delete from users", nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = suite.TestDatabase.Exec("delete from events", nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = suite.TestDatabase.Exec("insert into users (uhash) values (?)", []interface{}{
		suite.TestEnv["testUser"],
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = Create(testReminderOne(), suite.TestDatabase, suite.TestEnv["slackToken"])
	if err != nil {
		t.Fatal(err)
	}
	ClearQueue()
	err = BootQueue(suite.TestDatabase)
	if err != nil {
		t.Fatal(err)
	}
	err = CheckEventUsers(suite.TestDatabase, suite.TestEnv["slackToken"])
	if err != nil {
		t.Fatal(err)
	}
	i := Queue.Back()
	e, ok := i.Value.(*Event)
	if !ok {
		t.Fatalf("event not found")
	}
	if len(e.User.TZ()) < 1 {
		t.Errorf("expected timezone for event-user to be populated, got empty string")
	}
	if e.User.TZOffset() != 7200 {
		t.Errorf("expected timezone offset for event-user to be 7200, got %d", e.User.TZOffset())
	}
}

func checkQueueProcessResult(result []map[string]string, e *Event) error {
	if len(result) != 2 {
		return fmt.Errorf("expected there to be two events processed, got %d", len(result))
	}
	expectsHeading := fmt.Sprintf("Hi %s", e.UserTag())
	errs := make([]string, 0)
	msgOne := "Remind me to pick up the laundry every two hours"
	msgTwo := "Remind me to do tax forms every two hours"
	for _, v := range result {
		if v["heading"] != expectsHeading {
			errs = append(errs, fmt.Sprintf("expected heading to be '%s', got '%s'", expectsHeading, v["heading"]))
		}
		if !strings.Contains(v["message"], fmt.Sprintf("'%s'", msgOne)) &&
			!strings.Contains(v["message"], fmt.Sprintf("'%s'", msgTwo)) {
			errs = append(errs, fmt.Sprintf("expected message to be one of %s or %s, got '%s'", msgOne, msgTwo, v["message"]))
		}
	}
	if len(errs) > 0 {
		return errors.New(strings.Join(errs, ", "))
	}
	return nil
}

func checkEventNotInQueue(e *Event) error {
	for i := Queue.Front(); i != nil; i = i.Next() {
		if k, ok := i.Value.(*Event); ok {
			if k.ID == e.ID {
				return fmt.Errorf("expected event to have been removed, found event #%d in queue", k.ID)
			}
		}
	}
	return nil
}

func checkEventRemoved(e *Event) error {
	err := checkEventNotInQueue(e)
	if err != nil {
		return err
	}
	if !strings.Contains(e.Message(), "stop reminding you") {
		return fmt.Errorf("expected event message to contain '%s', got '%s'", "stop reminding you", e.Message())
	}
	return nil
}

func getTestUser() (err error) {
	eventExtract = jsonextract.JSONExtract{
		RawJSON: tests.TestEventPayload,
	}
	uhash, err := eventExtract.Extract("event/user")
	if err != nil {
		return err
	}
	testUser = make(map[string]interface{})
	testUser["uhash"] = uhash
	return err
}

func testReminderOne() string {
	return strings.Replace(
		payload(),
		"[message]",
		"Remind me to pick up the laundry every two hours",
		1,
	)
}

func testReminderTwo() string {
	return strings.Replace(
		payload(),
		"[message]",
		"Remind me to eat every hour",
		1,
	)
}

func testReminderTwoUpdate() string {
	return strings.Replace(
		payload(),
		"[message]",
		"Rather remind me to eat every three hours",
		1,
	)
}

func testReminderThree() string {
	return strings.Replace(
		payload(),
		"[message]",
		"Remind me to do tax forms every two hours",
		1,
	)
}

func testReminderThreeRemoval() string {
	return strings.Replace(
		payload(),
		"[message]",
		"done tax forms",
		1,
	)
}

func testReminderFour() string {
	return strings.Replace(
		payload(),
		"[message]",
		"Remind me to do something cool every day at 1 pm",
		1,
	)
}

func testReminderFive() string {
	return strings.Replace(
		payload(),
		"[message]",
		"Remind me to do something cool every day at 5 am",
		1,
	)
}

func testReminderSix() string {
	return strings.Replace(
		payload(),
		"[message]",
		"Remind me to do a weird thing every day at 1 am",
		1,
	)
}

func testReminderSeven() string {
	return strings.Replace(
		payload(),
		"[message]",
		"Remind me to be kind to a stranger every day at 20h17",
		1,
	)
}

func testReminderEight() string {
	return strings.Replace(
		payload(),
		"[message]",
		"Remind me to be mean to a bully every day at 1:34 pm",
		1,
	)
}

func testReminderNine() string {
	return strings.Replace(
		payload(),
		"[message]",
		"Remind me to help a sick animal every day at 4:34 am",
		1,
	)
}

func testUnparseable() string {
	return strings.Replace(
		payload(),
		"[message]",
		"I think you're pretty cool!",
		1,
	)
}

func testList() string {
	return strings.Replace(
		payload(),
		"[message]",
		"list",
		1,
	)
}

func payload() string {
	return strings.Replace(tests.TestEventPayload, "[timestamp]", fmt.Sprintf("%d", testingNow), 1)
}
