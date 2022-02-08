package slackapi

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/blainemoser/MySqlDB/database"
)

const userUrl = `https://slack.com/api/users.info?user=%s&pretty=1`

var (
	userExpects []string = []string{
		"user/tz", "user/tz_offset", "user/tz_label",
	}
)

type SlackCall struct {
	*database.Database
	http.Client
	slackToken string
}

func Slack(db *database.Database, slackToken string) *SlackCall {
	return &SlackCall{
		Database:   db,
		slackToken: slackToken,
		Client: http.Client{
			Timeout: time.Duration(30 * time.Second),
		},
	}
}

func (sc *SlackCall) Get(url string) (string, error) {
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", sc.slackToken))
	response, err := sc.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	return sc.getBody(response)
}

func (sc *SlackCall) getBody(response *http.Response) (string, error) {
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Printf("error response from Slack API %s\n", err.Error())
		return "", err
	}
	return string(body), nil
}

func (sc *SlackCall) GetUserDetails(uhash string) (string, error) {
	return sc.Get(fmt.Sprintf(userUrl, uhash))
}
