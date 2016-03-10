package go_mqtt_chat

import (
	"github.com/MobiusHorizons/go_mqtt_chat/crypto"
	"github.com/MobiusHorizons/go_mqtt_chat/message"
	"github.com/MobiusHorizons/go_mqtt_chat/mqtt"
	"github.com/MobiusHorizons/go_mqtt_chat/mqtt/event"
	"os"
)

type Client struct {
	connection *mqtt.Connection
	server     string
	username   string
	auth       *mqtt.Auth
	messages   chan event.MessageEvent
	presence   chan event.MessageEvent
	keys       chan event.MessageEvent
	pgp        *crypto.Crypto
}

// New creates a new client object which can be used to connect and chat
//  - username  : the nickname you use on the server
//  - server    : the connection strin for the server (ex: mqtt://test.mosquitto.org:1883)
//  - cryptoPwd : the passphrase for unlocking your pgp key.
//  - auth      : Authentication informantion Username/Password required by some connections
func New(nickname, server, cryptoPwd string, auth *mqtt.Auth) *Client {
	var err error
	c := new(Client)
	c.username = nickname
	c.server = server
	c.auth = auth
	c.messages = make(chan event.MessageEvent)
	c.presence = make(chan event.MessageEvent)
	c.keys = make(chan event.MessageEvent)
	c.pgp, err = crypto.New(os.Getenv("HOME"), cryptoPwd)
	if err != nil {
		panic(err)
	}

	return c
}

// Connect to server using the client object created by `New`.
// This function must be called before using any of the other methods.
func (c *Client) Connect() error {
	lwt := &mqtt.LWT{
		Topic:   []byte("users/" + c.username + "/presence"),
		Message: []byte("offline"),
		Retain:  true,
		Qos:     0,
	}

	connection, err := mqtt.Dial(c.server, c.auth, lwt)
	if err != nil {
		return err
	}
	c.connection = connection
	// set up listeners
	connection.Subscribe("users/"+c.username+"/messages", 0, c.messages)
	connection.Subscribe("users/+/presence", 0, c.presence)

	go func() {
		for m := range c.keys {
			username := m.Topic[len("users/") : len(m.Topic)-len("/public-key")]
			err := c.pgp.AddKey(username, m.Message)
			if err != nil {
				panic(err)
			}
		}
	}()

	if err != nil {
		return err
	}
	c.connection.Publish("users/"+c.username+"/public-key", c.pgp.PublicKey(), 0, true)
	c.connection.Publish("users/"+c.username+"/presence", []byte("online"), 0, true)
	return nil
}

// Say sends a message to another user
// Params :
//    - username : username to send message to
//    - messsage : message to send
func (c *Client) Say(username, body string) error {
	topic := "users/" + username + "/messages"
	outgoing_message := c.createMessage(username, body)
	err := c.connection.Publish(topic, outgoing_message, 0, false)
	if err != nil {
		return err
	}
	loop_message := c.createMessage(c.username, body)
	err = c.connection.Publish("users/"+c.username+"/messages", loop_message, 0, false)
	return err
}

// Presence blocks until presence info is received.
func (c *Client) Presence() *message.Message {
	m := <-c.presence
	username := m.Topic[len("users/") : len(m.Topic)-len("/presence")]
	value := m.Message
	return message.New(username, string(value))
}

// Listen blocks until a message is received.
func (c *Client) Listen() *message.Message {
	m := <-c.messages
	payload, err := c.pgp.Decrypt(m.Message)
	if err != nil {
		panic(err)
	}
	return message.DeSerialize(payload)
}

// Connect with another user.
// This function is used to get the user's public key.
func (c *Client) Meet(username string) error {
	topic := "users/" + username + "/public-key"
	c.connection.Subscribe(topic, 0, c.keys)
	return nil
}

// ListenNonBlocking will return a pedinng message if one is available.
// If no message is pedning 'nil' is returned
func (c *Client) ListenNonBlocking() *message.Message {
	select {
	case msg := <-c.messages:
		m := message.DeSerialize(msg.Message)
		return m
	default:
		return nil
	}
}

// Disconnect closes the connection with the server.
func (c *Client) Disconnect() {
	c.connection.Publish("users/"+c.username+"/presence", []byte("offline"), 0, true)
	c.connection.Terminate()
}

func (c *Client) createMessage(username, body string) []byte {
	m := message.New(c.username, body)
	cipherText, err := c.pgp.EncryptFor(username, m.Serialize())
	if err != nil {
		// TODO: manage returning errors from here.
		panic(err)
	}
	return cipherText
}
