package pagerduty

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

type PagerDutyConfig struct {
	Topic          string `json:"topic"`
	Severity       string `json:"severity"`
	IntegrationKey string `json:"integration_key"`
	URL            string `json:"url"`
}

type PagerDutyPayload struct {
	Summary   string      `json:"summary"`
	Source    string      `json:"source"`
	Severity  string      `json:"severity"`
	Timestamp string      `json:"timestamp,omitempty"`
	Component string      `json:"component,omitempty"`
	Group     string      `json:"group,omitempty"`
	Class     string      `json:"class,omitempty"`
	Details   interface{} `json:"custom_details,omitempty"`
}

type PagerDutyAlert struct {
	RoutingKey string           `json:"routing_key"`
	Action     string           `json:"event_action"`
	Payload    PagerDutyPayload `json:"payload,omitempty"`
}

func NewPagerDutyPayload(timestamp time.Time, topic string, message string, severity string) PagerDutyPayload {
	// dateString := timestamp.Format("15:04:05")
	return PagerDutyPayload{
		Summary:   fmt.Sprintf("%s: %s", topic, message),
		Source:    topic,
		Severity:  severity,
		Timestamp: timestamp.Format(time.RFC3339Nano),
	}
}

func NewPagerDutyAlert(routing_key string, action string, payload PagerDutyPayload) PagerDutyAlert {
	return PagerDutyAlert{
		RoutingKey: routing_key,
		Action:     action,
		Payload:    payload,
	}
}

func (msg *PagerDutyAlert) Serialize() []byte {
	jsonMsg, _ := json.Marshal(msg)
	return []byte(jsonMsg)
}

func (msg *PagerDutyAlert) Post(url string) error {
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

		if resp.StatusCode == 202 {
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

func (config *PagerDutyConfig) Start(queue <-chan timestamped_message.TimestampedMessage) {
	go func(queue <-chan timestamped_message.TimestampedMessage, config *PagerDutyConfig) {

		for rawMsg := range queue {
			payload := NewPagerDutyPayload(
				rawMsg.Timestamp,
				rawMsg.Topic(),
				string(rawMsg.Payload()),
				config.Severity,
			)
			msg := NewPagerDutyAlert(
				config.IntegrationKey,
				"trigger",
				payload,
			)
			err := msg.Post(config.URL)
			if err != nil {
				log.Fatal(err)
			}
		}
	}(queue, config)
}
