package message

import (
	"bytes"
)

type Message struct {
	From    string
	Message string
}

func New(from, message string) *Message {
	m := new(Message)
	m.From = from
	m.Message = message
	return m
}

// Deserialize a byte representation of the message.
func DeSerialize(packet []byte) *Message {
	from := bytes.IndexByte(packet, '\n')
	if from == -1 {
		return nil
	}

	m := new(Message)
	m.From = string(packet[:from])
	m.Message = string(packet[from+1:])
	return m
}

// Serialize the message for transport.
func (m *Message) Serialize() []byte {
	return []byte(m.From + "\n" + m.Message)
}

// Getter for the from field. Suitable for use on mobile.
func (m *Message) GetFrom() string {
	return m.From
}

// Getter for the message field. Suitable for use on mobile.
func (m *Message) GetMessage() string {
	return m.Message
}
