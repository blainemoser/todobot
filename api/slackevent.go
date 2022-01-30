package api

import (
	"fmt"
	"net/http"
	"strings"

	jsonextract "github.com/blainemoser/JsonExtract"
	slackresponse "github.com/blainemoser/slackResponse"
	"github.com/blainemoser/todobot/event"
)

var (
	challengeExpects []string = []string{
		"token", "challenge", "type",
	}
)

func (a *Api) slackEvent(w http.ResponseWriter, r *http.Request) {
	response := a.NewResponse(w, r)
	defer response.Respond()
	response.CheckMethod(http.MethodPost)
	body := response.getRequestBody()
	if response.evaluateSlackChallenge(body) == true {
		return
	}
	response.handleSlackEvent(body)
}

func (r *Response) evaluateSlackChallenge(body []byte) bool {
	challenge, err := r.isSlackChallenge(body)
	if err != nil {
		r.HandleError(http.StatusInternalServerError, "something went wrong", err)
	}
	if len(challenge) > 0 {
		r.message = []byte(fmt.Sprintf(`{"challenge": "%s"}`, challenge))
		return true
	}
	return false
}

func (a *Api) isSlackChallenge(body []byte) (string, error) {
	extract := jsonextract.JSONExtract{
		RawJSON: string(body),
	}
	return a.getSlackChallenge(extract)
}

func (a *Api) getSlackChallenge(extract jsonextract.JSONExtract) (hasString string, err error) {
	var ok bool
	for _, wants := range challengeExpects {
		has, err := extract.Extract(wants)
		if err != nil {
			return "", nil
		}
		if wants != "challenge" {
			continue
		}
		hasString, ok = has.(string)
		if !ok {
			return "", fmt.Errorf("could not parse the challenge")
		}
	}
	return hasString, nil
}

func (r *Response) handleSlackEvent(body []byte) {
	extract := jsonextract.JSONExtract{
		RawJSON: string(body),
	}
	if _, err := extract.Extract("token"); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "invalid") {
			r.HandleError(http.StatusBadRequest, "invalid request body", err)
			return
		}
	}
	r.newSlackEvent(body)
}

func (r *Response) newSlackEvent(body []byte) {
	e, err := event.CreateFromPayload(string(body), r.Database)
	if err != nil {
		r.HandleError(http.StatusInternalServerError, "something went wrong", err)
		return
	}
	if e != nil {
		r.eventResponse(e)
	}
}

func (r *Response) eventResponse(e *event.Event) {
	err := slackresponse.SlackPost(
		"thanks for that",
		fmt.Sprintf("<@%s> posted a message:\n'%s'", e.User.Hash(), e.Etext),
		"INFO",
		r.SlackURL,
		r.Log,
	)
	if err != nil {
		r.HandleError(http.StatusInternalServerError, "something went wrong", err)
		return
	}
}
