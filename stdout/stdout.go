package stdout

import (
	"fmt"

	"github.com/jlaundry/qtbot/timestamped_message"
)

type StdoutConfig struct {
	Topic string `json:"topic"`
}

func (config *StdoutConfig) Start(queue <-chan timestamped_message.TimestampedMessage) {
	go func(queue <-chan timestamped_message.TimestampedMessage) {
		for rawMsg := range queue {
			fmt.Printf("%s %s => %s\n", rawMsg.Timestamp, rawMsg.Topic(), string(rawMsg.Payload()))
		}
	}(queue)
}
