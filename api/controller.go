package api

import "net/http"

// controller defines the available routes and the functions that handle them
func (a *Api) controller() {
	http.HandleFunc("/slack-event", a.slackEvent)
	http.HandleFunc("/ping", a.ping)
}
