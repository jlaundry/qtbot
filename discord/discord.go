package discord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/jlaundry/qtbot/timestamped_message"
)

type DiscordConfig struct {
	Topic   string `json:"topic"`
	Webhook string `json:"webhook"`
}

type discordWebhookMessage struct {
	Content string `json:"content"`
}

func NewDiscordWebhookMessage(timestamp time.Time, topic string, message string) discordWebhookMessage {
	dateString := timestamp.Format("15:04:05")
	return discordWebhookMessage{
		Content: fmt.Sprintf("%s %s: `%s`", dateString, topic, message),
	}
}

func (msg *discordWebhookMessage) Serialize() []byte {
	jsonMsg, _ := json.Marshal(msg)
	return []byte(jsonMsg)
}

func (msg *discordWebhookMessage) Post(url string) error {
	client := &http.Client{}

	for {
		req, err := http.NewRequest("POST", url, bytes.NewReader(msg.Serialize()))
		if err != nil {
			log.Fatal(err)
		}

		req.Header.Add("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == 204 {
			return nil
		} else if resp.StatusCode == 429 {
			resetafter, err := strconv.ParseFloat(resp.Header.Get("X-RateLimit-Reset-After"), 32)
			if err != nil {
				resetafter = 2.0
			}
			if resetafter == 0.0 {
				resetafter = 3.0
			}
			sleepfor := time.Duration(resetafter) * time.Second
			log.Printf("%s (%d): sleeping for %s", url, resp.StatusCode, sleepfor)
			time.Sleep(sleepfor)
		} else {
			return fmt.Errorf("%s (%d): \n\nPOSTdata was: %s", url, resp.StatusCode, msg.Serialize())
		}
	}
}

func (config *DiscordConfig) Start(queue <-chan timestamped_message.TimestampedMessage) {
	go func(queue <-chan timestamped_message.TimestampedMessage) {

		for rawMsg := range queue {
			msg := NewDiscordWebhookMessage(rawMsg.Timestamp, rawMsg.Topic(), string(rawMsg.Payload()))
			err := msg.Post(config.Webhook)
			if err != nil {
				log.Fatal(err)
			}
		}
	}(queue)
}
