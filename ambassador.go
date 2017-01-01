package ambassador

import (
	"io"
	"net/http"
)

type Message struct {
	SenderId    string
	RecipientId string
	Timestamp   int64
	Body        interface{}
}

type Ambassador interface {
	Translate(r io.Reader) (messages []Message, err error)
	AskQuestion(recipientId string, text string, answers []map[string]string) (err error)
	SendText(recipientId string, text string) (err error)
	SendTemplate(recipientId string, elements interface{}) (err error)
}

func New(source, token string, client *http.Client) (a Ambassador) {
	switch source {
	case "facebook":
		return NewFBAmbassador(token, client)
	}
	return
}
