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
	jsonMsg, _ := json.Marshal(msg)
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
			_, err := client.Upload(context.TODO(), config.ImmutableId, config.StreamName, msg.Serialize(), nil)
			if err != nil {
				log.Fatalf("Error sending to azingest client.Upload: %e", err)
				//TODO: handle error
			}
		}
	}(queue, config)
}
