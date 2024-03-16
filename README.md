# qtbot

A simple golang app that helps with coordinating messages via MQTT, to help with home automation use-cases, including:

  * Sending startup messages (i.e., because you keep forgetting to turn off debug outputs)
  * Logging messages to Azure Log Analytics (with both the new DCR, and/or old MMA APIs)
  * Logging messages to Discord
  * Sending alerts to PagerDuty

## Setup

Create a `/opt/qtbot/qtbot.json` file like the below:

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
    "azure_monitor": [
        {
            "topic": "#",
            "data_collecton_endpoint": "https://qtbot-0000.region-1.ingest.monitor.azure.com",
            "immutable_id": "dcr-00000000000000000000000000000000",
            "stream_name": "Custom-qtbot_CL",
        }
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

Next, create a systemd service:

```ini
[Unit]
Description=QT Bot
Requires=mosquitto.service
After=mosquitto.service network-online.target
Wants=network-online.target

[Service]
Type=simple
WorkingDirectory=/opt/qtbot
User=qtbot
Group=qtbot
ExecStart=/opt/qtbot/qtbot
Restart=always

[Install]
WantedBy=multi-user.target
```

### Azure Monitor (Log Analytics DCR) Setup

Azure Monitor uses the `azidentity.NewDefaultAzureCredential`. If you're using a static Client ID/Secret, you'll need to set the `AZURE_TENANT_ID`, `AZURE_CLIENT_ID`, and `AZURE_CLIENT_SECRET` environment variables as well.

The best way to do this is via `systemctl edit qtbot`, and set

```ini
[Service]
Environment="AZURE_TENANT_ID=00000000-0000-0000-0000-000000000000"
Environment="AZURE_CLIENT_ID=00000000-0000-0000-0000-000000000000"
Environment="AZURE_CLIENT_SECRET=00-00~00~000000000000000000000000000~000"
```
