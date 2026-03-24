package discord

import (
	"fmt"
	"strings"
	"time"

	"github.com/Veraticus/findingway/internal/ffxiv"
	"github.com/bwmarrin/discordgo"
)

type Discord struct {
	Token    string
	Channels []*Channel
	Session  *discordgo.Session
}

type Channel struct {
	Name        string   `yaml:"name"`
	ID          string   `yaml:"id"`
	Duty        string   `yaml:"duty"`
	DataCentres []string `yaml:"dataCentres"`
	MessageID   string   `yaml:"messageID"`
	Enabled     bool     `yaml:"enabled"`
}

// Start the Discord session
func (d *Discord) Start() error {
	s, err := discordgo.New("Bot " + d.Token)
	if err != nil {
		return err
	}

	s.ShouldRetryOnRateLimit = false

	if err := s.Open(); err != nil {
		return err
	}

	d.Session = s

	for _, c := range d.Channels {
		if !c.Enabled {
			continue
		}

		if err := d.resetChannelMessage(c); err != nil {
			fmt.Printf("Reset failed for %s: %v\n", c.ID, err)
		}
	}

	return nil
}

// resetChannelMessage deletes old bot messages and sends a fresh "Loading" embed
func (d *Discord) resetChannelMessage(channel *Channel) error {
	botUser, err := d.Session.User("@me")
	if err != nil {
		return err
	}

	messages, err := d.Session.ChannelMessages(channel.ID, 100, "", "", "")
	if err != nil {
		return err
	}

	var botMsgs []string
	for _, m := range messages {
		if m.Author.ID == botUser.ID {
			botMsgs = append(botMsgs, m.ID)
		}
	}

	if len(botMsgs) > 0 {
		if len(botMsgs) > 1 {
			_ = d.Session.ChannelMessagesBulkDelete(channel.ID, botMsgs)
		} else {
			_ = d.Session.ChannelMessageDelete(channel.ID, botMsgs[0])
		}
	}

	msg, err := d.Session.ChannelMessageSendEmbed(channel.ID, &discordgo.MessageEmbed{
		Title:       "Loading listings...",
		Description: "Please wait...",
		Color:       0x6600ff,
	})
	if err != nil {
		return err
	}

	channel.MessageID = msg.ID
	return nil
}

// --- MAIN METHODS USED BY BOT ---

func (d *Discord) GetLastMessageID(channelID, dc string) string {
	ch := d.getChannel(channelID)
	if ch != nil {
		return ch.MessageID
	}
	return ""
}

func (d *Discord) UpdateEmbedMessage(channelID, messageID string, listings *ffxiv.Listings, duty string, dc string) error {
	ch := d.getChannel(channelID)
	if ch == nil {
		return fmt.Errorf("channel not found")
	}

	scoped := listings.ForDutyAndDataCentre(duty, dc)

	embed := &discordgo.MessageEmbed{
		Title: fmt.Sprintf("%s PFs (%s)", duty, dc),
		Description: fmt.Sprintf(
			"Found %d listings • <t:%d:R>",
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
			&discordgo.MessageEmbedField{
				Name:   l.Creator,
				Value:  l.PartyDisplay(),
				Inline: true,
			},
			&discordgo.MessageEmbedField{
				Name:   l.GetTags(),
				Value:  l.GetDescription(),
				Inline: true,
			},
			&discordgo.MessageEmbedField{
				Name:   l.GetExpires(),
				Value:  l.GetUpdated(),
				Inline: true,
			},
		)
		fieldCount += 3
	}

	return d.editMessage(ch, embed)
}

func (d *Discord) PostEmbedMessage(channelID string, listings *ffxiv.Listings, duty string, dc string) error {
	ch := d.getChannel(channelID)
	if ch == nil {
		return fmt.Errorf("channel not found")
	}

	scoped := listings.ForDutyAndDataCentre(duty, dc)

	embed := &discordgo.MessageEmbed{
		Title: fmt.Sprintf("%s PFs (%s)", duty, dc),
		Description: fmt.Sprintf(
			"Found %d listings • <t:%d:R>",
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
			&discordgo.MessageEmbedField{
				Name:   l.Creator,
				Value:  l.PartyDisplay(),
				Inline: true,
			},
			&discordgo.MessageEmbedField{
				Name:   l.GetTags(),
				Value:  l.GetDescription(),
				Inline: true,
			},
			&discordgo.MessageEmbedField{
				Name:   l.GetExpires(),
				Value:  l.GetUpdated(),
				Inline: true,
			},
		)
		fieldCount += 3
	}

	return d.editMessage(ch, embed)
}

// --- HELPERS ---

func (d *Discord) editMessage(channel *Channel, embed *discordgo.MessageEmbed) error {
	if channel.MessageID == "" {
		if err := d.resetChannelMessage(channel); err != nil {
			return err
		}
	}

	_, err := d.Session.ChannelMessageEditEmbed(channel.ID, channel.MessageID, embed)
	if err != nil {
		// message deleted externally → recreate
		if err := d.resetChannelMessage(channel); err != nil {
			return err
		}
		_, err = d.Session.ChannelMessageEditEmbed(channel.ID, channel.MessageID, embed)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *Discord) GetChannelByName(name string) *Channel {
	for _, c := range d.Channels {
		if c.Name == name {
			return c
		}
	}
	return nil
}

func (d *Discord) ResetChannelMessage(ch *Channel) error {
	ch.MessageID = ""
	return nil
}

// PostListings is now fully compatible with ffxiv.Listings
func (d *Discord) PostListings(channelID string, listings *ffxiv.Listings, duty string, dc string) error {
	return d.PostEmbedMessage(channelID, listings, duty, dc)
}

func (d *Discord) getChannel(id string) *Channel {
	for _, c := range d.Channels {
		if c.ID == id {
			return c
		}
	}
	return nil
}

// --- CHANNEL TOGGLE COMMAND HELPERS ---

func (d *Discord) EnableChannel(name string) {
	for _, c := range d.Channels {
		if c.Name == name {
			c.Enabled = true
			return
		}
	}
}

func (d *Discord) DisableChannel(name string) {
	for _, c := range d.Channels {
		if c.Name == name {
			c.Enabled = false
			return
		}
	}
}