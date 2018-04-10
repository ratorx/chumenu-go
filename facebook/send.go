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

const (
	Response     messageType = "RESPONSE"
	Subscription messageType = "NON_PROMOTIONAL_SUBSCRIPTION"
)

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

type QuickReply struct {
	Text string
}

func (qr *QuickReply) MarshalJSON() ([]byte, error) {
	type t struct {
		A string `json:"title"`
		B string `json:"content_type"`
		C string `json:"payload"`
	}

	temp := t{qr.Text, "text", ""}
	return json.Marshal(temp)
}

func NewQuickReplySlice(labels []string) []QuickReply {
	qrs := make([]QuickReply, 0, len(labels))
	for _, str := range labels {
		qrs = append(qrs, QuickReply{Text: str})
	}

	return qrs
}

type SendMessage struct {
	Message
	Replies []QuickReply `json:"quick_replies"`
}

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

func (c *SendClient) SendMessage(r string, text string, mType messageType, qr []QuickReply) error {
	return c.apiCall(&Payload{Recipient: Recipient{r}, Message: &SendMessage{Message: Message{Text: text, Metadata: c.Metadata}, Replies: qr}, Type: mType})
}

type MessageError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Subcode int    `json:"error_subcode"`
	Type    string `json:"type"`
}

func (m MessageError) Error() string {
	return fmt.Sprintf("send api: %v", m.Message)
}
