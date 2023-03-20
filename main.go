package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/jlaundry/qtbot/discord"
	"github.com/jlaundry/qtbot/log_analytics"
)

type MQTTServerConfig struct {
	Address  string `json:"address"`
	Username string `json:"username"`
	Password string `json:"password"`
	ClientID string `json:"client_id"`
}

type Config struct {
	MQTT         MQTTServerConfig                 `json:"mqtt_server"`
	LogAnalytics log_analytics.LogAnalyticsConfig `json:"log_analytics"`
	Discord      discord.DiscordConfig            `json:"discord"`
	Debug        bool                             `json:"debug"`
}

var config Config

func main() {
	// Setup SIGINT/SIGTERM handling
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		fmt.Println()
		fmt.Println(sig)
		done <- true
	}()

	// Load config
	jsonFile, err := os.Open("config.json")
	if err != nil {
		panic(err)
	}
	byt, _ := io.ReadAll(jsonFile)
	jsonFile.Close()

	if err := json.Unmarshal(byt, &config); err != nil {
		panic(err)
	}

	if config.MQTT.ClientID == "" {
		config.MQTT.ClientID = "qtbot"
	}

	opts := mqtt.NewClientOptions()
	opts.AddBroker(config.MQTT.Address)
	opts.SetClientID(config.MQTT.ClientID)
	if config.MQTT.Username != "" {
		opts.SetUsername(config.MQTT.Username)
		opts.SetPassword(config.MQTT.Password)
	}

	opts.SetDefaultPublishHandler(func(c mqtt.Client, m mqtt.Message) {
		if config.Debug {
			log.Printf("<- %s: %s\n", m.Topic(), m.Payload())
		}
	})

	opts.OnConnect = func(c mqtt.Client) {
		log.Println("MQTT connected")
	}
	opts.OnConnectionLost = func(c mqtt.Client, err error) {
		log.Fatalf("MQTT connectionLost: %e", err)
	}

	client := mqtt.NewClient(opts)
	fmt.Printf("Connecting to %s\n", config.MQTT.Address)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal(token.Error())
	}

	fmt.Println("Connected")
	defer client.Disconnect(2000)
	defer log.Println("Disconnecting MQTT")

	// Send wakeup messages

	// Listeners for Discord

	// Listeners for Log Analytics

	// Send Ready
	client.Publish(fmt.Sprintf("%s/online", config.MQTT.ClientID), 0, false, "true")

	logmsg := fmt.Sprintf("online, time is: %s", time.Now().UTC().Format(time.RFC3339Nano))
	logtopic := fmt.Sprintf("%s/log", config.MQTT.ClientID)
	client.Publish(logtopic, 0, false, logmsg)

	// HTTP endpoints

	<-done
	fmt.Println("SIGINT/SIGTERM received, exiting")
	client.Publish(fmt.Sprintf("%s/online", config.MQTT.ClientID), 0, false, "false")
}
