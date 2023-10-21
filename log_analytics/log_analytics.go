package log_analytics

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/jlaundry/qtbot/timestamped_message"
)

const (
	timeStampField = "TimeGenerated"
	MAX_RETRIES    = 10
	RETRY_WAIT     = 3.0
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

func (log *LogEntry) JsonString() string {
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

func (config *LogAnalyticsConfig) Post(entry LogEntry) error {

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

	client := &http.Client{
		Timeout: time.Second * 60,
	}
	ttl := MAX_RETRIES

	for {
		if ttl <= 0 {
			return fmt.Errorf("ttl exceeded trying to POST to %s after %d attempts \n\nPOSTdata was: %s", url, MAX_RETRIES, data)
		}

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
		if err == nil {
			defer resp.Body.Close()
		}

		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				log.Printf("read timeout on %s: %e", url, err)
				// Let us loop and try again...
			} else if errors.Is(err, syscall.ECONNRESET) {
				log.Printf("connection reset on %s: %e", url, err)
				// Let us loop and try again...
			} else {
				log.Fatalf("read error on %s, %s:", url, err)
			}

		} else if resp.StatusCode == 200 {
			return nil
		} else if resp.StatusCode == 503 {
			sleepfor := time.Duration(RETRY_WAIT) * time.Second
			log.Printf("%s (%d): sleeping for %s (ttl=%d)", url, resp.StatusCode, sleepfor, ttl)
			time.Sleep(sleepfor)
		} else {
			bodyString, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("%s (%d): %s\n\nPOSTdata was: %s", url, resp.StatusCode, bodyString, data)
		}

		// We didn't return, therefore decrement and try again...
		ttl--
	}
}

func (config *LogAnalyticsConfig) Start(queue <-chan timestamped_message.TimestampedMessage) {
	go func(queue <-chan timestamped_message.TimestampedMessage) {
		for msg := range queue {
			entry := NewLogEntry(msg.Timestamp, msg.Topic(), string(msg.Payload()))
			err := config.Post(entry)
			if err != nil {
				log.Fatal(err)
			}
		}
	}(queue)
}
