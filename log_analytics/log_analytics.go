package log_analytics

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jlaundry/qtbot/timestamped_message"
)

const (
	timeStampField = "TimeGenerated"
)

type LogAnalyticsConfig struct {
	URL           string
	Topic         string `json:"topic"`
	WorkspaceId   string `json:"workspace_id"`
	SharedKey     string `json:"shared_key"`
	CustomLogName string `json:"custom_log_name"`
}

type LogEntry struct {
	TimeGenerated string
	Topic         string
	Message       string
}

func NewLogEntry(timestamp time.Time, topic string, message string) LogEntry {
	return LogEntry{
		TimeGenerated: timestamp.UTC().Format("2006-01-02T15:04:05.000Z"), //time.RFC3339Nano),
		Topic:         topic,
		Message:       message,
	}
}

func (log LogEntry) JsonString() string {
	s, _ := json.Marshal(log)
	return string(s)
}

func buildSignature(message, secret string) (string, error) {

	keyBytes, err := base64.StdEncoding.DecodeString(secret)
	if err != nil {
		return "", err
	}

	mac := hmac.New(sha256.New, keyBytes)
	mac.Write([]byte(message))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil)), nil
}

func (config LogAnalyticsConfig) Post(entry LogEntry) error {

	dateString := time.Now().UTC().Format(time.RFC1123)
	dateString = strings.Replace(dateString, "UTC", "GMT", -1)

	data := entry.JsonString()
	stringToHash := "POST\n" + strconv.Itoa(len(data)) + "\napplication/json\n" + "x-ms-date:" + dateString + "\n/api/logs"
	hashedString, err := buildSignature(stringToHash, config.SharedKey)
	if err != nil {
		log.Println(err.Error())
		return err
	}

	signature := fmt.Sprintf("SharedKey %s:%s", config.WorkspaceId, hashedString)
	url := fmt.Sprintf("https://%s.ods.opinsights.azure.com/api/logs?api-version=2016-04-01", config.WorkspaceId)

	// Allow custom URL for testing
	if config.URL != "" {
		url = config.URL
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", url, bytes.NewReader([]byte(data)))
	if err != nil {
		return err
	}

	req.Header.Add("Log-Type", config.CustomLogName)
	req.Header.Add("Authorization", signature)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("x-ms-date", dateString)
	req.Header.Add("time-generated-field", timeStampField)

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		// pass
	} else {
		bodyString, _ := io.ReadAll(resp.Body)
		log.Fatalf("%s (%d): %s\n\nPOSTdata was: %s", url, resp.StatusCode, bodyString, data)
	}

	return err
}

func (config LogAnalyticsConfig) Start(queue <-chan timestamped_message.TimestampedMessage) {
	go func(queue <-chan timestamped_message.TimestampedMessage) {
		for msg := range queue {
			entry := NewLogEntry(msg.Timestamp, msg.Topic(), string(msg.Payload()))
			config.Post(entry)
		}
	}(queue)
}
