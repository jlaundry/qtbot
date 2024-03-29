package pagerduty

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestPagerDuty204(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(204)
		}))
	defer server.Close()

	config := PagerDutyConfig{
		"topic/example",
		"critical",
		"fake_integration_key",
		server.URL,
	}

	payload := NewPagerDutyPayload(time.Now(), config.Topic, "message", config.Severity)
	alert := NewPagerDutyAlert(config.IntegrationKey, "trigger", payload)
	alert.Post(server.URL)
}

func TestPagerDuty500(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(500)
		}))
	defer server.Close()

	config := PagerDutyConfig{
		"topic/example",
		"critical",
		"fake_integration_key",
		server.URL,
	}

	payload := NewPagerDutyPayload(time.Now(), config.Topic, "message", config.Severity)
	alert := NewPagerDutyAlert(config.IntegrationKey, "trigger", payload)
	err := alert.Post(server.URL)
	fmt.Printf("Err was: %e", err)
}
