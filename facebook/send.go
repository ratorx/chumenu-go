package facebook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

var retryCodes = [...]int{613, 1200}

type messageType string

// Allowed message types
const (
	Response     messageType = "RESPONSE"
	Subscription messageType = "NON_PROMOTIONAL_SUBSCRIPTION"
)

// SendClient is a struct containing the details required to connect to the Facebook endpoint in order to send messages
type SendClient struct {
	AccessToken string
	BaseURL     string
	Metadata    string
}

func (c *SendClient) getURL() string {
	if c.BaseURL == "" {
		c.BaseURL = APIBase
	}

	return c.BaseURL + "messages?access_token=" + c.AccessToken
}

// QuickReply is a struct representing a single Facebook QuickReply entry
type QuickReply struct {
	Text string
}

// MarshalJSON converts a QuickReply struct into the form expected by the Facebook endpoint
func (qr *QuickReply) MarshalJSON() ([]byte, error) {
	type t struct {
		A string `json:"title"`
		B string `json:"content_type"`
		C string `json:"payload"`
	}

	temp := t{qr.Text, "text", ""}
	return json.Marshal(temp)
}

// NewQuickReplySlice is a convenience function for building multiple QuickReply structs from a list of labels
func NewQuickReplySlice(labels []string) []QuickReply {
	qrs := make([]QuickReply, 0, len(labels))
	for _, str := range labels {
		qrs = append(qrs, QuickReply{Text: str})
	}

	return qrs
}

// SendMessage represents a message that can be sent by the Facebook page
type SendMessage struct {
	Message
	Replies []QuickReply `json:"quick_replies"`
}

// Payload represents the data expected by the endpoint when sending a message
type Payload struct {
	Recipient Recipient    `json:"recipient"`
	Message   *SendMessage `json:"message"`
	Type      messageType  `json:"messaging_type"`
}

func (c *SendClient) apiCall(payload *Payload) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	response, err := (&http.Client{}).Post(c.getURL(), "application/json", bytes.NewReader(b))
	if err != nil {
		return err
	}
	defer response.Body.Close()
	b, _ = ioutil.ReadAll(response.Body)

	temp := struct {
		Error MessageError `json:"error"`
	}{}
	json.Unmarshal(b, &temp)
	if temp.Error.Code != 0 {
		return temp.Error
	}

	return nil
}

// SendMessage is a convenience function to send a message to a particular user
func (c *SendClient) SendMessage(r string, text string, mType messageType, qr []QuickReply) error {
	return c.apiCall(&Payload{Recipient: Recipient{r}, Message: &SendMessage{Message: Message{Text: text, Metadata: c.Metadata}, Replies: qr}, Type: mType})
}

// MessageError represents the data received from Facebook when a erroneous request is made
type MessageError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Subcode int    `json:"error_subcode"`
	Type    string `json:"type"`
}

func (m MessageError) Error() string {
	return fmt.Sprintf("send api: %v", m.Message)
}
