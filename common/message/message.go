package message

type Message struct {
	Typ  Type `json:"type"`
	Data any  `json:"data"`
}

func NewMessage(typ Type, data any) *Message {
	return &Message{
		Typ:  typ,
		Data: data,
	}
}
