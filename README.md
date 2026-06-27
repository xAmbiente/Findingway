# Findingway

[![Continuous Deployment](https://github.com/xAmbiente/Findingway/actions/workflows/continuous-deployment.yml/badge.svg)](https://github.com/xAmbiente/Findingway/actions/workflows/continuous-deployment.yml)
[![License: Apache-2.0](https://img.shields.io/badge/License-Apache--2.0-yellow.svg)](https://opensource.org/license/apache-2.0)

Findingway is a Discord bot that watches [XIVPF](https://xivpf.com) for Final Fantasy XIV Party Finder listings and
posts them into configured Discord channels. It keeps a single message per channel updated in place, so the listings
stay current without flooding chat.

> [!NOTE]
> The bot scrapes XIVPF with Playwright, stores channel configuration in PostgreSQL through Prisma, and uses
> Redis-backed scheduled tasks for periodic updates.

## What It Does

- Scrapes XIVPF listings for the supported ultimate duties and mercantile-style posts
- Posts results as rich Discord components and edits the same message on each refresh
- Runs automatically every 3 minutes, with a manual `/scrape` fallback
- Lets guilds configure which channel receives which duty feed through `/settings`
- Handles deleted or missing messages by recreating them and storing the new message ID

## Supported Feeds

- The Epic of Alexander
- The Unending Coil of Bahamut
- The Weapon's Refrain
- Mercantile listings matched by keywords in the listing description

The mercantile feed currently matches these keywords: `merc`, `mercenary`, `carry`, `gil`, `payment`, `paid`, `boost`,
`sell`, `selling`, `pay`, `fee`, `wage`, and `commission`.

## Requirements

- Node.js 24 or newer
- PostgreSQL
- Redis
- A Discord application with a bot token

> [!IMPORTANT] The bot expects the following environment variables at runtime:
>
> - `DISCORD_TOKEN`
> - `DATABASE_URL`
> - `REDIS_HOST`
> - `REDIS_PORT`
> - `REDIS_PASSWORD`
> - `REDIS_TASK_DB`
> - Optional: `WEBHOOK_ERROR_ID` and `WEBHOOK_ERROR_TOKEN` for error reporting

## Quick Start

1. Clone the repository.

    ```bash
    git clone https://github.com/xAmbiente/Findingway.git
    cd Findingway
    ```

2. Install dependencies.

    ```bash
    yarn install
    ```

3. Provide the required environment variables.

    At minimum, set `DISCORD_TOKEN`, `DATABASE_URL`, and the Redis connection values listed above.

4. Generate the Prisma client.

    ```bash
    yarn prisma generate
    ```

5. Start the bot.

    ```bash
    yarn build
    yarn start
    ```

For local development with rebuild-on-change:

```bash
yarn dev
```

## Commands

- `/info` shows bot statistics, uptime, and server usage. Owners also get a server breakdown button.
- `/scrape` triggers an immediate XIVPF scrape and refresh cycle.
- `/settings set` assigns a duty feed to a channel.
- `/settings toggle` enables or disables an existing duty feed for the guild.
- `/settings show` displays the current channel mappings.

## How It Works

1. A scheduled task runs every 3 minutes.
2. The scraper opens XIVPF in Playwright, filters the listing page, and parses the visible PF entries.
3. Scraped listings are emitted as typed events for each feed.
4. Listeners build a Discord component payload and edit the existing message when possible.
5. Channel IDs and message IDs are persisted in PostgreSQL so updates survive restarts.

## Deployment

### Docker

`docker-compose.yml` starts PostgreSQL, Redis, and the bot container together. The Dockerfile installs Chromium for
Playwright, so the bot can scrape XIVPF inside the container without extra browser setup.

```bash
docker compose up --build
```

### Production Notes

- The bot only needs the `Guilds` intent.
- Channel configuration is stored per guild and per duty type.
- Error handling can forward failures to a Discord webhook when the optional webhook variables are present.

## Project Layout

- `src/commands/` - slash command implementations
- `src/listeners/` - event handlers for posting, errors, and shard lifecycle events
- `src/scheduled-tasks/` - the periodic scrape job
- `src/lib/` - scraper, helpers, setup, generated Prisma client, and utilities
- `prisma/` - Prisma schema and migration history
- `assets/` - emoji and emote assets used by the bot

## Development Scripts

- `yarn build` - compile the TypeScript sources
- `yarn dev` - watch mode with automatic restart after successful builds
- `yarn start` - run the compiled bot
- `yarn lint` - lint and fix source files
- `yarn format` - format the repository with Prettier
- `yarn typecheck` - run the TypeScript project build check
