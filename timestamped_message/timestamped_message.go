package timestamped_message

import (
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type TimestampedMessage struct {
	mqtt.Message
	Timestamp time.Time
}

func NewTimestampedMessage(msg mqtt.Message) TimestampedMessage {
	return TimestampedMessage{
		msg,
		time.Now(),
	}
}
