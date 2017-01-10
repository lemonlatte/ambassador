package ambassador

import (
	"io"
	"net/http"
)

type Message struct {
	SenderId    string
	RecipientId string
	Timestamp   int64
	Content     interface{}
}

type LocationContent struct {
	Lat float64
	Lon float64
}

type TextContent struct {
	Text string
}

type CommandContent struct {
	Payload string
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
