package discord

import (
	"fmt"
	"strings"
	"time"

	"github.com/Veraticus/findingway/internal/ffxiv"
	"github.com/Veraticus/findingway/internal/logger"
	"github.com/Veraticus/findingway/internal/store"
	"github.com/bwmarrin/discordgo"
)

// Discord owns the bot session and all channel state.
type Discord struct {
	Token    string
	Channels []*Channel
	Session  *discordgo.Session
	Store    *store.Store
}

// Close gracefully shuts down the internal discord session if open.
func (d *Discord) Close() error {
	if d == nil || d.Session == nil {
		return nil
	}
	return d.Session.Close()
}

// Channel mirrors one entry in config.yaml, augmented with runtime state.
type Channel struct {
	Name        string   `yaml:"name"`
	ID          string   `yaml:"id"`
	Duty        string   `yaml:"duty"`
	DataCentres []string `yaml:"dataCentres"`
	MessageID   string   `yaml:"messageID"`
	Enabled     bool     `yaml:"enabled"`
}

// Start opens the Discord WebSocket session and hydrates channel state from the DB.
func (d *Discord) Start() error {
	s, err := discordgo.New("Bot " + d.Token)
	if err != nil {
		return fmt.Errorf("create session: %w", err)
	}
	s.ShouldRetryOnRateLimit = false

	if err := s.Open(); err != nil {
		return fmt.Errorf("open websocket: %w", err)
	}
	d.Session = s

	// Hydrate each channel from the DB (DB wins over YAML).
	for _, c := range d.Channels {
		cs, err := d.Store.GetChannel(c.ID)
		if err != nil {
			logger.Warn("could not load DB state for channel %s (%s): %v", c.Name, c.ID, err)
			continue
		}
		c.Enabled = cs.Enabled
		c.MessageID = cs.MessageID
	}

	// Post a "Loading…" embed for every enabled channel that has no embed yet.
	for _, c := range d.Channels {
		if !c.Enabled || c.MessageID != "" {
			continue
		}
		if err := d.resetChannelMessage(c); err != nil {
			logger.Warn("initial reset failed for %s (%s): %v", c.Name, c.ID, err)
		}
	}

	return nil
}

// resetChannelMessage deletes any previous bot messages in the channel, then
// posts a fresh placeholder embed and stores its ID.
func (d *Discord) resetChannelMessage(ch *Channel) error {
	botUser, err := d.Session.User("@me")
	if err != nil {
		return fmt.Errorf("get self: %w", err)
	}

	msgs, err := d.Session.ChannelMessages(ch.ID, 100, "", "", "")
	if err != nil {
		return fmt.Errorf("list messages: %w", err)
	}

	var toDelete []string
	for _, m := range msgs {
		if m.Author.ID == botUser.ID {
			toDelete = append(toDelete, m.ID)
		}
	}
	switch len(toDelete) {
	case 0:
		// nothing to clean up
	case 1:
		_ = d.Session.ChannelMessageDelete(ch.ID, toDelete[0])
	default:
		_ = d.Session.ChannelMessagesBulkDelete(ch.ID, toDelete)
	}

	msg, err := d.Session.ChannelMessageSendEmbed(ch.ID, &discordgo.MessageEmbed{
		Title:       "Loading listings…",
		Description: "First update coming shortly.",
		Color:       0x6600ff,
	})
	if err != nil {
		return fmt.Errorf("send placeholder: %w", err)
	}

	ch.MessageID = msg.ID
	if err := d.Store.UpdateMessageID(ch.ID, msg.ID); err != nil {
		logger.Warn("could not persist message ID for %s: %v", ch.Name, err)
	}
	return nil
}

// UpdateEmbedMessage edits the pinned embed for channelID / dc.
func (d *Discord) UpdateEmbedMessage(channelID, _ string, listings *ffxiv.Listings, duty, dc string) error {
	ch := d.getChannel(channelID)
	if ch == nil {
		return fmt.Errorf("unknown channel: %s", channelID)
	}
	return d.editMessage(ch, d.buildEmbed(listings, duty, dc))
}

// PostEmbedMessage is an alias for UpdateEmbedMessage – the edit path handles
// missing message IDs by calling resetChannelMessage automatically.
func (d *Discord) PostEmbedMessage(channelID string, listings *ffxiv.Listings, duty, dc string) error {
	return d.UpdateEmbedMessage(channelID, "", listings, duty, dc)
}

// PostListings is the public entry-point used by the scraping loop.
func (d *Discord) PostListings(channelID string, listings *ffxiv.Listings, duty, dc string) error {
	return d.PostEmbedMessage(channelID, listings, duty, dc)
}

// GetLastMessageID returns the stored embed message ID for a channel.
func (d *Discord) GetLastMessageID(channelID, _ string) string {
	if ch := d.getChannel(channelID); ch != nil {
		return ch.MessageID
	}
	return ""
}

// ResetChannelMessage clears the stored message ID so a fresh embed is posted.
func (d *Discord) ResetChannelMessage(ch *Channel) error {
	ch.MessageID = ""
	return d.Store.ClearMessageID(ch.ID)
}

// EnableChannel turns a channel on and persists the change.
func (d *Discord) EnableChannel(name string) {
	if ch := d.GetChannelByName(name); ch != nil {
		ch.Enabled = true
		if err := d.Store.SetEnabled(ch.ID, true); err != nil {
			logger.Warn("could not persist enable for %s: %v", name, err)
		}
	}
}

// DisableChannel turns a channel off and persists the change.
func (d *Discord) DisableChannel(name string) {
	if ch := d.GetChannelByName(name); ch != nil {
		ch.Enabled = false
		if err := d.Store.SetEnabled(ch.ID, false); err != nil {
			logger.Warn("could not persist disable for %s: %v", name, err)
		}
	}
}

func (d *Discord) GetChannelByName(name string) *Channel {
	for _, c := range d.Channels {
		if c.Name == name {
			return c
		}
	}
	return nil
}

// ── internal helpers ─────────────────────────────────────────────────────────

func (d *Discord) getChannel(id string) *Channel {
	for _, c := range d.Channels {
		if c.ID == id {
			return c
		}
	}
	return nil
}

// buildEmbed constructs the Discord embed for a duty/dc combination.
func (d *Discord) buildEmbed(listings *ffxiv.Listings, duty, dc string) *discordgo.MessageEmbed {
	scoped := listings.ForDutyAndDataCentre(duty, dc)

	embed := &discordgo.MessageEmbed{
		Title: fmt.Sprintf("%s PFs (%s)", duty, dc),
		Description: fmt.Sprintf(
			"Found **%d** listings • <t:%d:R>",
			len(scoped.Listings),
			time.Now().Unix(),
		),
		Color: 0x6600ff,
		Footer: &discordgo.MessageEmbedFooter{
			Text: strings.Repeat("\u3000", 20),
		},
	}

	fieldCount := 0
	for _, l := range scoped.Listings {
		if fieldCount >= 24 {
			break
		}
		embed.Fields = append(embed.Fields,
			&discordgo.MessageEmbedField{Name: l.Creator, Value: l.PartyDisplay(), Inline: true},
			&discordgo.MessageEmbedField{Name: l.GetTags(), Value: l.GetDescription(), Inline: true},
			&discordgo.MessageEmbedField{Name: l.GetExpires(), Value: l.GetUpdated(), Inline: true},
		)
		fieldCount += 3
	}

	return embed
}

// editMessage edits the channel's pinned embed, recreating it if it was deleted.
func (d *Discord) editMessage(ch *Channel, embed *discordgo.MessageEmbed) error {
	if ch.MessageID == "" {
		if err := d.resetChannelMessage(ch); err != nil {
			return err
		}
	}

	_, err := d.Session.ChannelMessageEditEmbed(ch.ID, ch.MessageID, embed)
	if err != nil {
		// Embed was deleted externally — recreate and try once more.
		if resetErr := d.resetChannelMessage(ch); resetErr != nil {
			return resetErr
		}
		_, err = d.Session.ChannelMessageEditEmbed(ch.ID, ch.MessageID, embed)
	}
	return err
}
