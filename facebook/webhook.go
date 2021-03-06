package facebook

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

// MessagingEvent contains the details of a particular message received by the page
type MessagingEvent struct {
	Sender  Recipient `json:"sender"`
	Message *Message  `json:"message"`
}

type event struct {
	Messages []MessagingEvent `json:"messaging"`
}

type response struct {
	Events []event `json:"entry"`
}

// EventHandler handles MessagingEvents
type EventHandler interface {
	HandleEvent(me []MessagingEvent)
}

// Webhook contains the information required to create an endpoint for Facebook to call
type Webhook struct {
	AppSecret   string
	VerifyToken string
	Handler     EventHandler
	Debug       *log.Logger
}

func asciiRune(r rune) string {
	switch r {
	case '\\':
		return "\\"
	case '\'':
		return "'"
	}

	s := strconv.QuoteRuneToASCII(r)
	return s[1 : len(s)-1]
}

func asciiString(b []byte) (ret string) {
	for _, r := range string(b) {
		ret += asciiRune(r)
	}

	return ret
}

func (w *Webhook) checkSHA(sha string, body []byte) bool {
	expectedSum, err := hex.DecodeString(sha[5:])
	if err != nil {
		return false
	}

	mac := hmac.New(sha1.New, []byte(w.AppSecret))
	_, err = mac.Write([]byte(asciiString(body)))
	if err != nil {
		return false
	}

	actualSum := mac.Sum(nil)

	return hmac.Equal(expectedSum, actualSum)
}

func (w *Webhook) verify(response http.ResponseWriter, request *http.Request) {
	urlQueries := request.URL.Query()

	switch {
	case urlQueries.Get("hub.mode") != "subscribe":
		http.Error(response, "hub.mode is not subscribe", http.StatusBadRequest)
	case urlQueries.Get("hub.verify_token") != w.VerifyToken:
		http.Error(response, "invalid hub.verify_token", http.StatusUnauthorized)
	default:
		fmt.Fprintf(response, urlQueries.Get("hub.challenge"))
		w.Debug.Print("Webhook verified.")
	}
}

// ResponseHandler performs the task of handling a request made to the Webhook
func (w *Webhook) ResponseHandler(res http.ResponseWriter, request *http.Request) {

	switch request.Method {
	case "GET":
		w.verify(res, request)
	case "POST":
		body, err := ioutil.ReadAll(request.Body)
		if err != nil {
			http.Error(res, "Request body could not be parsed", http.StatusBadRequest)
			w.Debug.Printf("Request body could not be parsed (Error: %s)", err)
			return
		}

		if !w.checkSHA(request.Header.Get("X-Hub-Signature"), body) {
			res.WriteHeader(200)
			w.Debug.Print("Request verification failed")
			w.Debug.Print(string(body))
			return
		}
		r := response{}
		err = json.Unmarshal(body, &r)
		if err != nil {
			w.Debug.Printf("Invalid JSON received (Error: %s)", err)
		}

		for i := range r.Events {
			go w.Handler.HandleEvent(r.Events[i].Messages)
		}

		res.WriteHeader(200)
		w.Debug.Printf("Callback received")
	default:
		http.Error(res, "HTTP method not GET or POST", http.StatusMethodNotAllowed)
		w.Debug.Printf("HTTP method %s attempted", request.Method)
	}
}
