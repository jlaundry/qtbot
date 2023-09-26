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
	"github.com/jlaundry/qtbot/pagerduty"
	"github.com/jlaundry/qtbot/timestamped_message"
)

type MQTTServerConfig struct {
	Address  string `json:"address"`
	Username string `json:"username"`
	Password string `json:"password"`
	ClientID string `json:"client_id"`
}

type OnStart struct {
	Topic   string `json:"topic"`
	Message string `json:"message"`
}

type Config struct {
	Debug        bool                               `json:"debug"`
	MQTT         MQTTServerConfig                   `json:"mqtt_server"`
	OnStart      []OnStart                          `json:"on_start"`
	Discord      []discord.DiscordConfig            `json:"discord"`
	PagerDuty    []pagerduty.PagerDutyConfig        `json:"pagerduty"`
	LogAnalytics []log_analytics.LogAnalyticsConfig `json:"log_analytics"`
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
	jsonFile, err := os.Open("qtbot.json")
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

	// if config.Debug {
	// 	mqtt.ERROR = log.New(os.Stdout, "[ERROR] ", 0)
	// 	mqtt.CRITICAL = log.New(os.Stdout, "[CRIT] ", 0)
	// 	mqtt.WARN = log.New(os.Stdout, "[WARN]  ", 0)
	// 	mqtt.DEBUG = log.New(os.Stdout, "[DEBUG] ", 0)
	// }

	opts := mqtt.NewClientOptions()

	opts.AddBroker(config.MQTT.Address)
	opts.SetClientID(config.MQTT.ClientID)
	if config.MQTT.Username != "" {
		opts.SetUsername(config.MQTT.Username)
		opts.SetPassword(config.MQTT.Password)
	}

	opts.SetDefaultPublishHandler(func(c mqtt.Client, m mqtt.Message) {
		log.Printf("DEBUG defaultHandler <- %s: %s\n", m.Topic(), m.Payload())
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

	defer client.Disconnect(2000)
	defer log.Println("Disconnecting MQTT")

	// Listeners for Discord
	for i := range config.Discord {
		processor := config.Discord[i]
		queue := make(chan timestamped_message.TimestampedMessage)
		log.Printf("created queue %v for topic %s", queue, processor.Topic)
		defer close(queue)

		token := client.Subscribe(processor.Topic, 2, func(c mqtt.Client, m mqtt.Message) {
			// log.Printf("Discord (%s) %s: %s", processor.Topic, m.Topic(), m.Payload())
			queue <- timestamped_message.NewTimestampedMessage(m)
		})
		token.Wait()

		processor.Start(queue)
	}

	// Listeners for PagerDuty
	for i := range config.PagerDuty {
		processor := config.PagerDuty[i]
		queue := make(chan timestamped_message.TimestampedMessage)
		log.Printf("created queue %v for topic %s", queue, processor.Topic)
		defer close(queue)

		token := client.Subscribe(processor.Topic, 2, func(c mqtt.Client, m mqtt.Message) {
			// log.Printf("PagerDuty (%s) %s: %s", processor.Topic, m.Topic(), m.Payload())
			queue <- timestamped_message.NewTimestampedMessage(m)
		})
		token.Wait()

		processor.Start(queue)
	}

	// Listeners for Log Analytics
	for i := range config.LogAnalytics {
		processor := config.LogAnalytics[i]
		queue := make(chan timestamped_message.TimestampedMessage)
		log.Printf("created queue %v for topic %s", queue, processor.Topic)
		defer close(queue)

		token := client.Subscribe(processor.Topic, 2, func(c mqtt.Client, m mqtt.Message) {
			// log.Printf("LogAnalytics (%s) %s: %s", processor.Topic, m.Topic(), m.Payload())
			queue <- timestamped_message.NewTimestampedMessage(m)
		})
		token.Wait()

		processor.Start(queue)
	}

	// Send wakeup messages
	for i := range config.OnStart {
		on_start := config.OnStart[i]

		log.Printf("DEBUG Sending on_start message '%s' to %s", on_start.Message, on_start.Topic)
		PublishWithLogging(client, on_start.Topic, on_start.Message)
	}

	// Send Ready
	PublishWithLogging(client, fmt.Sprintf("%s/online", config.MQTT.ClientID), "true")

	logmsg := fmt.Sprintf("online, time is: %s", time.Now().UTC().Format(time.RFC3339Nano))
	logtopic := fmt.Sprintf("%s/log", config.MQTT.ClientID)
	PublishWithLogging(client, logtopic, logmsg)

	// HTTP endpoints

	<-done
	fmt.Println("SIGINT/SIGTERM received, exiting")
	PublishWithLogging(client, fmt.Sprintf("%s/online", config.MQTT.ClientID), "false")
}

func PublishWithLogging(client mqtt.Client, topic string, message string) {
	token := client.Publish(topic, 0, false, message)
	go func() {
		_ = token.Wait() // Can also use '<-t.Done()' in releases > 1.2.0
		if token.Error() != nil {
			log.Printf("ERROR Failed to send '%s' to %s: %s",
				message,
				topic,
				token.Error(),
			)
		}
	}()
}
