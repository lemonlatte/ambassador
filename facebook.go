package ambassador

import (
	"encoding/json"
)

type FBObject struct {
	Object string
	Entry  []FBEntry
}

type FBEntry struct {
	Id        string
	Time      int64
	Messaging []FBMessage
}

type FBSender struct {
	Id int64 `json:"id,string"`
}

type FBRecipient struct {
	Id int64 `json:"id,string"`
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
	Type     string          `json:"template_type"`
	Elements json.RawMessage `json:"elements"`
}

type FBButtonItem struct {
	Type    string `json:"type"`
	Title   string `json:"title"`
	Url     string `json:"url,omitempty"`
	Payload string `json:"payload,omitempty"`
}

type FBAmbassador struct {
	token string
}

func (a *FBAmbassador) Translate(b []byte) (messages []Message, err error) {
	return
}

func (a *FBAmbassador) SendText(text string) (err error) {
	return
}

func (a *FBAmbassador) SendTemplate(b []byte) (err error) {
	return
}
