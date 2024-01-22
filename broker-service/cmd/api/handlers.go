package main

import (
	"errors"
	"net/http"
)

type RequestPayload struct {
	Action string      `json:"action"`
	Auth   AuthPayload `json:"auth,omitempty"` // Will omit it if it's not there
}

type AuthPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (app *Config) Broker(w http.ResponseWriter, r *http.Request) {
	payload := jsonResponse{
		Error:   false,
		Message: "Hit the broker",
	}

	_ = app.writeJSON(w, http.StatusOK, payload)

	// out, _ := json.MarshalIndent(payload, "", "\t")
	// w.Header().Set("Content-Type", "application/json")
	// w.WriteHeader(http.StatusAccepted)
	// w.Write(out)
}

func (app *Config) HandleSubmission(w http.ResponseWriter, r *http.Request) {
	var requestPayload RequestPayload

	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	switch requestPayload.Action {
	case "auth":

	default:
		app.errorJSON(w, errors.New("unknown action"))

	}

}

func (app *Config) authenticate(w http.ResponseWriter, a AuthPayload) {

}
