package azure_monitor

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/monitor/azingest"
	"github.com/jlaundry/qtbot/timestamped_message"
)

const (
	MAX_RETRIES = 3
)

type AzureMonitorConfig struct {
	Topic                  string `json:"topic"`
	DataCollectionEndpoint string `json:"data_collecton_endpoint"`
	ImmutableId            string `json:"immutable_id"`
	StreamName             string `json:"stream_name"`
}

type LogEntry struct {
	TimeGenerated string
	Topic         string
	Message       string
}

func NewLogEntry(tsmsg timestamped_message.TimestampedMessage) LogEntry {
	return LogEntry{
		TimeGenerated: tsmsg.Timestamp.Format(time.RFC3339Nano),
		Topic:         tsmsg.Topic(),
		Message:       string(tsmsg.Payload()),
	}
}

func (msg *LogEntry) Serialize() []byte {
	// The shape of the data must be a JSON array with item structure that matches the format expected by the stream in the DCR.
	// If it is needed to send a single item within API call, the data should be sent as a single-item array.
	// -- https://learn.microsoft.com/en-us/azure/azure-monitor/logs/logs-ingestion-api-overview#body
	var body [1]LogEntry
	body[0] = *msg
	jsonMsg, _ := json.Marshal(body)
	//log.Printf("sending %s", jsonMsg)
	return []byte(jsonMsg)
}

func (config *AzureMonitorConfig) Start(queue <-chan timestamped_message.TimestampedMessage) {
	go func(queue <-chan timestamped_message.TimestampedMessage, config *AzureMonitorConfig) {

		cred, err := azidentity.NewDefaultAzureCredential(nil)
		if err != nil {
			log.Fatalf("Error trying to create azidentity.NewDefaultAzureCredential: %e", err)
		}

		client, err := azingest.NewClient(config.DataCollectionEndpoint, cred, nil)
		if err != nil {
			log.Fatalf("Error trying to create azingest.NewClient: %e", err)
		}
		_ = client

		for rawMsg := range queue {

			msg := NewLogEntry(rawMsg)
			ttl := MAX_RETRIES

			for {
				if ttl <= 0 {
					log.Fatalf("azingest ttl exceeded after %d attempts", MAX_RETRIES)
				}

				_, err := client.Upload(context.TODO(), config.ImmutableId, config.StreamName, msg.Serialize(), nil)
				if err != nil {
					log.Printf("Error sending to azingest client.Upload: %e ttl=%d", err, ttl)
					ttl--
				} else {
					break
				}
			}

		}
	}(queue, config)
}
