package store

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

// Store persists channel state (message IDs, enabled flags) across restarts.
type Store struct {
	db *sql.DB
}

type ChannelState struct {
	ChannelID string
	MessageID string
	Enabled   bool
}

func New(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	db.SetMaxOpenConns(1) // SQLite is single-writer

	if err := migrate(db); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func migrate(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS channels (
			channel_id TEXT PRIMARY KEY,
			message_id TEXT NOT NULL DEFAULT '',
			enabled    INTEGER NOT NULL DEFAULT 1
		);
	`)
	return err
}

// GetChannel returns stored state for a channel, or a default if not found.
func (s *Store) GetChannel(channelID string) (*ChannelState, error) {
	row := s.db.QueryRow(
		`SELECT channel_id, message_id, enabled FROM channels WHERE channel_id = ?`,
		channelID,
	)
	var cs ChannelState
	var enabled int
	err := row.Scan(&cs.ChannelID, &cs.MessageID, &enabled)
	if err == sql.ErrNoRows {
		return &ChannelState{ChannelID: channelID, MessageID: "", Enabled: true}, nil
	}
	if err != nil {
		return nil, err
	}
	cs.Enabled = enabled != 0
	return &cs, nil
}

// SaveChannel upserts channel state.
func (s *Store) SaveChannel(cs *ChannelState) error {
	enabled := 0
	if cs.Enabled {
		enabled = 1
	}
	_, err := s.db.Exec(
		`INSERT INTO channels (channel_id, message_id, enabled)
		 VALUES (?, ?, ?)
		 ON CONFLICT(channel_id) DO UPDATE SET
		   message_id = excluded.message_id,
		   enabled    = excluded.enabled`,
		cs.ChannelID, cs.MessageID, enabled,
	)
	return err
}

// SetMessageID updates just the message ID for a channel.
func (s *Store) SetMessageID(channelID, messageID string) error {
	return s.SaveChannel(&ChannelState{
		ChannelID: channelID,
		MessageID: messageID,
		Enabled:   true, // will be overwritten by upsert only if row doesn't exist
	})
}

// UpdateMessageID updates message_id without touching enabled.
func (s *Store) UpdateMessageID(channelID, messageID string) error {
	_, err := s.db.Exec(
		`INSERT INTO channels (channel_id, message_id, enabled)
		 VALUES (?, ?, 1)
		 ON CONFLICT(channel_id) DO UPDATE SET message_id = excluded.message_id`,
		channelID, messageID,
	)
	return err
}

// SetEnabled updates enabled flag without touching message_id.
func (s *Store) SetEnabled(channelID string, enabled bool) error {
	v := 0
	if enabled {
		v = 1
	}
	_, err := s.db.Exec(
		`INSERT INTO channels (channel_id, message_id, enabled)
		 VALUES (?, '', ?)
		 ON CONFLICT(channel_id) DO UPDATE SET enabled = excluded.enabled`,
		channelID, v,
	)
	return err
}

// ClearMessageID resets the stored message ID so a fresh embed gets posted.
func (s *Store) ClearMessageID(channelID string) error {
	return s.UpdateMessageID(channelID, "")
}
