package event

import (
	"fmt"
	"strconv"
	"strings"

	jsonextract "github.com/blainemoser/JsonExtract"
	"github.com/blainemoser/MySqlDB/database"
	utils "github.com/blainemoser/goutils"
	"github.com/blainemoser/todobot/user"
)

const createEvent = `insert into events (channel, ts, etext, etype, user_id) values (?,?,?,?,?)`

const findUser = `select * from users where uhash = ?`

type Event struct {
	*database.Database
	User      *user.User
	ID        int64
	Channel   string
	Timestamp float64
	Etext     string
	Etype     string
}

type EventInit struct {
	Channel   string
	Timestamp float64
	Etext     string
	Etype     string
	User      string
}

func Create(db *database.Database, init *EventInit) (*Event, error) {
	e := &Event{
		Database:  db,
		Channel:   init.Channel,
		Timestamp: init.Timestamp,
		Etext:     init.Etext,
		Etype:     init.Etype,
	}
	user, err := init.lookupUser(db)
	if err != nil {
		return nil, err
	}
	e.User = user
	return e.insert()
}

func (i *EventInit) lookupUser(db *database.Database) (*user.User, error) {
	result, err := db.QueryRaw(findUser, []interface{}{i.User})
	if err != nil {
		return nil, err
	}
	if len(result) < 1 {
		return user.Create(db, i.User)
	}
	return user.CreateFromRecord(result[0], db)
}

func (e *Event) insert() (*Event, error) {
	result, err := e.Exec(createEvent, []interface{}{
		e.Channel, e.Timestamp, e.Etext, e.Etype, e.User.ID(),
	})
	if err != nil {
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	e.ID = id
	return e, nil
}

func CreateFromPayload(payload string) (*EventInit, error) {
	eventExtract := jsonextract.JSONExtract{
		RawJSON: payload,
	}
	errs := []string{}
	ei := &EventInit{}
	for _, wants := range []string{
		"type", "text", "user", "ts", "channel",
	} {
		result, err := eventExtract.Extract(fmt.Sprintf("event/%s", wants))
		if err != nil {
			errs = append(errs, err.Error())
			continue
		}
		ei.addToEventInit(wants, result)
	}
	if len(errs) > 0 {
		return nil, fmt.Errorf(strings.Join(errs, ", "))
	}
	return ei, nil
}

func (ei *EventInit) addToEventInit(wants string, result interface{}) {
	switch wants {
	case "type":
		ei.Etype = utils.StringInterface(result)
		return
	case "text":
		ei.Etext = utils.StringInterface(result)
		return
	case "user":
		ei.User = utils.StringInterface(result)
		return
	case "ts":
		flInterface := utils.StringInterface(result)
		fl, _ := strconv.ParseFloat(flInterface, 10)
		ei.Timestamp = fl
		return
	case "channel":
		ei.Channel = utils.StringInterface(result)
		return
	}
}
