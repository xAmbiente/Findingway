# Findingway

A Discord bot that scrapes [XIVPF](https://xivpf.com) for Final Fantasy XIV Party Finder listings and posts them into dedicated Discord channels — one embed per channel, updated in place every 3 minutes.

---

## Features

- Scrapes live PF listings from xivpf.com
- Filters listings by duty and data centre
- Posts a single persistent embed per Discord channel, edited on every cycle (no spam)
- Supports multiple channels and duties simultaneously
- Automatically cleans up stale bot messages on startup
- Recovers gracefully if a message is deleted externally

---

## Supported Duties

| Channel | Duty |
|---------|------|
| UWU | The Weapon's Refrain (Ultimate) |
| UCoB | The Unending Coil of Bahamut (Ultimate) |
| TEA | The Epic of Alexander (Ultimate) |
| FRU | Futures Rewritten (Ultimate) |
| TOP | The Omega Protocol (Ultimate) |
| DSR | Dragonsong's Reprise (Ultimate) |

---

## Requirements

- Go 1.21+
- A Discord bot token with the following permissions:
  - `Send Messages`
  - `Read Message History`
  - `Manage Messages` (for bulk delete on startup)
  - `Embed Links`

---

## Setup

### 1. Clone the repository

```bash
git clone https://github.com/Veraticus/findingway.git
cd findingway
```

### 2. Install dependencies

```bash
go mod tidy
```

### 3. Configure channels

Edit `config.yaml` in the project root:

```yaml
channels:
  - id: "YOUR_CHANNEL_ID"
    name: "uwu"
    duty: "The Weapon's Refrain (Ultimate)"
    dataCentres: [Light]
  - id: "YOUR_CHANNEL_ID"
    name: "ucob"
    duty: "The Unending Coil of Bahamut (Ultimate)"
    dataCentres: [Light]
```

To get a channel ID in Discord: enable Developer Mode under `Settings → Advanced`, then right-click any channel and select **Copy Channel ID**.

### 4. Set your bot token

In `main.go`, replace the `discordToken` value with your own bot token from the [Discord Developer Portal](https://discord.com/developers/applications).

### 5. Build and run

```bash
go build -o findingway ./cmd/findingway
./findingway
```

---

## How It Works

1. On startup, the bot connects to Discord and scans each configured channel.
2. All existing bot messages in each channel are deleted.
3. A single placeholder embed is posted per channel and its message ID is stored in memory.
4. Every 3 minutes, the bot scrapes xivpf.com and edits that same embed with fresh listings.
5. If a listing was updated within the last 4 minutes, only recently-updated listings are shown to keep the embed relevant.

---

## Project Structure

```
findingway/
├── cmd/
│   └── findingway/
│       └── main.go          # Entry point
├── internal/
│   ├── discord/
│   │   └── discord.go       # Discord session, embed posting, message management
│   ├── ffxiv/
│   │   └── listings.go      # Listing types and filtering logic
│   └── scraper/
│       └── scraper.go       # xivpf.com scraper
├── config.yaml              # Channel configuration
├── go.mod
└── go.sum
```

---

## Configuration Reference

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Discord channel ID |
| `name` | string | Human-readable label (used in logs) |
| `duty` | string | Exact duty name as it appears on xivpf.com |
| `dataCentres` | list | One or more data centres to filter by (e.g. `[Light]`) |

---

## License

MIT