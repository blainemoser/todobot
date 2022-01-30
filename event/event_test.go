package event

import (
	"fmt"
	"strings"
	"testing"

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
	now          int64 = 1643505910
)

func TestMain(m *testing.M) {
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
	e, err := Create(testReminderOne(), suite.TestDatabase)
	if err != nil {
		t.Fatal(err)
	}
	expects := now + (2 * int64(hour))
	if e.Next != expects {
		t.Fatalf("expected next to be %d, got %d", expects, e.Next)
	}
}

func TestUpdate(t *testing.T) {
	ClearQueue()
	e, err := Create(testReminderTwo(), suite.TestDatabase)
	if err != nil {
		t.Fatal(err)
	}
	expects := now + int64(hour)
	if e.Next != expects {
		t.Fatalf("expected next on event one to be %d, got %d", expects, e.Next)
	}
	eUpdate, err := Create(testReminderTwoUpdate(), suite.TestDatabase)
	if err != nil {
		t.Fatal(err)
	}
	if e.ID != eUpdate.ID {
		t.Errorf("expected events to have the same id, got %d and %d", e.ID, eUpdate.ID)
	}
	expects = now + int64(3*hour) // the update asks to change it to three hours
	if e.Next != expects {
		t.Fatalf("expected next on event update to be %d, got %d", expects, eUpdate.Next)
	}
}

func TestRemoval(t *testing.T) {
	ClearQueue()
	e, err := Create(testReminderThree(), suite.TestDatabase)
	if err != nil {
		t.Fatal(err)
	}
	expects := now + int64(2*hour)
	if e.Next != expects {
		t.Errorf("expected next on event one to be %d, got %d", expects, e.Next)
	}
	eUpdate, err := Create(testReminderThreeRemoval(), suite.TestDatabase)
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

func checkEventRemoved(e *Event) error {
	for i := Queue.Front(); i != nil; i = i.Next() {
		if k, ok := i.Value.(*Event); ok {
			if k.ID == e.ID {
				return fmt.Errorf("expected event to have been removed, found event #%d, in queue", k.ID)
			}
		}
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

func payload() string {
	return strings.Replace(tests.TestEventPayload, "[timestamp]", fmt.Sprintf("%d", now), 1)
}
