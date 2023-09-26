package discord

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestDiscord204(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(204)
	}))
	defer svr.Close()

	msg := NewDiscordWebhookMessage(time.Now(), "topic", "message")
	msg.Post(svr.URL)
}

func TestDiscord500(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))
	defer svr.Close()

	msg := NewDiscordWebhookMessage(time.Now(), "topic", "message")
	err := msg.Post(svr.URL)
	fmt.Printf("Err was: %e", err)
}
