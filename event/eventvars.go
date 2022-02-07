package event

import (
	"container/list"
	"regexp"

	"github.com/blainemoser/todobot/user"
)

const createEvent = `insert into events (channel, ts, schedule, etext, etype, user_id) values (?,?,?,?,?,?)`

const updateEvent = `update events set channel = ?, ts = ?, schedule = ?, etext = ?, etype = ? where id = ?`

const findUser = `select * from users where uhash = ?`

const bootQueueQuery = `select e.*, u.uhash from events e join users u on u.id = e.user_id where e.schedule > 0`

const day int = 86400

const hour int = 3600

const tFormat = `2006-01-02 15:04:05`

var (
	testingMode                 = false
	testingNow  int64           = 1643505910
	reminders   map[string]bool = map[string]bool{
		"reminder": true,
		"remind":   true,
		"schedule": true,
	}
	cancellations map[string]bool = map[string]bool{
		"cancel":        true,
		"done":          true,
		"remove":        true,
		"complete":      true,
		"completed":     true,
		"finished":      true,
		"resolved":      true,
		"turned around": true,
		"dispensed":     true,
		"removed":       true,
	}
	numbers map[string]int = map[string]int{
		"twentyone":    21,
		"twenty-one":   21,
		"twenty one":   21,
		"twentytwo":    22,
		"twenty-two":   22,
		"twenty two":   22,
		"twentythree":  23,
		"twenty-three": 23,
		"twenty three": 23,
		"twentyfour":   24,
		"twenty-four":  24,
		"twenty four":  24,
		"twenty":       20,
		"nineteen":     19,
		"eighteen":     18,
		"seventeen":    17,
		"sixteen":      16,
		"fifteen":      15,
		"fourteen":     14,
		"thirteen":     13,
		"twelve":       12,
		"eleven":       11,
		"ten":          10,
		"nine":         9,
		"eight":        8,
		"seven":        7,
		"six":          6,
		"five":         5,
		"four":         4,
		"three":        3,
		"two":          2,
		"one":          1,
	}
	digits map[string]int = map[string]int{
		"24": 24,
		"23": 23,
		"22": 22,
		"21": 21,
		"20": 20,
		"19": 19,
		"18": 18,
		"17": 17,
		"16": 16,
		"15": 15,
		"14": 14,
		"13": 13,
		"12": 12,
		"11": 11,
		"10": 10,
		"9":  9,
		"8":  8,
		"7":  7,
		"6":  6,
		"5":  5,
		"4":  4,
		"3":  3,
		"2":  2,
		"1":  1,
	}
	Users          map[int64]*user.User = make(map[int64]*user.User)
	NewLines                            = regexp.MustCompile(`\n+`)
	MultiSpace                          = regexp.MustCompile(`[ ]{2,}`)
	Queue          *list.List           = list.New()
	removeTag                           = regexp.MustCompile(`<@(.*?)>`)
	twentyFourHour                      = regexp.MustCompile(`(.*?)h(.*?)`)
	colonTime                           = regexp.MustCompile(`(.*?):(.*?)`)
)
