package facebook

const (
	APIBase = "https://graph.facebook.com/v2.11/me/"
)

type Message struct {
	Text     string `json:"text"`
	Metadata string `json:"metadata"`
}

func (m Message) String() string {
	return m.Text
}

type Recipient struct {
	ID string `json:"id"`
}

func (r Recipient) String() string {
	return r.ID
}
