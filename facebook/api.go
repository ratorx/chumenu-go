package facebook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	defaultAPIURL  = "https://graph.facebook.com/v2.6/me/messages"
	defaultWait    = 50 * time.Millisecond
	defaultRetries = 5
)

type Client struct {
	APIURL         string
	AccessToken    string
	allowedRetries uint
}

type Message struct {
	Text     string `json:"text"`
	Metadata string `json:"metadata"`
}

type MessengerError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Subcode int    `json:"error_subcode"`
	Type    string `json:"type"`
}

func (m MessengerError) Error() string {
	return fmt.Sprintf("Facebook: %v", m.Message)
}

func (c *Client) getURL() string {
	if c.APIURL == "" {
		c.APIURL = defaultAPIURL
	}
	return c.APIURL + "?access_token=" + c.AccessToken
}

func (c *Client) apiCall(recipient string, message *Message, allowedRetries uint) error {
	type Recipient struct {
		ID string `json:"id"`
	}

	type Payload struct {
		R Recipient `json:"recipient"`
		M *Message  `json:"message"`
	}

	payload, err := json.Marshal(&Payload{Recipient{recipient}, message})
	if err != nil {
		return err
	}

	response, err := (&http.Client{}).Post(c.getURL(), "application/json", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	defer response.Body.Close()

	var messengerError MessengerError
	err = json.NewDecoder(response.Body).Decode(&messengerError)
	if err != nil {
		return fmt.Errorf("Facebook: Decode error occurred while parsing Send API response from Facebook")
	}

	if messengerError.Message == "" {
		return nil
	}

	switch messengerError.Code {
	case 613, 1200: // Rate Limit or temporary timeout
		if allowedRetries > 0 {
			response.Body.Close()
			time.Sleep(defaultWait)
			return c.apiCall(recipient, message, allowedRetries-1)
		}
	}
	return messengerError
}

func (c *Client) TextMessage(recipient string, message string) error {
	if c.allowedRetries == 0 {
		c.allowedRetries = defaultRetries
	}
	return c.apiCall(recipient, &Message{message, "Sent by Churchill Menus Bot"}, c.allowedRetries)
}
