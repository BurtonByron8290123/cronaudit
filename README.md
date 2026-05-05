# cronaudit

Lightweight daemon that monitors cron job execution and sends alerts on failure or drift.

## Installation

```bash
go install github.com/yourusername/cronaudit@latest
```

Or build from source:

```bash
git clone https://github.com/yourusername/cronaudit.git && cd cronaudit && make build
```

## Usage

Define your monitored jobs in `cronaudit.yaml`:

```yaml
jobs:
  - name: daily-backup
    schedule: "0 2 * * *"
    timeout: 5m
    alert_on:
      - failure
      - drift
    notify:
      slack: "https://hooks.slack.com/services/..."

  - name: hourly-sync
    schedule: "0 * * * *"
    timeout: 30s
    alert_on:
      - failure
```

Start the daemon:

```bash
cronaudit --config cronaudit.yaml
```

Wrap your existing cron commands to report execution status:

```bash
# In your crontab
0 2 * * * cronaudit exec --job daily-backup -- /usr/local/bin/backup.sh
```

Check status of monitored jobs:

```bash
cronaudit status
```

## Configuration

| Field | Description | Default |
|---|---|---|
| `schedule` | Cron expression for expected run time | required |
| `timeout` | Max allowed execution duration | `1m` |
| `drift` | Allowed schedule drift before alerting | `5m` |
| `alert_on` | List of conditions to alert on (`failure`, `drift`, `timeout`) | `[failure]` |
| `notify.slack` | Slack incoming webhook URL for alerts | — |
| `notify.email` | Email address to send alerts to | — |

## License

MIT © [yourusername](https://github.com/yourusername)
