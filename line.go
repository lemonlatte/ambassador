package ambassador

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
)

const LineBotReplyURI = "https://api.line.me/v2/bot/message/reply"

type LineObject struct {
	Events []LineEvent `json:"events"`
}

type LineEvent struct {
	ReplyToken string       `json:"replyToken"`
	Type       string       `json:"type"`
	Timestamp  int64        `json:"timestamp"`
	Source     LineSource   `json:"source"`
	Message    LineMessage  `json:"message"`
	Postback   LinePostback `json:"postback"`
}

type LineSource struct {
	Type   string `json:"type"`
	UserId string `json:"userId"`
}

type LineMessage struct {
	Id   string `json:"id"`
	Type string `json:"type"`
	Text string `json:"text"`

	Title     string  `json:"title"`
	Address   string  `json:"address"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type LinePostback struct {
	Payload string `json:"data"`
}

type LineAmbassador struct {
	sync.Mutex
	channelToken string
	client       *http.Client
	messages     []interface{}
}

func (l *LineAmbassador) Translate(r io.Reader) (messages []Message, err error) {
	var v LineObject
	d := json.NewDecoder(r)
	err = d.Decode(&v)
	if err != nil {
		return
	}

	messages = make([]Message, 0, 10)

	for _, event := range v.Events {
		msg := Message{
			SenderId:   event.Source.UserId,
			ReplyToken: event.ReplyToken,
			Timestamp:  event.Timestamp,
		}
		switch event.Type {
		case "message":
			switch event.Message.Type {
			case "location":
				msg.Content = &LocationContent{
					Lat: event.Message.Latitude,
					Lon: event.Message.Longitude,
				}
			case "text":
				msg.Content = &TextContent{Text: event.Message.Text}
			default:
			}
		case "postback":
			msg.Content = &CommandContent{Payload: event.Postback.Payload}
		default:
		}
		messages = append(messages, msg)
	}

	return
}

func (l *LineAmbassador) sendReply(recipientId string, messages interface{}) (err error) {
	payload := map[string]interface{}{
		"replyToken": recipientId,
		"messages":   l.messages,
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return
	}

	req, _ := http.NewRequest("POST", LineBotReplyURI, bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+l.channelToken)
	resp, err := l.client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		buffer := &bytes.Buffer{}
		_, err = io.Copy(buffer, resp.Body)
		if err != nil {
			return
		}
		err = fmt.Errorf("fail to reply a line message. status: %s, body: %s",
			resp.Status, buffer.String())
	}
	return
}

func (l *LineAmbassador) AskQuestion(text string, answers []map[string]string) (err error) {
	actions := []map[string]string{}
	var upperBound int
	if upperBound = len(answers) - 4; upperBound < 0 {
		upperBound = 0
	}
	for _, answer := range answers[upperBound:] {
		ansLabel, ok1 := answer["title"]
		ansData, ok2 := answer["payload"]
		if ok1 && ok2 {
			actions = append(actions, map[string]string{
				"type":  "postback",
				"label": ansLabel,
				"data":  ansData,
				"text":  ansLabel,
			})
		}
	}

	question := map[string]interface{}{
		"type":    "template",
		"altText": "this is a buttons template",
		"template": map[string]interface{}{
			"type": "buttons",
			// "title": "問題",
			"text":    text,
			"actions": actions,
		},
	}

	l.Lock()
	defer l.Unlock()
	l.messages = append(l.messages, question)
	return
}

func (l *LineAmbassador) SendText(text string) (err error) {
	textMessage := []map[string]string{
		{"type": "text", "text": text},
	}
	l.Lock()
	defer l.Unlock()
	l.messages = append(l.messages, textMessage)
	return
}

func (l *LineAmbassador) SendTemplate(elements interface{}) (err error) {
	columns := []map[string]interface{}{}
	colItems, ok := elements.([]Carousel)
	if !ok {
		return fmt.Errorf("can not type assert the elements")
	}

	for i, col := range colItems {
		if i > 5 {
			break
		}
		item := map[string]interface{}{
			"title":             col.Title,
			"text":              col.Text,
			"thumbnailImageUrl": col.ImageUrl,
		}

		actions := []map[string]string{}
		if col.ItemUrl != "" {
			actions = append(actions,
				map[string]string{
					"type": "uri", "label": "Item Link",
					"uri": col.ItemUrl,
				})
		}

		for _, btn := range col.Buttons {
			if len(actions) > 4 {
				break
			}
			action := map[string]string{"label": btn.Label}

			switch btn.Type {
			case "url":
				action["type"] = "uri"
				action["uri"] = btn.Data
			}
			actions = append(actions, action)
		}

		if len(actions) > 0 {
			item["actions"] = actions
		}

		columns = append(columns, item)
	}

	carousel := map[string]interface{}{
		"type":    "template",
		"altText": "this is a carousel template",
		"template": map[string]interface{}{
			"type":    "carousel",
			"columns": columns,
		},
	}
	l.Lock()
	defer l.Unlock()
	l.messages = append(l.messages, carousel)
	return
}

func (l *LineAmbassador) Send(recipientId string) (err error) {
	defer l.cleanMessage()
	err = l.sendReply(recipientId, l.messages)
	return
}

func (l *LineAmbassador) cleanMessage() {
	l.Lock()
	defer l.Unlock()
	l.messages = []interface{}{}
}

func NewLineAmbassador(channelToken string, client *http.Client) *LineAmbassador {
	if client == nil {
		client = http.DefaultClient
	}
	return &LineAmbassador{
		channelToken: channelToken,
		client:       client,
	}
}
