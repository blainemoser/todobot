package api

import (
	"net/http"
)

func (a *Api) ping(w http.ResponseWriter, r *http.Request) {
	response := a.NewResponse(w, r)
	defer response.Respond()
	response.CheckMethod(http.MethodGet)
	response.message = []byte(`{"error": false, "message": "pong"}`)
}
