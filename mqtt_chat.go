package mqtt_chat

import (
	"net/url"
	"crypto/tls"
	"github.com/yosssi/gmq/mqtt"
	"github.com/yosssi/gmq/mqtt/client"
)

type EventType int;
const (
	EventError        = iota;
	EventMessage      = iota;
	EventConnected    = iota;
	EventDisconnected = iota;
);

type Event struct{
	Type EventType;
	Value interface{};
	Error error;
};

type Connection struct{
	mqtt_client client.Client
	address string
	tls bool
	events chan Event
};

type Auth struct {
	Username string
	Password string
}

func Dial(address string, auth *Auth) (Connection, error) {
	conn := make(Connection);

	a, err := url.Parse(address);
	if (a.Scheme == "mqtts"){
		conn.tls = true;
	}

	connOpts := &client.ConnectOptions{
		Network   : "tcp",
		Address   : a.Host,
	};

	if (conn.tls){
		connOpts.TLSConfig = &tls.Config{};
	}

	if (auth != nil){
		connOpts.UserName = []byte(auth.Username);
		connOpts.Password = []byte(auth.Password);
	}

	conn.events = make(chan event, 1);

	conn.mqtt_client := client.New(&client.Options{
		// Define the processing of the error handler.
		ErrorHandler: func(err error) {
			conn.events <- &Event{
				Type: EventError,
				Error : error,
			}
		},
	});

	err := cli.Connect(conOpts);

	if err != nil {
		conn.events <- &Event{
			Type : EventError,
			Error: error,
		};
		return conn, error;
	}


	reader := bufio.NewReader(os.Stdin)
	for line, err := reader.ReadString('\n'); err != io.EOF; line, err = reader.ReadString('\n') {
		fmt.Println(line)
	}
}


func (conn *Connection) Subscribe(topic string, qos int){
	qualities = []byte{
		mqtt.QoS0, 
		mqtt.QoS1, 
		mqtt.QoS2,
	};

	conn.mqtt_client.Subscribe(&client.SubscribeOptions{
		SubReqs: []*client.SubReq{
			&client.SubReq{
				TopicFilter: []byte(topic),
				QoS:         qualities[qos],
				Handler: func(topicName, message []byte) {
					conn.events <- &Event{
						Type : EventMessage,
						value : &interface {
							Topic : topicName,
							message : message
						},
					};
				},
			},
		},
	})
}

func (conn * Connection) Publish(topic, message string, qos int, retain bool){
	qualities = []byte{
		mqtt.QoS0, 
		mqtt.QoS1, 
		mqtt.QoS2,
	};

}
