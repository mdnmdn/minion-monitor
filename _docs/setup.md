# Setup & Usage

## Installation

1. Clone the repository.
2. Ensure you have Go installed (1.18+).
3. Install dependencies:
   ```bash
   go get gopkg.in/yaml.v3 github.com/pelletier/go-toml/v2 golang.org/x/crypto/ssh
   ```
4. Build the tool:
   ```bash
   go build -o minion-mon
   ```

## Running the tool

Run with the default `hosts.yaml`:
```bash
./minion-mon
```

Run with a specific config file (YAML or TOML):
```bash
./minion-mon -config my-hosts.toml
```

Enable verbose mode (show top processes):
```bash
./minion-mon -v
```

## Configuration

The tool supports YAML and TOML configurations. Refer to `hosts.yaml` for a complete example.

Key configuration fields:
- `project.name`: Name used in reports.
- `servers`: Map of server aliases.
  - `host`: Hostname or IP.
  - `credentials`: `user`, `ssh-key`, `password`.
  - `webapps`: Map of web application URLs.
  - `docker.status`: Whether to check for running containers.
- `alert.telegram`:
  - `enabled`: true/false.
  - `mode`: `always` or `error`.
  - `token`: Telegram Bot API token.
  - `chat-id`: Target chat ID.
