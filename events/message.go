package message

type MessageEvent {
	Topic string
	Message []byte
}

func (m * MessageEvent) String(){

