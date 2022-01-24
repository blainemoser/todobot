package event

import (
	"testing"
	"time"

	jsonextract "github.com/blainemoser/JsonExtract"
	"github.com/blainemoser/todobot/testsuite"
	"github.com/blainemoser/todobot/user"
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
	suite         *testsuite.TestSuite
	eventExtract  jsonextract.JSONExtract
	testUser      map[string]interface{}
	newTUser      *user.User
	testEventInit *EventInit
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
	err = getTestEventInit()
	if err != nil {
		panic(err)
	}
	suite.ResultCode = m.Run()
}

func TestCreate(t *testing.T) {
	e, err := Create(suite.TestDatabase, testEventInit)
	if err != nil {
		t.Fatal(err)
	}
	checkEvent(e, t)
	uid := e.User.ID()
	// The purpose of this second test is to make sure the user stays the same; this time it's a lookup and not a create
	testEventInit.Channel = "C02NLG80AAA"
	testEventInit.Timestamp = float64(time.Now().UnixNano())
	etwo, err := Create(suite.TestDatabase, testEventInit)
	checkEvent(etwo, t)
	if etwo.User.ID() != uid {
		t.Fatalf("expected event user id to be %d, got %d", uid, etwo.User.ID())
	}
}

func checkEvent(e *Event, t *testing.T) {
	if e.ID < 1 {
		t.Fatalf("expected event id to be greater than 0, got %d", e.ID)
	}
	if e.Etext != testEventInit.Etext {
		t.Fatalf("expected event text to be '%s', got '%s'", testEventInit.Etext, e.Etext)
	}
	if e.Etype != testEventInit.Etype {
		t.Fatalf("expected event type to be '%s', got '%s'", testEventInit.Etype, e.Etype)
	}
	if e.Channel != testEventInit.Channel {
		t.Fatalf("expected event channel to be '%s', got '%s'", testEventInit.Channel, e.Channel)
	}
	if e.Timestamp != testEventInit.Timestamp {
		t.Fatalf("expected event channel to be '%f', got '%f'", testEventInit.Timestamp, e.Timestamp)
	}
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
	return err
}

func getTestEventInit() (err error) {
	testEventInit, err = CreateFromPayload(testEvent)
	return err
}
