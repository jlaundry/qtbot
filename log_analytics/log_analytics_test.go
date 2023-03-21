package log_analytics

import (
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

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
