package go_mqtt_chat

import (
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

func New(username, server string, auth *mqtt.Auth) *Client {
	c := new(Client)
	c.username = username
	c.server = server
	c.auth = auth
	c.messages = make(chan event.MessageEvent)
	c.keys = make(chan event.MessageEvent)
	return c
}

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

func (c *Client) Say(username, message string) error {
	topic := "/users/" + username + "/messages"
	outgoing_message := c.createMessage(username, message)
	err := c.connection.Publish(topic, outgoing_message, 0, false)
	return err
}

func (c *Client) Listen() string {
	m := <-c.messages
	return string(m.Message)
}

func (c *Client) createMessage(username, message string) []byte {
	// this will encrypt later
	return []byte(c.username + ": " + message)
}
