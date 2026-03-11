# Minion Monitor

A simple, lightweight Go CLI tool to monitor your infrastructure via SSH and HTTP. It generates reports in various formats (Text, Markdown, JSON, YAML) and can send alerts to Telegram.

## Features

- **Server Monitoring (SSH):**
  - OS & Kernel info, Logged-in users.
  - Disk usage (human-readable).
  - Memory & CPU usage.
  - Load average & Uptime.
  - Clock alignment check (warning if drift > 60s).
  - Pending Apt updates (with security updates count).
  - Docker container statuses & uptime.
  - Optional historical metrics via `sar`.
- **Web App Monitoring:**
  - HTTP Status checks.
  - SSL Certificate expiration warnings (15-day threshold).
- **Flexible Reporting:**
  - Outputs in Text, Markdown, JSON, or YAML.
  - Telegram notifications (Always or Only on Error).
- **Automation Ready:**
  - Supports `--hard-fail` for CI/CD integration.
  - Simple YAML/TOML configuration.

## Installation

### Via Go Install (Recommended)
You can install the tool globally on your system:
```bash
go install github.com/mdnmdn/minion-monitor@latest
```
*Note: Ensure your `$GOPATH/bin` is in your `$PATH`.*

### From Source
```bash
just build
```
Or manually:
```bash
go build -o bin/minion-mon .
```

## Quick Start

1. Create a `hosts.yaml` (see `sample-hosts.yaml`).
2. Run the monitor:
   ```bash
   ./bin/minion-mon
   ```
3. Generate a Markdown report:
   ```bash
   ./bin/minion-mon --format markdown
   ```

## Documentation

Detailed documentation can be found in the `_docs/` folder:
- [Monitoring Checks](_docs/checks.md)
- [Setup & Usage](_docs/setup.md)
- [Telegram Setup](_docs/telegram.md)

## License
MIT
