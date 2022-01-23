package api

import (
	"fmt"
	"io/ioutil"
	"net/http"

	logging "github.com/blainemoser/Logging"
	"github.com/blainemoser/MySqlDB/database"
)

const errorResponse = `{"error": true, "message": "%s"}`

type Api struct {
	Port int
	*logging.Log
	*database.Database
}

type Response struct {
	*Api
	r       *http.Request
	w       http.ResponseWriter
	code    int
	err     error
	message []byte
}

func (a *Api) NewResponse(w http.ResponseWriter, r *http.Request) *Response {
	resp := &Response{
		Api: a, r: r, w: w, err: nil, message: []byte(""), code: 200,
	}
	return resp
}

func (r *Response) Respond() {
	p := recover()
	if p != nil {
		r.writeError(p)
		return
	}
	defer r.closeRequest()
	r.w.Header().Add("Content-Type", "application/json")
	r.w.WriteHeader(r.code)
	r.w.Write(r.message)
}

func (r *Response) writeError(p interface{}) {
	defer r.closeRequest()
	r.w.Header().Add("Content-Type", "application/json")
	if err, ok := p.(error); ok {
		r.ErrLog(err, false)
	}
	r.w.WriteHeader(r.code)
	r.w.Write(r.message)
}

func Boot(port int, logger *logging.Log, db *database.Database) *Api {
	api := &Api{
		Port:     port,
		Log:      logger,
		Database: db,
	}
	api.controller()
	return api
}

func (a *Api) Run() error {
	a.Write(fmt.Sprintf("Todo Bot API started on port %d", a.Port), "INFO")
	return http.ListenAndServe(fmt.Sprintf(":%d", a.Port), nil)
}

func (r *Response) getBody() (data []byte, err error) {
	if r == nil || r.r.Response == nil || r.r.Response.Body == nil {
		return nil, fmt.Errorf("nil response")
	}
	defer r.r.Response.Body.Close()
	data, err = ioutil.ReadAll(r.r.Response.Body)
	if err != nil {
		return nil, err
	}
	if len(data) < 1 {
		return nil, fmt.Errorf("no response received")
	}
	return data, err
}

func (r *Response) getRequestBody() []byte {
	if r == nil || r.r.Body == nil {
		r.HandleError(http.StatusBadRequest, "request has no body", nil)
	}
	data, err := ioutil.ReadAll(r.r.Body)
	if err != nil {
		r.HandleError(http.StatusInternalServerError, "something went wrong", err)
	}
	if len(data) < 1 {
		r.HandleError(http.StatusBadRequest, "request has no body", nil)
	}
	return data
}

func (r *Response) CheckMethod(expectedMethod string) {
	if r.r.Method != expectedMethod {
		err := fmt.Errorf("method '%s' is not allowed", r.r.Method)
		r.HandleError(http.StatusMethodNotAllowed, err.Error(), err)
	}
}

func (r *Response) HandleError(code int, message string, err error) {
	r.code = code
	r.message = []byte(fmt.Sprintf(errorResponse, message))
	r.err = err
	panic(r.err)
}

func (r *Response) closeRequest() {
	if r == nil {
		return
	}
	if r.r.Response != nil && r.r.Response.Body != nil {
		r.r.Response.Body.Close()
	}
	if r.r.Body != nil {
		r.r.Body.Close()
	}
}
