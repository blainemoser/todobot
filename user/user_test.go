package user

import (
	"fmt"
	"testing"

	jsonextract "github.com/blainemoser/JsonExtract"
	utils "github.com/blainemoser/goutils"
	"github.com/blainemoser/todobot/testsuite"
)

const testEvent = `{
    "token": "vcSe2kbpQsFkpBJyVdnM4o5M",
    "team_id": "T02CCAKL1JB",
    "api_app_id": "A02NLFZUPB7",
    "event": {
        "client_msg_id": "202c5873-3f9c-45d9-b64b-4b6c90ba746c",
        "type": "app_mention",
        "text": "<@U02P23SK0SV> Her there.",
        "user": "U02DGLZ7ABA",
        "ts": "1643038674.000200",
        "team": "T02CCAKL1JB",
        "blocks": [
            {
                "type": "rich_text",
                "block_id": "oJh2",
                "elements": [
                    {
                        "type": "rich_text_section",
                        "elements": [
                            {
                                "type": "user",
                                "user_id": "U02P23SK0SV"
                            },
                            {
                                "type": "text",
                                "text": " Her there."
                            }
                        ]
                    }
                ]
            }
        ],
        "channel": "C02NLG80TEH",
        "event_ts": "1643038674.000200"
    },
    "type": "event_callback",
    "event_id": "Ev02VBF8EFPD",
    "event_time": 1643038674,
    "authorizations": [
        {
            "enterprise_id": null,
            "team_id": "T02CCAKL1JB",
            "user_id": "U02P23SK0SV",
            "is_bot": true,
            "is_enterprise_install": false
        }
    ],
    "is_ext_shared_channel": false,
    "event_context": "4-eyJldCI6ImFwcF9tZW50aW9uIiwidGlkIjoiVDAyQ0NBS0wxSkIiLCJhaWQiOiJBMDJOTEZaVVBCNyIsImNpZCI6IkMwMk5MRzgwVEVIIn0"
}`

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
		RawJSON: testEvent,
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
	makeUser()
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
	hash := utils.StringInterface(testUser["uhash"])
	if len(hash) < 1 {
		return fmt.Errorf("expected user hash to be set, got '%v'", testUser["uhash"])
	}
	user, err := Create(suite.TestDatabase, hash)
	if err != nil {
		return err
	}
	if user == nil {
		return fmt.Errorf("user not set")
	}
	if user.id < 1 {
		return fmt.Errorf("user has no id set")
	}
	newTUser = user
	return nil
}
