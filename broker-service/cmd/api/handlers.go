package main

import (
	"broker/event"
	"broker/logs"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/rpc"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type RequestPayload struct {
	Action string      `json:"action"`
	Auth   AuthPayload `json:"auth,omitempty"` // Will omit it if it's not there
	Log    LogPayload  `json:"log,omitempty"`
	Mail   MailPayload `json:"mail,omitempty"`
}

type MailPayload struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Subject string `json:"subject"`
	Message string `json:"message"`
}

type AuthPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LogPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
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
	//Handle submission actually expects to receive some
	//kind of payload and that's described up here
	var requestPayload RequestPayload

	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	switch requestPayload.Action {
	case "auth":
		app.authenticate(w, requestPayload.Auth)
	case "log":
		//log via RPC
		app.logItemViaRPC(w, requestPayload.Log)
		// log via RabbitMQ
		//app.logEventViaRabbit(w, requestPayload.Log)
		//app.logItem(w, requestPayload.Log)
	case "mail":
		app.sendMail(w, requestPayload.Mail)
	default:
		//app.errorJSON(w, errors.New("unknown action"))
		app.errorJSON(w, errors.New("unknown action: "+requestPayload.Action))

	}

}

func (app *Config) sendMail(w http.ResponseWriter, msg MailPayload) {
	jsonData, _ := json.MarshalIndent(msg, "", "\t")

	//call the mail service
	mailServiceURL := "http://mailer-service/send"

	//post to mail service
	request, err := http.NewRequest("POST", mailServiceURL, bytes.NewBuffer(jsonData))
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	defer response.Body.Close()

	//make sure we get back the right status code
	if response.StatusCode != http.StatusAccepted {
		app.errorJSON(w, errors.New("error calling mail service"))
		return
	}

	//send back json
	var payload jsonResponse
	payload.Error = false
	payload.Message = "Message sent to " + msg.To

	app.writeJSON(w, http.StatusAccepted, payload)

}

func (app *Config) logItem(w http.ResponseWriter, entry LogPayload) {
	jsonData, _ := json.MarshalIndent(entry, "", "\t")

	logServiceURL := "http://logger-service/log"

	request, err := http.NewRequest("POST", logServiceURL, bytes.NewBuffer(jsonData))
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	response, err := client.Do(request)
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusAccepted {
		app.errorJSON(w, err)
		return
	}

	var payload jsonResponse
	payload.Error = false
	payload.Message = "logged"

	app.writeJSON(w, http.StatusAccepted, payload)

}

// authenticate calls the authentication microservice and sends back the appropriate respone
func (app *Config) authenticate(w http.ResponseWriter, a AuthPayload) {

	//create some json we'll send to the auth microservice
	jsonData, _ := json.MarshalIndent(a, "", "\t")

	// call the service
	request, err := http.NewRequest("POST", "http://authentication-service/authenticate", bytes.NewBuffer(jsonData))
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	client := &http.Client{}
	response, err := client.Do(request)

	if err != nil {
		app.errorJSON(w, err)
		return
	}
	defer response.Body.Close()

	// make sure we get back the correct status code
	if response.StatusCode == http.StatusUnauthorized {
		app.errorJSON(w, errors.New("invalid credentials"))
		return
	} else if response.StatusCode != http.StatusAccepted {
		app.errorJSON(w, errors.New("error calling auth service"))
		return
	}

	// create a variable we'll read response.Body info
	var jsonFromService jsonResponse

	// decode the json from the auth service
	// err = json.NewDecoder(response.Body).Decode(&jsonFromService)
	// if err != nil {
	// 	app.errorJSON(w, err)
	// 	return
	// }

	if err := json.NewDecoder(response.Body).Decode(&jsonFromService); err != nil {
		app.errorJSON(w, err)
		return
	}

	if jsonFromService.Error {
		app.errorJSON(w, err, http.StatusUnauthorized)
		return
	}

	var payload jsonResponse
	payload.Error = false
	payload.Message = "Authenticated!"
	payload.Data = jsonFromService.Data

	app.writeJSON(w, http.StatusAccepted, payload)
}

// login an item by emitting an event to RabbitMQ
func (app *Config) logEventViaRabbit(w http.ResponseWriter, l LogPayload) {
	err := app.pushToQueue(l.Name, l.Data)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	var payload jsonResponse
	payload.Error = false
	payload.Message = "logged via RabbitMQ"

	app.writeJSON(w, http.StatusAccepted, payload)

}

//push to a queue

func (app *Config) pushToQueue(name, msg string) error {

	//it requires a connection
	emitter, err := event.NewEventEmitter(app.Rabbit)
	if err != nil {
		return err
	}
	//if we pass we need to push it to the queue

	payload := LogPayload{
		Name: name,
		Data: msg,
	}

	//Push json to a queue

	j, _ := json.MarshalIndent(&payload, "", "\t")
	//severity
	err = emitter.Push(string(j), "log.INFO")
	if err != nil {
		return err
	}

	return nil
}

type RPCPayload struct {
	Name string
	Data string
}

func (app *Config) logItemViaRPC(w http.ResponseWriter, l LogPayload) {
	// I specify in here the name of the microservice from my Docker compose yml, logger-service
	client, err := rpc.Dial("tcp", "logger-service:5001")
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	// But if I have that client I need to create a payload
	// You need to create a type that exactly matches the one that the remote our RPC server expects to get.
	//Lets create a struct outside this func.
	rpcPayload := RPCPayload{
		Name: l.Name,
		Data: l.Data,
	}
	// So now I have my data, my payload, I'm going to push and I'm going to get some kind of result back
	var result string
	// And here I tell it exactly what I want to call.
	//And it's going to be our RPC server, which is the type that's created on the serverend,
	//and then the name of the function I want to call. it's called logInfo
	// And of course that means that any method I want to expose to our RPC on the serverend must be exported.
	//it has to start with a capital letter or it's just not going to work.
	//The second parameter I'm going to pass it is my data which I just created is called RPC payload
	//And the last thing of course is the response from the server and that's a reference to the variable called result
	err = client.Call("RPCServer.LogInfo", rpcPayload, &result)
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	//And finally, I can actually write some JSON back to the end user.
	//So if everything went well, if I got past this, then I'm all set to give my response.
	payload := jsonResponse{
		Error:   false,
		Message: result,
	}

	app.writeJSON(w, http.StatusAccepted, payload)

}

func (app *Config) LogViaGRPC(w http.ResponseWriter, r *http.Request) {
	//this handler receives a json payload
	var requestPayload RequestPayload
	// read my json payload
	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	//connect to our server from grpc package
	//port 50001
	// we have to have valid credantials to connect.
	// We don't need credantials because we're running everything in it's own Docker cluster
	// but we still need to pass them. grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock()
	conn, err := grpc.Dial("logger-service:50001", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	defer conn.Close()

	//now i need to create a client
	c := logs.NewLogServiceClient(conn)
	//now we need a context
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	//if we run out of time, we cancel
	defer cancel()

	//Next, call WriteLog
	//logs.LogRequest was define in protofile
	_, err = c.WriteLog(ctx, &logs.LogRequest{
		//it has refrence to logs.Log
		LogEntry: &logs.Log{
			Name: requestPayload.Log.Name,
			Data: requestPayload.Log.Data,
		},
	})
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	//if we get pass that we've writtent to the log using gRPC

	var payload jsonResponse
	payload.Error = false
	payload.Message = "logged"

	app.writeJSON(w, http.StatusAccepted, payload)

}
