package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"github.com/yosssi/gmq/mqtt"
	"github.com/yosssi/gmq/mqtt/client"
	"io"
	"os"
)

func main() {
	cli := client.New(&client.Options{
		// Define the processing of the error handler.
		ErrorHandler: func(err error) {
			fmt.Println(err)
		},
	})
	defer cli.Terminate()
	err := cli.Connect(&client.ConnectOptions{
		Network:   "tcp",
		Address:   "mqtt.mobiushorizons.com:8883",
		TLSConfig: &tls.Config{},
		ClientID:  []byte("mqtt_chat"),
	})

	if err != nil {
		panic(err)
	}

	cli.Subscribe(&client.SubscribeOptions{
		SubReqs: []*client.SubReq{
			&client.SubReq{
				TopicFilter: []byte("chat"),
				QoS:         mqtt.QoS0,
				Handler: func(topicName, message []byte) {
					fmt.Println(string(topicName), string(message))
				},
			},
		},
	})

	reader := bufio.NewReader(os.Stdin)
	for line, err := reader.ReadString('\n'); err != io.EOF; line, err = reader.ReadString('\n') {
		fmt.Println(line)
	}
}
