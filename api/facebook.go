package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	// "log"
	"net/http"
	"net/url"
	// "os"
	"strings"
)

const api = "https://graph.facebook.com/v2.6/me/messages"
const access_token = "INSERT PAGE_ACCESS_TOKEN HERE"

var values = url.Values{
	"access_token": []string{access_token},
}

// type Quick_Reply struct {
// 	content_type string // text or location
// 	title        string // Button caption - required only if content_type is text
// 	payload      string // Button callback - required only if content_type is text
// 	image_url    string // url of image for text replies
// }

type Message struct {
	Text string `json:"text"`
	// quick_replies []Quick_Reply
	Metadata string `json:"metadata"`
}

type Recipient struct {
	ID string `json:"id"`
}

type Data struct {
	Recipient        *Recipient `json:"recipient"`
	Message          *Message   `json:"message"`
	SenderAction     string     `json:"sender_action"`
	NotificationType string     `json:"notification_type"`
}

func getURLString(url_string string, values url.Values) string {
	temp := []string{url_string, "?", values.Encode()}

	return strings.Join(temp, "")
}

func PostRequest(data *Data) bool {
	b, err := json.Marshal(*data)
	if err != nil {
		panic(err)
	}

	// Send the POST request
	client := &http.Client{}
	resp, err := client.Post(getURLString(api, values), "application/json", bytes.NewReader(b))
	if err != nil {
		panic(err)
	}

	// Parse Response
	if resp.StatusCode != 200 {
		respData, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}
		fmt.Println(resp.StatusCode, string(respData))
	}
	defer resp.Body.Close()

	return resp.StatusCode != http.StatusOK
}

func SimpleMessage(message string, recipients []string) []bool {
	m := Message{Text: message, Metadata: "Sent by Churchill Menu bot."}
	r := Recipient{}
	d := Data{Message: &m, Recipient: &r}
	success := make([]bool, 0, len(recipients))

	for i := range recipients {
		r = Recipient{ID: recipients[i]}
		success = append(success, PostRequest(&d))
	}

	return success
}

// Build up messaging capabilities to support messages with attachements
