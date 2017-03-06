package ambassador

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
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
	Read      *FBMessageRead     `json:"read,omitempty"`
}

type FBMessageContent struct {
	Text        string                `json:"text"`
	Seq         int64                 `json:"seq,omitempty"`
	IsEcho      bool                  `json:"is_echo,omitempty"`
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

type FBMessageRead struct {
	Watermark int64 `json:"watermark"`
	Seq       int64 `json:"seq"`
}

type FBButtonItem struct {
	Type        string `json:"type"`
	Title       string `json:"title,omitempty"`
	Url         string `json:"url,omitempty"`
	Payload     string `json:"payload,omitempty"`
	HeightRatio string `json:"webview_height_ratio,omitempty"`
	Extensions  bool   `json:"messenger_extensions,omitempty"`
}

type FBAmbassador struct {
	sync.Mutex
	token        string
	client       *http.Client
	messages     []interface{}
	lastMessages []interface{}
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
				if attachments := fbMsg.Content.Attachments; len(attachments) != 0 {
					a := attachments[0]
					if a.Type == "location" {
						payload := FBLocationAttachment{}
						err = json.Unmarshal(a.Payload, &payload)
						if err != nil {
							return
						}
						msg.Content = &LocationContent{
							Lat: payload.Coordinates.Latitude,
							Lon: payload.Coordinates.Longitude,
						}
					} else {
						msg.Content = fbMsg.Content
					}
				} else if fbMsg.Content.QuickReplay != nil {
					msg.Content = &CommandContent{Payload: fbMsg.Content.QuickReplay.Payload}
				} else if fbMsg.Content.IsEcho {
					msg.Content = fbMsg.Content
				} else {
					msg.Content = &TextContent{Text: fbMsg.Content.Text}
				}
			} else if fbMsg.Delivery != nil {
				msg.Content = fbMsg.Delivery
			} else if fbMsg.Postback != nil {
				msg.Content = &CommandContent{Payload: fbMsg.Postback.Payload}
			} else if fbMsg.Read != nil {
				msg.Content = fbMsg.Read
			}
			messages = append(messages, msg)
		}
	}
	return
}

// send function will unmarshal any object into json string and then
// submit a http request to the facebook messenger api endpoint
func (a *FBAmbassador) sendMessages(recipientId string) (err error) {
	fbApiUrl := FBMessengerBaseURI + a.token

	for _, msgPayload := range a.messages {
		payload, ok := msgPayload.(map[string]interface{})
		if !ok {
			return fmt.Errorf("fail to type assert message: %+v", msgPayload)
		}
		payload["recipient"] = FBRecipient{recipientId}

		b, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		resp, err := a.client.Post(fbApiUrl, "application/json", bytes.NewBuffer(b))
		if err != nil {
			return err
		}

		if resp.StatusCode != 200 {
			buffer := &bytes.Buffer{}
			_, err := io.Copy(buffer, resp.Body)
			resp.Body.Close()

			if err != nil {
				return err
			}
			return fmt.Errorf("fail to deliver an fb message. status: %s, body: %s",
				resp.Status, buffer.String())
		}
		resp.Body.Close()
	}
	return
}

// AskQuestion sends a question style text to a recipient.
func (a *FBAmbassador) AskQuestion(text string, answers []map[string]string) (err error) {
	message := map[string]interface{}{
		"text":          text,
		"quick_replies": answers,
	}
	payload := map[string]interface{}{
		"message": message,
	}

	a.Lock()
	defer a.Unlock()
	a.messages = append(a.messages, payload)
	return
}

// SendText sends a text message to a recipient.
func (a *FBAmbassador) SendText(text string) (err error) {
	message := map[string]string{"text": text}
	payload := map[string]interface{}{
		"message": message,
	}

	a.Lock()
	defer a.Unlock()
	a.messages = append(a.messages, payload)
	return
}

// SendTemplate sends a template message to a recipient.
func (a *FBAmbassador) SendTemplate(elements interface{}) (err error) {

	columns := []map[string]interface{}{}
	colItems, ok := elements.([]Carousel)
	if !ok {
		return fmt.Errorf("can not type assert the elements")
	}

	for i, col := range colItems {
		if i > 10 {
			break
		}

		element := map[string]interface{}{
			"title":     col.Title,
			"image_url": col.ImageUrl,
			"item_url":  col.ItemUrl,
			"subtitle":  col.Text,
		}
		buttons := []FBButtonItem{}
		for _, btn := range col.Buttons {
			var fbBtn FBButtonItem
			switch btn.Type {
			case "share":
				fbBtn.Type = "element_share"
			case "account_link":
				fbBtn.Type = btn.Type
				fbBtn.Url = btn.Data
			case "url":
				fbBtn.Title = btn.Label
				fbBtn.Type = "web_url"
				fbBtn.Url = btn.Data
				fbBtn.Extensions = btn.Extensions
				fbBtn.HeightRatio = btn.HeightRatio
			}
			buttons = append(buttons, fbBtn)
		}
		if len(buttons) > 0 {
			element["buttons"] = buttons
		}

		columns = append(columns, element)
	}

	msgPayload := FBMessageTemplate{
		Type:     "generic",
		Elements: columns,
	}

	msgBuf, err := json.Marshal(&msgPayload)
	if err != nil {
		return
	}

	payload := map[string]interface{}{
		"message": map[string]interface{}{
			"attachment": &FBMessageAttachment{
				Type:    "template",
				Payload: json.RawMessage(msgBuf),
			},
		},
	}

	a.Lock()
	defer a.Unlock()
	a.messages = append(a.messages, payload)
	return
}

func (a *FBAmbassador) cleanMessage() {
	a.Lock()
	defer a.Unlock()
	a.lastMessages = a.messages
	a.messages = []interface{}{}
}

func (a *FBAmbassador) GetLastSent() []interface{} {
	return a.lastMessages
}

func (a *FBAmbassador) Send(recipientId string) (err error) {
	defer a.cleanMessage()
	err = a.sendMessages(recipientId)
	if err != nil {
		b, _ := json.Marshal(a.messages)
		return fmt.Errorf("%s, %s", err.Error(), b)
	}
	return
}
