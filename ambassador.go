package ambassador

type Message struct {
	SenderId    string
	RecipientId string
	Timestamp   int64
	Body        interface{}
}

type Ambassador interface {
	Translate(b []byte) (messages []Message, err error)
	SendText(text string) (err error)
	SendTemplate(b []byte) (err error)
}

func New(source, token string) (a Ambassador) {
	switch source {
	case "facebook":
		return &FBAmbassador{token}
	}
	return
}
