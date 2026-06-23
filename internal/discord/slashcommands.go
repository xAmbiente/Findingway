package discord

import (
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

// ─────────────────────────────────────────────────────────────
// SlashCommandManager
// ─────────────────────────────────────────────────────────────

type SlashCommandManager struct {
	Session *discordgo.Session
	Bot     BotInterface
}

type BotInterface interface {
	LoadConfig() error
	GetWaitTime() time.Duration
	SetWaitTime(d time.Duration)
	GetChannels() []*Channel
	GetChannelByName(name string) *Channel
	EnableChannel(name string)
	DisableChannel(name string)
	ForceScrape() error
	// Announcements
	GetAnnouncementsChannel() string
	SetAnnouncementsChannel(id string)
	// Listing info
	GetCachedListingCount() int
	ClearMercCache()
}

// ─────────────────────────────────────────────────────────────
// Register Slash Commands
// ─────────────────────────────────────────────────────────────

func (s *SlashCommandManager) RegisterCommands(guildID string) error {

	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "status",
			Description: "Show bot status",
		},
		{
			Name:        "reload",
			Description: "Reload config",
		},
		{
			Name:        "interval",
			Description: "Set scrape interval",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "duration",
					Description: "e.g. 2m, 30s",
					Required:    true,
				},
			},
		},
		{
			Name:        "scrape",
			Description: "Trigger immediate scrape",
		},
		{
			Name:        "enable",
			Description: "Enable a monitored channel",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "name",
					Description: "Channel name (as configured)",
					Required:    true,
				},
			},
		},
		{
			Name:        "disable",
			Description: "Disable a monitored channel",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "name",
					Description: "Channel name (as configured)",
					Required:    true,
				},
			},
		},
		{
			Name:        "channels",
			Description: "List all configured channels and their status",
		},
		{
			Name:        "announce",
			Description: "Set or clear the announcements channel for merc/gill alerts",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "channel_id",
					Description: "Channel ID, or 'off' to disable",
					Required:    false,
				},
			},
		},
		{
			Name:        "listingcount",
			Description: "Show active listing counts per channel (excludes expired)",
		},
		{
			Name:        "clearold",
			Description: "Clear the merc-announcement seen-listing cache",
		},
		{
			Name:        "resetmsg",
			Description: "Clear embed message so a fresh one is posted next cycle",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "name",
					Description: "Channel name (as configured)",
					Required:    true,
				},
			},
		},
	}

	// Resolve application (bot) ID reliably
	app, err := s.Session.User("@me")
	if err != nil {
		return fmt.Errorf("could not resolve bot user: %w", err)
	}

	for _, cmd := range commands {
		_, err := s.Session.ApplicationCommandCreate(app.ID, guildID, cmd)
		if err != nil {
			return fmt.Errorf("failed to create command %s: %w", cmd.Name, err)
		}
	}

	return nil
}

// ─────────────────────────────────────────────────────────────
// Interaction Handler
// ─────────────────────────────────────────────────────────────

func (s *SlashCommandManager) HandleInteraction(i *discordgo.InteractionCreate) {

	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}

	data := i.ApplicationCommandData()

	switch data.Name {

	case "status":
		s.handleStatus(i)

	case "reload":
		s.handleReload(i)

	case "interval":
		s.handleInterval(i)

	case "scrape":
		s.handleScrape(i)

	case "enable":
		s.handleEnable(i)

	case "disable":
		s.handleDisable(i)

	case "channels":
		s.handleChannels(i)

	case "announce":
		s.handleAnnounce(i)

	case "listingcount":
		s.handleListingCount(i)

	case "clearold":
		s.handleClearOld(i)

	case "resetmsg":
		s.handleResetMsg(i)
	}
}

// ─────────────────────────────────────────────────────────────
// Handlers
// ─────────────────────────────────────────────────────────────

func (s *SlashCommandManager) respond(i *discordgo.InteractionCreate, msg string) {
	_ = s.Session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: msg,
		},
	})
}

func (s *SlashCommandManager) handleStatus(i *discordgo.InteractionCreate) {
	ch := s.Bot.GetChannels()

	enabled := 0
	for _, c := range ch {
		if c.Enabled {
			enabled++
		}
	}

	ann := s.Bot.GetAnnouncementsChannel()
	if ann == "" {
		ann = "not set"
	} else {
		ann = fmt.Sprintf("<#%s>", ann)
	}

	s.respond(i, fmt.Sprintf(
		"📊 Channels: %d total / %d enabled | Interval: %s | Announcements: %s",
		len(ch), enabled, s.Bot.GetWaitTime(), ann,
	))
}

func (s *SlashCommandManager) handleReload(i *discordgo.InteractionCreate) {
	if err := s.Bot.LoadConfig(); err != nil {
		s.respond(i, "❌ Reload failed")
		return
	}
	s.respond(i, "✅ Config reloaded")
}

func (s *SlashCommandManager) handleInterval(i *discordgo.InteractionCreate) {
	val := i.ApplicationCommandData().Options[0].StringValue()

	d, err := time.ParseDuration(val)
	if err != nil || d < 30*time.Second {
		s.respond(i, "❌ Invalid duration (min 30s)")
		return
	}

	s.Bot.SetWaitTime(d)
	s.respond(i, fmt.Sprintf("✅ Interval set to %s", d))
}

func (s *SlashCommandManager) handleScrape(i *discordgo.InteractionCreate) {
	go s.Bot.ForceScrape()
	s.respond(i, "🚀 Scraping triggered")
}

func (s *SlashCommandManager) handleEnable(i *discordgo.InteractionCreate) {
	name := i.ApplicationCommandData().Options[0].StringValue()
	s.Bot.EnableChannel(name)
	s.respond(i, fmt.Sprintf("🟢 Enabled **%s**", strings.ToUpper(name)))
}

func (s *SlashCommandManager) handleDisable(i *discordgo.InteractionCreate) {
	name := i.ApplicationCommandData().Options[0].StringValue()
	s.Bot.DisableChannel(name)
	s.respond(i, fmt.Sprintf("🔴 Disabled **%s**", strings.ToUpper(name)))
}

func (s *SlashCommandManager) handleChannels(i *discordgo.InteractionCreate) {
	channels := s.Bot.GetChannels()
	if len(channels) == 0 {
		s.respond(i, "No channels configured")
		return
	}
	var lines []string
	for _, c := range channels {
		icon := "🔴"
		if c.Enabled {
			icon = "🟢"
		}
		lines = append(lines, fmt.Sprintf("%s **%s** — %s [DC: %s]", icon, c.Name, c.Duty, strings.Join(c.DataCentres, ", ")))
	}
	s.respond(i, strings.Join(lines, "\n"))
}

func (s *SlashCommandManager) handleAnnounce(i *discordgo.InteractionCreate) {
	opts := i.ApplicationCommandData().Options
	if len(opts) == 0 {
		cur := s.Bot.GetAnnouncementsChannel()
		if cur == "" {
			s.respond(i, "ℹ️ Announcements channel not set. Use `/announce channel_id:<id>` to configure.")
		} else {
			s.respond(i, fmt.Sprintf("ℹ️ Current announcements channel: <#%s>", cur))
		}
		return
	}
	arg := opts[0].StringValue()
	if strings.EqualFold(arg, "off") {
		s.Bot.SetAnnouncementsChannel("")
		s.respond(i, "✅ Merc/payment announcements **disabled**")
		return
	}
	s.Bot.SetAnnouncementsChannel(arg)
	s.respond(i, fmt.Sprintf("✅ Announcements channel set to <#%s>", arg))
}

func (s *SlashCommandManager) handleListingCount(i *discordgo.InteractionCreate) {
	total := s.Bot.GetCachedListingCount()
	if total < 0 {
		s.respond(i, "❌ No cached listings yet — run `/scrape` first")
		return
	}
	s.respond(i, fmt.Sprintf("📋 %d listings currently in cache (expired listings are hidden from embeds)", total))
}

func (s *SlashCommandManager) handleClearOld(i *discordgo.InteractionCreate) {
	s.Bot.ClearMercCache()
	s.respond(i, "✅ Merc announcement cache cleared")
}

func (s *SlashCommandManager) handleResetMsg(i *discordgo.InteractionCreate) {
	name := i.ApplicationCommandData().Options[0].StringValue()
	ch := s.Bot.GetChannelByName(name)
	if ch == nil {
		s.respond(i, "❌ Channel not found")
		return
	}
	ch.MessageID = ""
	s.respond(i, fmt.Sprintf("✅ Embed reset for **%s** — fresh post next cycle", name))
}
