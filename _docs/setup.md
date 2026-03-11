# Setup & Usage

## Installation

Ensure you have **Go 1.18+** installed on your system.

### Via Go Install (Recommended)
This is the fastest way to install the tool globally:
```bash
go install github.com/mdnmdn/minion-monitor@latest
```
*Note: Make sure your `go/bin` directory is added to your system's `PATH`.*

### From Source (Clone)
1. Clone the repository.
2. (Optional) Install [just](https://github.com/casey/just).

#### Building
If you have `just` installed:
```bash
just build
```

Otherwise, use the standard Go command:
```bash
go build -o bin/minion-mon .
```

*Note: You do **not** need to manually `go get` dependencies; Go will automatically download them from `go.mod` during the build process.*

## Running the tool

The binary is located in the `bin/` directory after building.

Run with the default `hosts.yaml`:
```bash
./bin/minion-mon
```

### Parameters

- `-c, --config`: Path to config file (YAML or TOML).
- `-f, --format`: Output format (`text`, `markdown`, `json`, `yaml`).
- `-v, --verbose`: Enable verbose mode (shows top 10 processes).
- `--hard-fail`: Exit with a non-zero code if any error/warning is detected (useful for CI/CD).

### Examples

Run with a specific config:
```bash
./bin/minion-mon -c my-hosts.toml
```

Generate a Markdown report:
```bash
./bin/minion-mon --format markdown > report.md
```

## Configuration

The tool supports **YAML** and **TOML** configurations. You can use either format by naming your file appropriately (e.g., `hosts.yaml` or `hosts.toml`).

### Structure Overview

| Field | Description | Required |
|-------|-------------|----------|
| `project.name` | The title of your report. | Yes |
| `servers` | A dictionary where keys are server aliases. | Yes |
| `alert.telegram` | Configuration for notifications. | No |

### Server Configuration

Each server entry in the `servers` map supports:

- **`host`**: The IP address or hostname.
- **`credentials`**:
  - `user`: SSH username.
  - `ssh-key`: Path to your private key (supports `~/`).
  - `password`: (Optional) SSH password if key is not used.
- **`webapps`**: A map of web applications to monitor.
  - `url`: The full HTTP/HTTPS URL.
  - `ignore-certificate`: Set to `true` for self-signed certs.
- **`docker.status`**: Boolean. If `true`, lists running containers and their uptime.
- **`sar.enabled`**: Boolean. If `true`, calculates the 24h average CPU/Mem using `sysstat`.

### Notification Configuration

- **`enabled`**: Boolean.
- **`mode`**: 
  - `always`: Sends a report on every run.
  - `error`: Sends a report **only** if issues are detected (down sites, security updates, etc.).
- **`token`**: Your Telegram Bot API token.
- **`chat-id`**: The numeric ID of the target chat.

### Configuration Examples

#### YAML (`hosts.yaml`)
```yaml
project:
  name: "Production Cluster"
servers:
  web-01:
    host: "1.2.3.4"
    credentials:
      user: "ubuntu"
      ssh-key: "~/.ssh/id_rsa"
    docker:
      status: true
    webapps:
      main-site:
        url: "https://example.com"
alert:
  telegram:
    enabled: true
    mode: "error"
    token: "12345:ABCDE"
    chat-id: "987654"
```

#### TOML (`hosts.toml`)
```toml
[project]
name = "Production Cluster"

[servers.web-01]
host = "1.2.3.4"
docker = { status = true }

[servers.web-01.credentials]
user = "ubuntu"
ssh-key = "~/.ssh/id_rsa"

[servers.web-01.webapps.main-site]
url = "https://example.com"

[alert.telegram]
enabled = true
mode = "error"
token = "12345:ABCDE"
chat-id = "987654"
```
