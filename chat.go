package go_mqtt_chat

import (
	"github.com/MobiusHorizons/go_mqtt_chat/message"
	"github.com/MobiusHorizons/go_mqtt_chat/mqtt"
	"github.com/MobiusHorizons/go_mqtt_chat/mqtt/event"
)

type Client struct {
	connection *mqtt.Connection
	server     string
	username   string
	auth       *mqtt.Auth
	messages   chan event.MessageEvent
	keys       chan event.MessageEvent
}

// New creates a new client object which can be used to connect and chat
//  - username : the nickname you use on the server
//  - server   : the connection strin for the server (ex: mqtt://test.mosquitto.org:1883)
//  - auth     : Authentication informantion Username/Password required by some connections
func New(nickname, server string, auth *mqtt.Auth) *Client {
	c := new(Client)
	c.username = nickname
	c.server = server
	c.auth = auth
	c.messages = make(chan event.MessageEvent)
	c.keys = make(chan event.MessageEvent)
	return c
}

// Connect to server using the client object created by `New`.
// This function must be called before using any of the other methods.
func (c *Client) Connect() error {
	connection, err := mqtt.Dial(c.server, c.auth)
	if err != nil {
		return err
	}
	c.connection = connection
	// set up listeners
	connection.Subscribe("/users/"+c.username+"/messages", 0, c.messages)
	//  connection.Subscribe("/users/" + c.username+"/keys", 0, c.keys)
	if err != nil {
		return err
	}

	return nil
}

// Say sends a message to another user
// Params :
//    - username : username to send message to
//    - messsage : message to send
func (c *Client) Say(username, body string) error {
	topic := "/users/" + username + "/messages"
	outgoing_message := c.createMessage(username, body)
	err := c.connection.Publish(topic, outgoing_message, 0, false)
	if err != nil {
		return err
	}
	err = c.connection.Publish("/users/"+c.username+"/messages", outgoing_message, 0, false)
	return err
}

// Listen blocks until a message is received.
func (c *Client) Listen() *message.Message {
	m := <-c.messages
	return message.DeSerialize(m.Message)
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
	c.connection.Terminate()
}

func (c *Client) createMessage(username, body string) []byte {
	// this will encrypt later
	m := message.New(username, body)
	return m.Serialize()
}
