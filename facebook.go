package ambassador

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const FBMessengerBaseURI = "https://graph.facebook.com/v2.6/me/messages?access_token="

type FBObject struct {
	Object string
	Entry  []FBEntry
}

type FBEntry struct {
	Id      string
	Time    int64
	Messags []FBMessage `json:"messaging"`
}

type FBSender struct {
	Id string `json:"id"`
}

type FBRecipient struct {
	Id string `json:"id"`
}

type FBMessage struct {
	Sender    FBSender           `json:"sender,omitempty"`
	Recipient FBRecipient        `json:"recipient,omitempty"`
	Timestamp int64              `json:"timestamp,omitempty"`
	Content   *FBMessageContent  `json:"message,omitempty"`
	Delivery  *FBMessageDelivery `json:"delivery,omitempty"`
	Postback  *FBMessagePostback `json:"postback,omitempty"`
}

type FBMessageContent struct {
	Text        string                `json:"text"`
	Seq         int64                 `json:"seq,omitempty"`
	Attachments []FBMessageAttachment `json:"attachments,omitempty"`
	QuickReplay *FBMessageQuickReply  `json:"quick_reply,omitempty"`
}

type FBMessageQuickReply struct {
	Payload string
}

type FBMessageDelivery struct {
	Watermark int64 `json:"watermark"`
	Seq       int64 `json:"seq"`
}

type FBMessagePostback struct {
	Payload string `json:"payload"`
}

type FBMessageAttachment struct {
	Title   string          `json:",omitempty"`
	Url     string          `json:",omitempty"`
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type FBLocationAttachment struct {
	Coordinates Location `json:"coordinates"`
}

type Location struct {
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"long"`
}

type FBMessageTemplate struct {
	Type     string      `json:"template_type"`
	Elements interface{} `json:"elements"`
}

type FBButtonItem struct {
	Type    string `json:"type"`
	Title   string `json:"title"`
	Url     string `json:"url,omitempty"`
	Payload string `json:"payload,omitempty"`
}

type FBAmbassador struct {
	token  string
	client *http.Client
}

func NewFBAmbassador(token string, client *http.Client) *FBAmbassador {
	if client == nil {
		client = http.DefaultClient
	}
	return &FBAmbassador{
		token:  token,
		client: client,
	}
}

// Translate will turn a facebook messenger object into messages
func (a *FBAmbassador) Translate(r io.Reader) (messages []Message, err error) {
	var v FBObject
	d := json.NewDecoder(r)
	err = d.Decode(&v)
	if err != nil {
		return
	}

	messages = make([]Message, 0, 10)

	for _, entry := range v.Entry {
		for _, fbMsg := range entry.Messags {
			msg := Message{
				SenderId:    fbMsg.Sender.Id,
				RecipientId: fbMsg.Recipient.Id,
				Timestamp:   fbMsg.Timestamp,
			}
			if fbMsg.Content != nil {
				msg.Body = fbMsg.Content
			} else if fbMsg.Delivery != nil {
				msg.Body = fbMsg.Delivery
			} else if fbMsg.Postback != nil {
				msg.Body = fbMsg.Postback
			}

			messages = append(messages, msg)
		}
	}
	return
}

// send function will unmarshal any object into json string and then
// submit a http request to the facebook messenger api endpoint
func (a *FBAmbassador) send(payload interface{}) (err error) {
	b, err := json.Marshal(payload)
	if err != nil {
		return
	}

	fbApiUrl := FBMessengerBaseURI + a.token
	resp, err := a.client.Post(fbApiUrl, "application/json", bytes.NewBuffer(b))
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
		err = fmt.Errorf("fail to deliver an fb message. status: %s, body: %s",
			resp.Status, buffer.String())
	}
	return
}

// AskQuestion sends a question style text to a recipient.
func (a *FBAmbassador) AskQuestion(recipientId string, text string, answers []map[string]string) (err error) {
	message := map[string]interface{}{
		"text":          text,
		"quick_replies": answers,
	}
	payload := map[string]interface{}{
		"recipient": FBRecipient{recipientId},
		"message":   message,
	}

	err = a.send(payload)
	return
}

// SendText sends a text message to a recipient.
func (a *FBAmbassador) SendText(recipientId string, text string) (err error) {
	message := map[string]string{"text": text}
	payload := map[string]interface{}{
		"recipient": FBRecipient{recipientId},
		"message":   message,
	}

	err = a.send(payload)
	return
}

// SendTemplate sends a template message to a recipient.
func (a *FBAmbassador) SendTemplate(recipientId string, template interface{}) (err error) {
	msgPayload := FBMessageTemplate{
		Type:     "generic",
		Elements: template,
	}

	msgBuf, err := json.Marshal(&msgPayload)
	if err != nil {
		return
	}

	payload := map[string]interface{}{
		"recipient": FBRecipient{recipientId},
		"message": map[string]interface{}{
			"attachment": &FBMessageAttachment{
				Type:    "template",
				Payload: json.RawMessage(msgBuf),
			},
		},
	}

	err = a.send(payload)
	return
}
