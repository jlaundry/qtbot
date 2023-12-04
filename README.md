# qtbot

A simple golang app that helps with coordinating messages via MQTT, to help with home automation use-cases, including:

  * Sending startup messages (i.e., because you keep forgetting to turn off debug outputs)
  * Logging messages to Azure Log Analytics
  * Logging messages to Discord

## Setup

Create a `qtbot.json` file like the below:

```json
{
    "debug": false,
    "mqtt_server": {
        "address": "tcp://localhost:1883",
        "username": "qtbot",
        "password": "changeme",
        "client_id": "qtbot"  // Optional, defaults to qtbot
    },
    "on_start": [
        {
            "topic": "appliance/debug/set",
            "message": "off"
        }
    ],

    "stdout": [
        {
            "topic": "#"
        },
    ],
    "discord": [
        {
            "topic": "alert/#",
            "webhook": "https://discord.com/api/webhooks/1/your_key_here"
        },
    ],
    "pagerduty": [
        {
            "topic": "alert/#",
            "severity": "warning",
            "integration_key": "00000000000000000000000000000000",
            "url": "https://events.pagerduty.com/v2/enqueue"
        }
    ],
    "log_analytics": [
        {
            "topic": "#",
            "workspace_id": "00000000-0000-0000-0000-000000000000",
            "shared_key": "",
            "custom_log_name": "MQTTLog"
        },
    ]
}
```
