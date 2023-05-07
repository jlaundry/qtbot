# qtbot

A simple golang app that helps with coordinating messages via MQTT, to help with home automation use-cases, including:

  * Logging channel messages to Azure Log Analytics and/or Discord
  * (more to come...)

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

    "discord": [
        {
            "topic": "alert/#",
            "webhook": "https://discord.com/api/webhooks/1/your_key_here"
        },
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
