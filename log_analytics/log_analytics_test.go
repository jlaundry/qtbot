package log_analytics

import (
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestLogAnalytics503(t *testing.T) {

	retries_before_200 := 2

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		retries_before_200--

		if retries_before_200 <= 0 {
			w.WriteHeader(200)
		}

		w.WriteHeader(503)
	}))
	defer server.Close()

	timestamp := time.Now()
	entry := NewLogEntry(timestamp, "topic", "message")
	config := LogAnalyticsConfig{
		URL:           server.URL,
		WorkspaceId:   "test",
		SharedKey:     "test",
		CustomLogName: "MQTTLog",
	}

	config.Post(entry)
}

func TestPostLogAnalytics(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(``))
		}))
	defer server.Close()

	log.Print(server.URL)

	timestamp := time.Now()
	entry := NewLogEntry(timestamp, "topic", "message")
	config := LogAnalyticsConfig{
		URL:           server.URL,
		WorkspaceId:   "test",
		SharedKey:     "test",
		CustomLogName: "MQTTLog",
	}

	config.Post(entry)

}
