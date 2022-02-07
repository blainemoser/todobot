package user

import (
	"fmt"
	"testing"

	jsonextract "github.com/blainemoser/JsonExtract"
	utils "github.com/blainemoser/goutils"
	"github.com/blainemoser/todobot/tests"
	"github.com/blainemoser/todobot/testsuite"
)

var (
	suite        *testsuite.TestSuite
	eventExtract jsonextract.JSONExtract
	testUser     map[string]interface{}
	newTUser     *User
)

func TestMain(m *testing.M) {
	var err error
	suite, err = testsuite.Initialize("user")
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
	testUser["uname"] = "Blaine Moser"
	return err
}

func TestCreate(t *testing.T) {
	err := makeUser()
	if err != nil {
		t.Fatal(err)
	}
}

func TestFind(t *testing.T) {
	err := makeUser()
	if err != nil {
		t.Fatal(err)
	}
	u, err := Find(suite.TestDatabase, newTUser.id)
	if err != nil {
		t.Fatal(err)
	}
	if u == nil {
		t.Fatalf("user not found")
	}
	if u.id != newTUser.id {
		t.Fatalf("user ids do not match, found %d, want %d", u.id, newTUser.id)
	}
}

func TestCreateFromRecord(t *testing.T) {
	err := makeUser()
	if err != nil {
		t.Fatal(err)
	}
	record, err := suite.TestDatabase.QueryRaw("select * from users where uhash = ?", []interface{}{newTUser.uhash})
	if err != nil {
		t.Fatal(err)
	}
	if len(record) < 1 {
		t.Fatalf("record not found")
	}
	u, err := CreateFromRecord(record[0], suite.TestDatabase)
	if err != nil {
		t.Fatal(err)
	}
	if u.id != newTUser.id {
		t.Fatalf("expected user to had the id %d, got %d", newTUser.id, u.id)
	}
}

func TestHash(t *testing.T) {
	err := makeUser()
	if err != nil {
		t.Fatal(err)
	}
	hash := newTUser.Hash()
	if hash != newTUser.uhash {
		t.Fatalf("expected hash to be '%s', got '%s'", newTUser.uhash, hash)
	}
}

func TestID(t *testing.T) {
	err := makeUser()
	if err != nil {
		t.Fatal(err)
	}
	id := newTUser.ID()
	if id != newTUser.id {
		t.Fatalf("expected id to be %d, got %d", newTUser.id, id)
	}
}

func makeUser() error {
	if newTUser != nil {
		return nil
	}
	ui, err := getUserInit()
	if err != nil {
		return err
	}
	user, err := Create(ui)
	if err = checkUserError(user, err); err != nil {
		return err
	}
	newTUser = user
	return nil
}

func TestUpdate(t *testing.T) {
	err := makeUser()
	if err != nil {
		t.Fatal(err)
	}
	newTUser, err = newTUser.SetTZ("America/Los_Angeles").SetTZLabel("America Los Angeles").SetTZOffset(-28800).Update()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(newTUser)
}

func checkUserError(user *User, err error) error {
	if err != nil {
		return err
	}
	if user == nil {
		return fmt.Errorf("user not set")
	}
	if user.id < 1 {
		return fmt.Errorf("user has no id set")
	}
	if user.tzOffset < 1 {
		return fmt.Errorf("user has no timezone-offset set")
	}
	return nil
}

func getUserInit() (*UserInit, error) {
	hash := utils.StringInterface(testUser["uhash"])
	if len(hash) < 1 {
		return nil, fmt.Errorf("expected user hash to be set, got '%v'", testUser["uhash"])
	}
	return NewInit(suite.TestDatabase, fmt.Sprintf(tests.TestUserPayload, hash), hash)
}
