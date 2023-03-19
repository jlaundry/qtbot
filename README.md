# qtbot

A simple golang app that helps with coordinating messages via MQTT, to help with home automation use-cases, including:

  * Logging channel messages to Azure Log Analytics and/or Discord
  * Creating REST endpoints to send actions to specific topics (devices), or groups of devices

## Setup

Create a `config.json` file like the below:

```json
{
    "mqtt_server": {
        "address": "tcp://localhost:1883",
        "username": "qtbot",
        "password": "changeme",
        "client_id": "qtbot"  // Optional, defaults to qtbot
    },
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
    ],
    "rest_api": {
        "listen_address": ":8080",
        "devices": [
            {
                "name": "deviceA",
                "actions": ["on", "off", "toggle"],
                "topic": "devices/deviceA"
            }
        ],
        "groups": [
            {
                "name": "groupA",
                "actions": ["on", "off", "toggle"],
                "devices": ["deviceA"]
            }
        ],
    }
}
```
