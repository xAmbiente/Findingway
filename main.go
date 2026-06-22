package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Veraticus/findingway/internal/discord"
	"github.com/Veraticus/findingway/internal/scraper"
	"github.com/Veraticus/findingway/internal/store"
	"github.com/bwmarrin/discordgo"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Channels []*discord.Channel `yaml:"channels"`
}

type Bot struct {
	discordToken string
	configPath   string
	dbPath       string
	discord      *discord.Discord
	scraper      *scraper.Scraper
	store        *store.Store
	cfg          Config
	dg           *discordgo.Session
	waitTime     time.Duration
}

func NewBot(discordToken, configPath, dbPath string) (*Bot, error) {
	fmt.Println("[LOG] Initializing bot")

	dg, err := discordgo.New("Bot " + discordToken)
	if err != nil {
		return nil, fmt.Errorf("create discord session: %w", err)
	}

	db, err := store.New(dbPath)
	if err != nil {
		return nil, fmt.Errorf("open store: %w", err)
	}

	return &Bot{
		discordToken: discordToken,
		configPath:   configPath,
		dbPath:       dbPath,
		scraper:      &scraper.Scraper{Url: "https://xivpf.com"},
		store:        db,
		dg:           dg,
		waitTime:     3 * time.Minute,
	}, nil
}

func (b *Bot) LoadConfig() error {
	fmt.Println("[LOG] Loading config from", b.configPath)
	data, err := os.ReadFile(b.configPath)
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("parse config: %w", err)
	}

	b.cfg = cfg
	fmt.Println("[LOG] Config loaded:", len(cfg.Channels), "channels")
	return nil
}

func (b *Bot) InitializeDiscord() error {
	fmt.Println("[LOG] Initializing Discord")
	d := &discord.Discord{
		Token:    b.discordToken,
		Channels: b.cfg.Channels,
		Store:    b.store,
	}

	if err := d.Start(); err != nil {
		return fmt.Errorf("start discord: %w", err)
	}

	b.discord = d
	fmt.Println("[LOG] Discord session started")
	return nil
}

func (b *Bot) GracefulShutdown() {
	fmt.Println("[LOG] Shutting down")
	if b.discord != nil && b.discord.Session != nil {
		b.discord.Session.Close()
	}
	if b.store != nil {
		b.store.Close()
	}
	fmt.Println("[LOG] Shutdown complete")
}

func (b *Bot) StartScrapingLoop() {
	fmt.Println("[LOG] Starting scraping loop")
	for {
		fmt.Println("[LOG] Scraping...")
		listings, err := b.scraper.Scrape()
		if err != nil {
			fmt.Println("[ERROR] Scraper error:", err)
			time.Sleep(30 * time.Second)
			continue
		}

		b.scraper.LastListings = listings

		if len(listings.Listings) == 0 {
			fmt.Println("[LOG] No listings found, sleeping", b.waitTime)
			time.Sleep(b.waitTime)
			continue
		}

		for _, c := range b.discord.Channels {
			if c == nil || !c.Enabled {
				continue
			}

			for _, dc := range c.DataCentres {
				msgID := b.discord.GetLastMessageID(c.ID, dc)
				var err error
				if msgID != "" {
					fmt.Printf("[LOG] Updating embed for %s (%s)\n", c.Name, dc)
					err = b.discord.UpdateEmbedMessage(c.ID, msgID, listings, c.Duty, dc)
				} else {
					fmt.Printf("[LOG] Posting new embed for %s (%s)\n", c.Name, dc)
					err = b.discord.PostEmbedMessage(c.ID, listings, c.Duty, dc)
				}
				if err != nil {
					fmt.Printf("[ERROR] Discord error for %s (%s): %v\n", c.Name, dc, err)
				}
			}
		}

		fmt.Println("[LOG] Scrape complete, sleeping", b.waitTime)
		time.Sleep(b.waitTime)
	}
}

func (b *Bot) MessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	const prefix = "*"
	content := m.Content
	fmt.Printf("[LOG] Message from %s: %s\n", m.Author.Username, content)

	switch {
	case content == prefix+"reload":
		if err := b.LoadConfig(); err != nil {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Reload failed: %v", err))
			fmt.Println("[ERROR] Reload failed:", err)
		} else {
			b.discord.Channels = b.cfg.Channels
			s.ChannelMessageSend(m.ChannelID, "Config reloaded")
		}

	case content == prefix+"status":
		embed := &discordgo.MessageEmbed{
			Title: "Bot Status",
			Description: fmt.Sprintf(
				"Channels: %d\nScrape interval: %s\nDB: %s",
				len(b.discord.Channels), b.waitTime, b.dbPath,
			),
			Color: 0x00ff99,
		}
		s.ChannelMessageSendEmbed(m.ChannelID, embed)

	case content == prefix+"help":
		embed := &discordgo.MessageEmbed{
			Title: "Commands",
			Description: "```\n" +
				"*reload              Reload config from disk\n" +
				"*status              Show bot status\n" +
				"*scrape              Trigger immediate scrape\n" +
				"*interval <dur>      Set scrape interval (e.g. 2m)\n" +
				"*enable <name>       Enable a channel\n" +
				"*disable <name>      Disable a channel\n" +
				"*toggle <name>       Toggle enable/disable\n" +
				"*channels            List all configured channels\n" +
				"*lastmsg <name>      Show last bot message ID\n" +
				"*resetmsg <name>     Reset embed message for a channel\n" +
				"*forcepost <name>    Force-post listings immediately\n" +
				"*help                Show this help\n" +
				"```",
			Color: 0xff9900,
		}
		s.ChannelMessageSendEmbed(m.ChannelID, embed)

	case strings.HasPrefix(content, prefix+"interval "):
		val := strings.TrimPrefix(content, prefix+"interval ")
		d, err := time.ParseDuration(val)
		if err != nil || d < 30*time.Second {
			s.ChannelMessageSend(m.ChannelID, "Invalid duration (minimum 30s)")
			return
		}
		b.waitTime = d
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Scrape interval set to %s", d))

	case content == prefix+"scrape":
		fmt.Println("[LOG] Manual scrape triggered")
		go b.StartScrapingLoop()
		s.ChannelMessageSend(m.ChannelID, "Started scrape loop")

	case strings.HasPrefix(content, prefix+"enable "):
		name := strings.TrimPrefix(content, prefix+"enable ")
		b.discord.EnableChannel(name)
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Channel **%s** enabled", name))

	case strings.HasPrefix(content, prefix+"disable "):
		name := strings.TrimPrefix(content, prefix+"disable ")
		b.discord.DisableChannel(name)
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Channel **%s** disabled", name))

	case strings.HasPrefix(content, prefix+"toggle "):
		name := strings.TrimPrefix(content, prefix+"toggle ")
		ch := b.discord.GetChannelByName(name)
		if ch == nil {
			s.ChannelMessageSend(m.ChannelID, "Channel not found")
			return
		}
		if ch.Enabled {
			b.discord.DisableChannel(name)
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Channel **%s** disabled", name))
		} else {
			b.discord.EnableChannel(name)
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Channel **%s** enabled", name))
		}

	case content == prefix+"channels":
		var lines []string
		for _, c := range b.discord.Channels {
			status := "disabled"
			if c.Enabled {
				status = "enabled"
			}
			lines = append(lines, fmt.Sprintf("• **%s** — %s (%s)", c.Name, status, c.Duty))
		}
		if len(lines) == 0 {
			s.ChannelMessageSend(m.ChannelID, "No channels configured")
			return
		}
		s.ChannelMessageSend(m.ChannelID, "Configured channels:\n"+strings.Join(lines, "\n"))

	case strings.HasPrefix(content, prefix+"lastmsg "):
		name := strings.TrimPrefix(content, prefix+"lastmsg ")
		ch := b.discord.GetChannelByName(name)
		if ch == nil {
			s.ChannelMessageSend(m.ChannelID, "Channel not found")
			return
		}
		if ch.MessageID == "" {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("No stored message for **%s**", name))
		} else {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Last message ID for **%s**: `%s`", name, ch.MessageID))
		}

	case strings.HasPrefix(content, prefix+"resetmsg "):
		name := strings.TrimPrefix(content, prefix+"resetmsg ")
		ch := b.discord.GetChannelByName(name)
		if ch == nil {
			s.ChannelMessageSend(m.ChannelID, "Channel not found")
			return
		}
		if err := b.discord.ResetChannelMessage(ch); err != nil {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Reset failed: %v", err))
		} else {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Message for **%s** reset", name))
		}

	case strings.HasPrefix(content, prefix+"forcepost "):
		name := strings.TrimPrefix(content, prefix+"forcepost ")
		ch := b.discord.GetChannelByName(name)
		if ch == nil {
			s.ChannelMessageSend(m.ChannelID, "Channel not found")
			return
		}
		latest := b.scraper.LatestListings()
		if latest == nil {
			s.ChannelMessageSend(m.ChannelID, "No listings cached yet — run *scrape first")
			return
		}
		go func() {
			for _, dc := range ch.DataCentres {
				if err := b.discord.PostListings(ch.ID, latest, ch.Duty, dc); err != nil {
					fmt.Printf("[ERROR] forcepost for %s (%s): %v\n", name, dc, err)
				}
			}
		}()
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Force-posting listings for **%s**", name))
	}
}

func main() {
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		fmt.Fprintln(os.Stderr, "[FATAL] DISCORD_TOKEN env var not set")
		os.Exit(1)
	}

	configPath := "config.yaml"
	if p := os.Getenv("CONFIG_PATH"); p != "" {
		configPath = p
	}

	dbPath := "findingway.db"
	if p := os.Getenv("DB_PATH"); p != "" {
		dbPath = p
	}

	bot, err := NewBot(token, configPath, dbPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "[FATAL]", err)
		os.Exit(1)
	}

	if err := bot.LoadConfig(); err != nil {
		fmt.Fprintln(os.Stderr, "[FATAL]", err)
		os.Exit(1)
	}

	if err := bot.InitializeDiscord(); err != nil {
		fmt.Fprintln(os.Stderr, "[FATAL]", err)
		os.Exit(1)
	}

	bot.dg.AddHandler(bot.MessageCreate)

	go bot.StartScrapingLoop()

	fmt.Println("[LOG] Bot running — press CTRL+C to exit")
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	bot.GracefulShutdown()
}
