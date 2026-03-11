# Telegram Notification Setup

To enable Telegram notifications, you need to create a bot and get a chat ID.

## 1. Create a Bot

- Start a chat with `@BotFather` on Telegram.
- Send `/newbot` and follow the instructions.
- Save the **Token** (e.g., `123456789:ABCDefghIJKLmnoPQRstuvwxYZ`).

## 2. Get your Chat ID

- Send a message to your new bot.
- Visit `https://api.telegram.org/bot<YOUR_TOKEN>/getUpdates`.
- Look for `"chat":{"id":123456789}` in the JSON response.

## 3. Configure the Tool

Add the token and chat ID to your `hosts.yaml` or `hosts.toml`:

```yaml
alert:
  telegram:
    enabled: true
    mode: error # 'always' or 'error'
    token: "123456789:ABCDefghIJKLmnoPQRstuvwxYZ"
    chat-id: "123456789"
```

## Notification Modes

- `always`: Sends a report every time the tool is run.
- `error`: Only sends a report if at least one check fails (e.g., server down, disk full, SSL expiring).
