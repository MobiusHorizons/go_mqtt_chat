package mqtt

import (
	"crypto/tls"
	"github.com/MobiusHorizons/go_mqtt_chat/mqtt/event"
	"github.com/satori/go.uuid"
	"github.com/yosssi/gmq/mqtt"
	"github.com/yosssi/gmq/mqtt/client"
	"net/url"
)

type Connection struct {
	mqtt_client *client.Client
	address     string
	tls         bool
	Errors      chan error
	Messages    chan event.MessageEvent
	Connection  chan event.ConnectionEvent
}

type Auth struct {
	Username string
	Password string
}

type LWT struct {
	Topic   []byte
	Message []byte
	Retain  bool
	Qos     int
}

// used for going from an int to a mqtt qos
var qualities []byte

func Dial(address string, auth *Auth, lwt *LWT) (*Connection, error) {
	qualities = []byte{
		mqtt.QoS0,
		mqtt.QoS1,
		mqtt.QoS2,
	}
	conn := new(Connection)

	a, err := url.Parse(address)
	if a.Scheme == "mqtts" {
		conn.tls = true
	}
	clientID := uuid.NewV4()
	connOpts := &client.ConnectOptions{
		Network:  "tcp",
		Address:  a.Host,
		ClientID: clientID.Bytes(),
	}

	if lwt != nil {
		connOpts.WillTopic = lwt.Topic
		connOpts.WillMessage = lwt.Message
		connOpts.WillRetain = lwt.Retain
		connOpts.WillQoS = qualities[lwt.Qos]
	}

	if conn.tls {
		connOpts.TLSConfig = &tls.Config{}
	}

	if auth != nil {
		connOpts.UserName = []byte(auth.Username)
		connOpts.Password = []byte(auth.Password)
	}

	conn.Errors = make(chan error, 1)
	conn.Messages = make(chan event.MessageEvent)
	conn.Connection = make(chan event.ConnectionEvent, 1)

	cli := client.New(&client.Options{
		// Define the processing of the error handler.
		ErrorHandler: func(err error) {
			conn.Errors <- err
		},
	})

	err = cli.Connect(connOpts)

	if err != nil {
		conn.Errors <- err
		return conn, err
	}

	conn.mqtt_client = cli
	return conn, nil
}

func (conn *Connection) Subscribe(topic string, qos int, out chan event.MessageEvent) {
	conn.mqtt_client.Subscribe(&client.SubscribeOptions{
		SubReqs: []*client.SubReq{
			&client.SubReq{
				TopicFilter: []byte(topic),
				QoS:         qualities[qos],
				Handler: func(topicName, message []byte) {
					m := event.MessageEvent{
						Topic:   string(topicName),
						Message: message,
					}
					if out != nil {
						out <- m
					} else {
						conn.Messages <- m
					}
				},
			},
		},
	})
}

func (conn *Connection) Publish(topic string, message []byte, qos int, retain bool) error {
	err := conn.mqtt_client.Publish(&client.PublishOptions{
		QoS:       qualities[qos],
		Retain:    retain,
		TopicName: []byte(topic),
		Message:   []byte(message),
	})
	return err
}

func (conn *Connection) Terminate() {
	conn.mqtt_client.Terminate()
}
