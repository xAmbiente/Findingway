# Findingway

[![Continuous Deployment](https://github.com/xAmbiente/Findingway/actions/workflows/continuous-deployment.yml/badge.svg)](https://github.com/xAmbiente/Findingway/actions/workflows/continuous-deployment.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A Discord bot that monitors [XIVPF](https://xivpf.com) for Final Fantasy XIV Party Finder listings and posts them into dedicated Discord channels. Each channel displays a single auto-updating embed that refreshes every 3 minutes, eliminating spam and keeping listings current.

## Overview

Findingway streamlines the party-finding experience for FFXIV players by aggregating Ultimate raid listings directly in Discord. Instead of constantly checking the PF website, players see live updates in their guild Discord, filtered by duty and data center.

## Features

- **Live PF Scraping**: Pulls current Party Finder listings from xivpf.com
- **Smart Filtering**: Filter by duty and data center per channel
- **No Spam**: Single persistent embed per channel, edited in-place (no message floods)
- **Multi-Channel Support**: Monitor multiple duties across multiple channels simultaneously
- **Auto-Cleanup**: Removes stale bot messages on startup
- **Resilience**: Gracefully recovers if messages are deleted externally
- **Efficient Updates**: Only shows recently-updated listings to keep embeds relevant

## Supported Duties

| Duty                                    | Abbreviation |
| --------------------------------------- | ------------ |
| The Weapon's Refrain (Ultimate)         | UWU          |
| The Unending Coil of Bahamut (Ultimate) | UCoB         |
| The Epic of Alexander (Ultimate)        | TEA          |
| Futures Rewritten (Ultimate)            | FRU          |
| The Omega Protocol (Ultimate)           | TOP          |
| Dragonsong's Reprise (Ultimate)         | DSR          |

## Prerequisites

- **Go** 1.22 or later
- **Discord Bot Token** from the [Discord Developer Portal](https://discord.com/developers/applications)
- **Bot Permissions** (minimum required):
  - Send Messages
  - Read Message History
  - Manage Messages
  - Embed Links

## Quick Start

### 1. Clone the repository

```bash
git clone https://github.com/xambiente/findingway.git
cd findingway
```

### 2. Install dependencies

```bash
go mod tidy
```

### 3. Configure channels

Edit `config.yaml` in the project root to define which Discord channels monitor which duties:

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

**To get a channel ID**: Enable Developer Mode in Discord (`Settings → Advanced → Developer Mode`), then right-click any channel and select **Copy Channel ID**.

### 4. Set environment variables

Create a `.env` file or export the Discord bot token:

```bash
export DISCORD_TOKEN="your-bot-token-here"
```

Alternatively, pass it as a command-line argument (see `./findingway --help`).

### 5. Build and run

```bash
go build -o findingway .
./findingway
```

## How It Works

The bot operates on a simple update cycle:

1. **Startup**: Connects to Discord and scans configured channels
2. **Cleanup**: Removes all existing bot messages from each channel
3. **Initialize**: Posts a placeholder embed per channel and stores the message ID
4. **Update Loop** (every 3 minutes):
   - Scrapes xivpf.com for current Party Finder listings
   - Filters by configured duty and data center
   - Edits the existing embed with fresh results
   - Includes only recently-updated listings (last 4 minutes) to keep results fresh

This approach ensures a clean, spam-free channel while maintaining up-to-date information.

## Development

### Build for local development

```bash
make build
```

### Run tests

```bash
make test
```

### Build for specific platforms

```bash
# Linux (Docker-compatible)
make build-linux

# macOS ARM64
make build-macos
```

## Deployment

### Docker

Build and run the container:

```bash
make docker
make docker-run
```

The included `docker-compose.yml` automatically configures the environment. Ensure `config.yaml` and your Discord token are available to the container.

### Fly.io

The project includes `fly.toml` for deployment to [Fly.io](https://fly.io):

```bash
fly deploy
```

## Project Structure

```
findingway/
├── main.go                  # Bot entry point and main event loop
├── internal/
│   ├── discord/             # Discord API integration
│   │   ├── discord.go       # Session, embed posting, message management
│   │   └── discord_test.go
│   ├── ffxiv/               # FFXIV-specific models and logic
│   │   ├── ffxiv.go         # Core FFXIV types and utilities
│   │   ├── job.go           # Job class definitions
│   │   ├── role.go          # Role (Tank, Healer, DPS) definitions
│   │   └── ffxiv_test.go
│   ├── scraper/             # Web scraping logic
│   │   ├── scraper.go       # xivpf.com scraper implementation
│   │   └── scraper_test.go
│   ├── store/               # Data persistence
│   │   └── store.go         # SQLite database operations
│   └── logger/              # Logging utilities
│       └── logger.go
├── assets/
│   ├── emojis.json          # Discord emoji definitions
│   └── emotes/              # Custom emote assets
├── config.yaml              # Channel configuration
├── docker-compose.yml       # Docker Compose configuration
├── Dockerfile               # Container image definition
├── go.mod                   # Go module dependencies
├── Makefile                 # Build and development targets
└── fly.toml                 # Fly.io deployment configuration
```

## Configuration Reference

The `config.yaml` file defines which Discord channels monitor which duties:

```yaml
channels:
  - id: "CHANNEL_ID_HERE"
    name: "Short Label"
    duty: "Exact Duty Name"
    dataCentres:
      - DataCenterName
```

| Field         | Type   | Description                                     | Example                             |
| ------------- | ------ | ----------------------------------------------- | ----------------------------------- |
| `id`          | string | Discord channel ID (obtain with Developer Mode) | `"1234567890123456789"`             |
| `name`        | string | Human-readable label for logs and reference     | `"uwu"`                             |
| `duty`        | string | Exact duty name as it appears on xivpf.com      | `"The Weapon's Refrain (Ultimate)"` |
| `dataCentres` | list   | Filter by one or more data centres              | `[Light]`, `[Light, Chaos]`         |

## Troubleshooting

**Bot doesn't connect to Discord**

- Verify your bot token is valid and set in `DISCORD_TOKEN` environment variable
- Ensure your bot has been invited to the server with proper permissions

**Embeds not updating**

- Check that channel IDs in `config.yaml` are correct
- Verify the bot has "Manage Messages" permission in the channels
- Review logs for scraping errors (network issues, website changes)

**Messages getting deleted**

- This is handled gracefully — the bot will post a new message on the next update cycle
- Check your Discord audit log if deletions are unexpected

## Related Resources

- [XIVPF](https://xivpf.com) - Final Fantasy XIV Party Finder tool
- [Discord.js Documentation](https://discord.js.org/) - Discord API reference
- [Colly Web Scraper](http://go-colly.org/) - Go web scraping framework used by this project
