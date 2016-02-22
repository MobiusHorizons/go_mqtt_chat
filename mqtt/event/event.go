package event

type ConnectionEvent struct {
	Connected bool
}

type MessageEvent struct {
	Topic   string
	Message []byte
}

func (m *MessageEvent) String() string {
	return string(m.Topic) + " " + string(m.Message)
}
