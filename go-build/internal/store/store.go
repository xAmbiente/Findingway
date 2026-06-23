// Package store persists channel state (embed message IDs, enabled flags)
// in a local SQLite database using modernc.org/sqlite (pure Go, no CGO).
package store

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

// Store is a thin wrapper around a single SQLite file.
type Store struct {
	db *sql.DB
}

// ChannelState is the persisted row for one Discord channel.
type ChannelState struct {
	ChannelID string
	MessageID string
	Enabled   bool
}

// New opens (or creates) the SQLite database at path and runs migrations.
func New(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite at %q: %w", path, err)
	}

	// SQLite supports only one writer at a time.
	db.SetMaxOpenConns(1)

	if err := migrate(db); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return &Store{db: db}, nil
}

// Close releases the database handle.
func (s *Store) Close() error {
	return s.db.Close()
}

func migrate(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS channels (
			channel_id TEXT    PRIMARY KEY,
			message_id TEXT    NOT NULL DEFAULT '',
			enabled    INTEGER NOT NULL DEFAULT 1
		);
	`)
	return err
}

// GetChannel returns the persisted state for channelID.
// If no row exists yet a default (enabled=true, messageID="") is returned.
func (s *Store) GetChannel(channelID string) (*ChannelState, error) {
	var cs ChannelState
	var enabled int
	err := s.db.QueryRow(
		`SELECT channel_id, message_id, enabled FROM channels WHERE channel_id = ?`,
		channelID,
	).Scan(&cs.ChannelID, &cs.MessageID, &enabled)

	if err == sql.ErrNoRows {
		return &ChannelState{ChannelID: channelID, Enabled: true}, nil
	}
	if err != nil {
		return nil, err
	}
	cs.Enabled = enabled != 0
	return &cs, nil
}

// UpdateMessageID sets the embed message ID for a channel without changing its
// enabled state (inserts a new row with enabled=true if one doesn't exist yet).
func (s *Store) UpdateMessageID(channelID, messageID string) error {
	_, err := s.db.Exec(`
		INSERT INTO channels (channel_id, message_id, enabled) VALUES (?, ?, 1)
		ON CONFLICT(channel_id) DO UPDATE SET message_id = excluded.message_id`,
		channelID, messageID,
	)
	return err
}

// ClearMessageID resets the embed message ID so a fresh embed is posted next cycle.
func (s *Store) ClearMessageID(channelID string) error {
	return s.UpdateMessageID(channelID, "")
}

// SetEnabled toggles a channel on or off without touching its message ID.
func (s *Store) SetEnabled(channelID string, enabled bool) error {
	v := 0
	if enabled {
		v = 1
	}
	_, err := s.db.Exec(`
		INSERT INTO channels (channel_id, message_id, enabled) VALUES (?, '', ?)
		ON CONFLICT(channel_id) DO UPDATE SET enabled = excluded.enabled`,
		channelID, v,
	)
	return err
}
