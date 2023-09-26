package log_analytics

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestLogAnalytics200(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(``))
		}))
	defer server.Close()

	log.Print(server.URL)

	entry := NewLogEntry(time.Now(), "topic", "message")
	config := LogAnalyticsConfig{
		URL:           server.URL,
		WorkspaceId:   "test",
		SharedKey:     "test",
		CustomLogName: "MQTTLog",
	}

	config.Post(entry)

}

func TestLogAnalytics503(t *testing.T) {

	retries_before_200 := 3

	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			retries_before_200--

			if retries_before_200 <= 0 {
				w.WriteHeader(200)
			}

			w.WriteHeader(503)
		}))
	defer server.Close()

	entry := NewLogEntry(time.Now(), "topic", "message")
	config := LogAnalyticsConfig{
		URL:           server.URL,
		WorkspaceId:   "test",
		SharedKey:     "test",
		CustomLogName: "MQTTLog",
	}

	config.Post(entry)
}

func TestLogAnalytics500(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(500)
		}))
	defer server.Close()

	entry := NewLogEntry(time.Now(), "topic", "message")
	config := LogAnalyticsConfig{
		URL:           server.URL,
		WorkspaceId:   "test",
		SharedKey:     "test",
		CustomLogName: "MQTTLog",
	}

	err := config.Post(entry)
	fmt.Printf("Err was: %e", err)
}
