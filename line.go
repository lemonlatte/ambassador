package ambassador

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
			SenderId:    event.Source.UserId,
			Timestamp:   event.Timestamp,
			RecipientId: event.ReplyToken,
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
		"messages":   messages,
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

func (l *LineAmbassador) AskQuestion(recipientId string, text string, answers []map[string]string) (err error) {
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

	quesPayload := map[string]interface{}{
		"type":    "template",
		"altText": "this is a buttons template",
		"template": map[string]interface{}{
			"type": "buttons",
			// "title": "問題",
			"text":    text,
			"actions": actions,
		},
	}

	err = l.sendReply(recipientId, []interface{}{quesPayload})
	return
}

func (l *LineAmbassador) SendText(recipientId string, text string) (err error) {
	messages := []map[string]string{
		{"type": "text", "text": text},
	}

	err = l.sendReply(recipientId, messages)
	return
}

func (l *LineAmbassador) SendTemplate(recipientId string, elements interface{}, bulked bool) (err error) {

	columns := []map[string]interface{}{}
	colItems, ok := elements.([]map[string]interface{})
	if !ok {
		return fmt.Errorf("can not assert the elements' type")
	}

	for _, col := range colItems {
		columns = append(columns, map[string]interface{}{
			"title":             col["title"],
			"text":              col["subtitle"],
			"thumbnailImageUrl": col["image_url"].(string),
			"actions": []map[string]string{
				{"type": "uri", "label": "檢視",
					"uri": col["item_url"].(string),
				},
			},
		})
	}

	carousel := map[string]interface{}{
		"type":    "template",
		"altText": "this is a carousel template",
		"template": map[string]interface{}{
			"type":    "carousel",
			"columns": columns,
		},
	}

	err = l.sendReply(recipientId, []interface{}{carousel})
	if err != nil {
		b, _ := json.Marshal([]interface{}{carousel})
		return fmt.Errorf("%s, %s", err.Error(), b)
	}
	return
}

func (l *LineAmbassador) Send(recipientId string) (err error) {
	return
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
