package facebook

const (
	// APIBase is the URL of the Facebook endpoint this package supports
	APIBase = "https://graph.facebook.com/v2.11/me/"
)

// Message is a struct which contains the common fields which are sent/received by the Facebook page
type Message struct {
	Text     string `json:"text"`
	Metadata string `json:"metadata"`
}

func (m Message) String() string {
	return m.Text
}

// Recipient is a struct which contains the unique ID of the recipient
type Recipient struct {
	ID string `json:"id"`
}

func (r Recipient) String() string {
	return r.ID
}
