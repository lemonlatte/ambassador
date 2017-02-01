package ambassador

import (
	"io"
	"net/http"
)

type Message struct {
	SenderId    string
	ReplyToken  string
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
	AskQuestion(text string, answers []map[string]string) (err error)
	SendText(text string) (err error)
	SendTemplate(elements interface{}) (err error)
	Send(recipientId string) (err error)
}

type CarouselButton struct {
	Label string
	Type  string
	Data  string
}

type Carousel struct {
	Title    string
	Text     string
	ImageUrl string
	ItemUrl  string
	Buttons  []CarouselButton
}

func New(source, token string, client *http.Client) (a Ambassador) {
	switch source {
	case "facebook":
		return NewFBAmbassador(token, client)
	}
	return
}
