package discord

import (
	"fmt"
	"strings"
	"time"

	"github.com/Veraticus/findingway/internal/ffxiv"
	"github.com/bwmarrin/discordgo"
)

type Discord struct {
	Token string

	Session  *discordgo.Session
	Channels []*Channel `yaml:"channels"`
}

type Channel struct {
	Name        string   `yaml:"name"`
	ID          string   `yaml:"id"`
	Duty        string   `yaml:"duty"`
	DataCentres []string `yaml:"dataCentres"`
	MessageID   string   `yaml:"messageID"`
}

func (d *Discord) Start() error {
	s, err := discordgo.New("Bot " + d.Token)
	if err != nil {
		return fmt.Errorf("Could not start Discord: %v", err)
	}
	s.ShouldRetryOnRateLimit = false

	err = s.Open()
	if err != nil {
		return fmt.Errorf("Could not open Discord session: %v", err)
	}

	d.Session = s

	// For each channel: delete ALL existing bot messages, then send one
	// fresh placeholder so we always have exactly one message to edit.
	for _, c := range d.Channels {
		if err := d.resetChannelMessage(c); err != nil {
			fmt.Printf("Warning: could not reset channel %s: %v\n", c.ID, err)
		}
	}

	return nil
}

// resetChannelMessage deletes every bot message in the channel and sends a
// single placeholder embed. The new message ID is stored on the Channel so
// every subsequent cycle just edits it.
func (d *Discord) resetChannelMessage(channel *Channel) error {
	botUser, err := d.Session.User("@me")
	if err != nil {
		return fmt.Errorf("could not get bot user: %w", err)
	}

	messages, err := d.Session.ChannelMessages(channel.ID, 100, "", "", "")
	if err != nil {
		return fmt.Errorf("could not list messages: %w", err)
	}

	var botMsgIDs []string
	for _, m := range messages {
		if m.Author.ID == botUser.ID {
			botMsgIDs = append(botMsgIDs, m.ID)
		}
	}

	if len(botMsgIDs) >= 2 {
		if err := d.Session.ChannelMessagesBulkDelete(channel.ID, botMsgIDs); err != nil {
			fmt.Printf("Bulk delete failed, falling back to individual deletes: %v\n", err)
			for _, id := range botMsgIDs {
				_ = d.Session.ChannelMessageDelete(channel.ID, id)
			}
		}
	} else if len(botMsgIDs) == 1 {
		if err := d.Session.ChannelMessageDelete(channel.ID, botMsgIDs[0]); err != nil {
			fmt.Printf("Could not delete single bot message: %v\n", err)
		}
	}

	placeholder := &discordgo.MessageEmbed{
		Title:       "Loading listings...",
		Description: "Please wait.",
		Color:       0x6600ff,
	}
	msg, err := d.Session.ChannelMessageSendComplex(channel.ID, &discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{placeholder},
	})
	if err != nil {
		return fmt.Errorf("could not send placeholder message: %w", err)
	}

	channel.MessageID = msg.ID
	fmt.Printf("Channel %s initialised with message ID %s\n", channel.ID, msg.ID)
	return nil
}

func (d *Discord) PostListings(channelId string, listings *ffxiv.Listings, duty string, dataCentre string) error {
	scopedListings := listings.ForDutyAndDataCentre(duty, dataCentre)

	mostRecent, err := scopedListings.MostRecentUpdated()
	if err != nil {
		return fmt.Errorf("Could not find most recently updated duty: %w", err)
	}
	if mostRecent != nil {
		mostRecentUpdated, err := mostRecent.UpdatedAt()
		if err != nil {
			return fmt.Errorf("Could not find most recently updatedAt: %w", err)
		}
		if mostRecentUpdated.After(time.Now().Add(-4 * time.Minute)) {
			scopedListings, err = scopedListings.UpdatedWithinLast(4 * time.Minute)
			if err != nil {
				return fmt.Errorf("Could not find most recent listings: %w", err)
			}
		}
	}

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("%s PFs (%v)", duty, dataCentre),
		Description: fmt.Sprintf("Found %v listings %v", len(scopedListings.Listings), fmt.Sprintf("<t:%v:R>", time.Now().Unix())),
		Color:       0x6600ff,
		Footer: &discordgo.MessageEmbedFooter{
			Text: strings.Repeat("\u3000", 20),
		},
	}

	for _, listing := range scopedListings.Listings {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   listing.Creator,
			Value:  listing.PartyDisplay(),
			Inline: true,
		})
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   listing.GetTags(),
			Value:  listing.GetDescription(),
			Inline: true,
		})
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   listing.GetExpires(),
			Value:  listing.GetUpdated(),
			Inline: true,
		})
	}

	return d.editMessage(channelId, embed)
}

// editMessage edits the channel's one persistent message. If the message was
// somehow deleted externally, it re-runs the full reset so we get back to a
// clean single-message state.
func (d *Discord) editMessage(channelId string, embed *discordgo.MessageEmbed) error {
	var channel *Channel
	for _, c := range d.Channels {
		if c.ID == channelId {
			channel = c
			break
		}
	}
	if channel == nil {
		return fmt.Errorf("channel %s not found in configuration", channelId)
	}

	if channel.MessageID == "" {
		fmt.Println("No message ID set, re-initialising channel...")
		if err := d.resetChannelMessage(channel); err != nil {
			return fmt.Errorf("could not re-initialise channel: %w", err)
		}
	}

	_, err := d.Session.ChannelMessageEditEmbed(channelId, channel.MessageID, embed)
	if err != nil {
		fmt.Printf("Edit failed (message deleted?), re-initialising: %v\n", err)
		if resetErr := d.resetChannelMessage(channel); resetErr != nil {
			return fmt.Errorf("could not re-initialise channel after edit failure: %w", resetErr)
		}
		_, err = d.Session.ChannelMessageEditEmbed(channelId, channel.MessageID, embed)
		if err != nil {
			return fmt.Errorf("could not edit message after re-initialise: %w", err)
		}
	}

	fmt.Printf("Edited message %s in channel %s\n", channel.MessageID, channelId)
	return nil
}