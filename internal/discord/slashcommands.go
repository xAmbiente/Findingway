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

// BotInterface keeps dependency clean (your Bot struct implements this)
type BotInterface interface {
	LoadConfig() error
	GetWaitTime() time.Duration
	SetWaitTime(d time.Duration)
	GetChannels() []*Channel
	GetChannelByName(name string) *Channel
	EnableChannel(name string)
	DisableChannel(name string)
	ForceScrape() error
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
			Description: "Enable channel",
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
			Description: "Disable channel",
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

	s.respond(i, fmt.Sprintf("📊 Channels: %d total / %d enabled | Interval: %s",
		len(ch), enabled, s.Bot.GetWaitTime()))
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
	s.respond(i, fmt.Sprintf("🟢 Enabled %s", strings.ToUpper(name)))
}

func (s *SlashCommandManager) handleDisable(i *discordgo.InteractionCreate) {
	name := i.ApplicationCommandData().Options[0].StringValue()
	s.Bot.DisableChannel(name)
	s.respond(i, fmt.Sprintf("🔴 Disabled %s", strings.ToUpper(name)))
}