package discord

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestDiscord204(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(204)
		}))
	defer server.Close()

	msg := NewDiscordWebhookMessage(time.Now(), "topic", "message")
	msg.Post(server.URL)
}

func TestDiscord500(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(500)
		}))
	defer server.Close()

	msg := NewDiscordWebhookMessage(time.Now(), "topic", "message")
	err := msg.Post(server.URL)
	fmt.Printf("Err was: %e", err)
}
