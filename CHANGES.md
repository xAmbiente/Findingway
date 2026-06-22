# Changes from original

## SQLite persistence (replaces unused Postgres migrations)

The old `migrations/` folder contained Postgres-style SQL that was never wired up — 
channel state like `MessageID` and `Enabled` was lost on every restart.

Now the bot uses **SQLite** via `modernc.org/sqlite` (pure Go, zero dependencies, no CGO needed):

- `internal/store/store.go` — thin wrapper around a single `channels` table
- Channel `MessageID` and `Enabled` state persist across restarts automatically
- DB file defaults to `findingway.db` (configurable via `DB_PATH` env var)
- Docker volume at `/findingway/data/` for the DB file

## Environment variables

| Variable        | Default          | Description                    |
|-----------------|------------------|--------------------------------|
| `DISCORD_TOKEN` | *(required)*     | Your Discord bot token         |
| `CONFIG_PATH`   | `config.yaml`    | Path to channel config         |
| `DB_PATH`       | `findingway.db`  | Path to SQLite database file   |

The token is no longer hardcoded in `bat.bat` — set `DISCORD_TOKEN` in your environment.

## Other improvements

- Disabled channels are skipped in the scraping loop (were previously posted anyway)
- `*reload` also re-syncs the channel list on the Discord struct
- `*forcepost` gracefully handles no cached listings
- `*interval` enforces a 30-second minimum to avoid API spam
- `*channels` output shows duty name alongside enabled state
- Embed building extracted into `buildEmbed()` — no more duplicated code
- All log lines use `[LOG]`/`[ERROR]`/`[WARN]` consistently

## After upgrading

```bash
go mod tidy   # downloads modernc.org/sqlite and its deps
go build .
```
