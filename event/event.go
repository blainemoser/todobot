package event

import (
	"container/list"
	"fmt"
	"strconv"
	"strings"

	jsonextract "github.com/blainemoser/JsonExtract"
	"github.com/blainemoser/MySqlDB/database"
	utils "github.com/blainemoser/goutils"
	"github.com/blainemoser/todobot/user"
)

type Event struct {
	*database.Database
	User      *user.User
	ID        int64
	Channel   string
	Timestamp float64
	Schedule  int64
	Etext     string
	Etype     string
	Emessage  string
	Next      int64
	text      []string
}

type EventInit struct {
	Channel   string
	Timestamp float64
	Schedule  int64
	Etext     string
	Etype     string
	User      string
}

func ClearQueue() {
	if Queue != nil {
		Queue = nil
		Queue = list.New()
	}
}

func Create(payload string, db *database.Database) (*Event, error) {
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
	return ei.schedule(db)
}

func (e *Event) Message() string {
	if len(e.Emessage) < 1 {
		return e.noneMessage()
	}
	if strings.Contains(e.Emessage, "stop reminding you") {
		return fmt.Sprintf("OK %s, %s\n'%s'", e.userTag(), e.Emessage, e.echo())
	}
	return fmt.Sprintf("OK %s, I'll remind you %s\n'%s'", e.userTag(), e.Emessage, e.echo())
}

func (e *Event) Processed() string {
	e.Timestamp = float64(e.Next)
	e.setNext()
	return fmt.Sprintf("Hi %s, a friendly reminder:\n'%s'", e.userTag(), e.echo())
}

func (e *Event) noneMessage() string {
	return fmt.Sprintf(
		"Sorry, %s\nI couldn't understand what you're asking...\n%s",
		e.userTag(),
		exampleMessage(),
	)
}

func exampleMessage() string {
	return fmt.Sprintf(
		"Try asking me to schedule an event:\n'%s'\nOr, remove an existing one:\n'%s'",
		"Remind me to call my lawyer every day\nRemind me to log my time every two hours",
		"Done calling my lawyer\nDone logging my time",
	)
}

func (e *Event) userTag() string {
	return fmt.Sprintf("<@%s>", e.User.Hash())
}

func (ei *EventInit) schedule(db *database.Database) (*Event, error) {
	e, err := ei.createOrUpdate(db)
	if err != nil {
		return nil, err
	}
	return ei.setSchedule(e)
}

func (e *Event) insert() (*Event, error) {
	result, err := e.Exec(createEvent, []interface{}{
		e.Channel, e.Timestamp, e.Schedule, e.Etext, e.Etype, e.User.ID(),
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

func (e *Event) update(ei *EventInit) (*Event, error) {
	e.Channel = ei.Channel
	e.Timestamp = ei.Timestamp
	e.Schedule = ei.Schedule
	e.Etext = ei.Etext
	e.Etype = ei.Etype
	e.text = sanitize(e.echo())
	_, err := e.Exec(updateEvent, []interface{}{
		e.Channel, e.Timestamp, e.Schedule, e.Etext, e.Etype, e.ID,
	})
	return e, err
}

func (ei *EventInit) createOrUpdate(db *database.Database) (e interface{}, err error) {
	if Queue.Len() < 1 {
		return ei.create(db)
	}
	if e = ei.exists(); e != nil {
		return e, nil
	}
	return ei.create(db)
}

func (ei *EventInit) exists() interface{} {
	for i := Queue.Front(); i != nil; i = i.Next() {
		if k, ok := i.Value.(*Event); ok {
			if ei.matches(k) {
				return i
			}
		}
	}
	return nil
}

func (ei *EventInit) matches(e *Event) bool {
	if ei.User != e.User.Hash() {
		return false
	}
	eiText := sanitize(ei.echo())
	eText := sanitize(e.echo())
	matchQuotient := getMatchQuotient(eiText, eText)
	return matchQuotient >= 0.5
}

func excludeWord(word string) bool {
	return numbers[word] > 0 ||
		digits[word] > 0 ||
		word == "every" ||
		strings.Contains(word, "hour") ||
		strings.Contains(word, "day")
}

func removeWords(list []string) []string {
	result := make([]string, 0)
	for _, word := range list {
		if excludeWord(word) == true {
			continue
		}
		result = append(result, word)
	}
	return result
}

func findMQ(shorter, longer []string) float64 {
	var count float64
	count = 0
	listMap := makeMapOfStrings(shorter)
	for _, v := range longer {
		if listMap[v] {
			delete(listMap, v)
			count++
		}
	}
	if count <= 0 {
		return 0
	}
	return count / float64(len(shorter))
}

func makeMapOfStrings(input []string) (m map[string]bool) {
	m = make(map[string]bool)
	for _, v := range input {
		m[v] = true
	}
	return m
}

func getMatchQuotient(eiText, eText []string) float64 {
	eiText = removeWords(eiText)
	eText = removeWords(eText)
	if len(eiText) < 1 || len(eText) < 1 {
		return 0
	}
	if len(eiText) < len(eText) {
		return findMQ(eiText, eText)
	}
	return findMQ(eText, eiText)
}

func (ei *EventInit) create(db *database.Database) (*Event, error) {
	e := &Event{
		Database:  db,
		Channel:   ei.Channel,
		Timestamp: ei.Timestamp,
		Schedule:  ei.Schedule,
		Etext:     ei.Etext,
		Etype:     ei.Etype,
	}
	user, err := ei.lookupUser(db)
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

func sanitize(etext string) []string {
	etext = NewLines.ReplaceAllString(etext, " ")
	etext = MultiSpace.ReplaceAllString(etext, " ")
	etext = strings.ToLower(etext)
	return strings.Split(etext, " ")
}

func (ei *EventInit) setSchedule(input interface{}) (*Event, error) {
	if input == nil {
		return nil, nil
	}
	switch e := input.(type) {
	case *list.Element:
		return ei.handleExisting(e)
	case *Event:
		e.updateSchedule()
		e.pushToQueue()
		return e, nil
	default:
		return nil, fmt.Errorf("type unrecognised")
	}
}

func (e *Event) pushToQueue() {
	if e.Schedule < 1 {
		// Don't push the event if it's not parseable
		e.Next = 0
		return
	}
	Queue.PushBack(e)
}

func (ei *EventInit) handleExisting(input *list.Element) (e *Event, err error) {
	var ok bool
	if e, ok = input.Value.(*Event); ok {
		e, err = e.update(ei)
		if err != nil {
			return nil, err
		}
		if e.isScheduleRemoval() == true {
			e.Emessage = "I'll stop reminding you"
			Queue.Remove(input)
			return e, nil
		}
		e.updateSchedule()
		return e, nil
	}
	return nil, fmt.Errorf("could not find existing event in list")
}

func (e *Event) isScheduleAdd() bool {
	if e.text == nil {
		e.text = sanitize(e.echo())
	}
	for _, add := range e.text {
		if reminders[add] {
			return true
		}
	}
	return false
}

func (e *Event) isScheduleRemoval() bool {
	if e.text == nil {
		e.text = sanitize(e.echo())
	}
	for _, remove := range e.text {
		if cancellations[remove] {
			return true
		}
	}
	return false
}

func (e *Event) updateSchedule() (err error) {
	if !e.isScheduleAdd() {
		e.Schedule = 0
	} else if strings.Contains(strings.ToLower(e.Etext), "every") {
		e.every()
	} else {
		e.Schedule = int64(day)
		e.Emessage = "every day"
	}
	e.setNext()
	return err
}

func (e *Event) setNext() {
	e.Next = int64(e.Timestamp + float64(e.Schedule))
}

func (e *Event) every() (err error) {
	number := e.getNumber()
	if strings.Contains(strings.ToLower(e.Etext), "day") {
		e.scheduleDay(number)
		return nil
	}
	if strings.Contains(strings.ToLower(e.Etext), "hour") {
		e.scheduleHour(number)
		return nil
	}
	return fmt.Errorf("nothing is parsed")
}

func (e *Event) scheduleDay(number int) {
	if number < 1 {
		e.Emessage = "every day"
		e.Schedule = int64(day)
		return
	}
	e.Emessage = fmt.Sprintf("every %d day(s)", number)
	e.Schedule = int64(number * day)
}

func (e *Event) scheduleHour(number int) {
	if number < 1 {
		e.Emessage = "every hour"
		e.Schedule = int64(hour)
		return
	}
	e.Emessage = fmt.Sprintf("every %d hour(s)", number)
	e.Schedule = int64(number * hour)
}

func (e *Event) getNumber() int {
	textRaw := e.echo()
	for num, figure := range numbers {
		if strings.Contains(textRaw, num) {
			return figure
		}
	}
	for figureString, digit := range digits {
		if strings.Contains(textRaw, figureString) {
			return digit
		}
	}
	return 0
}

func (e *Event) echo() string {
	return strings.Trim(removeTag.ReplaceAllString(e.Etext, ""), " ")
}

func (e *EventInit) echo() string {
	return strings.Trim(removeTag.ReplaceAllString(e.Etext, ""), " ")
}
