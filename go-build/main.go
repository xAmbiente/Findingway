package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Veraticus/findingway/internal/discord"
	"github.com/Veraticus/findingway/internal/ffxiv"
	"github.com/Veraticus/findingway/internal/logger"
	"github.com/Veraticus/findingway/internal/scraper"
	"github.com/Veraticus/findingway/internal/store"
	"github.com/bwmarrin/discordgo"
	"gopkg.in/yaml.v3"
)

// ── Config ────────────────────────────────────────────────────────────────────

type Config struct {
	Channels             []*discord.Channel `yaml:"channels"`
	AnnouncementsChannel string             `yaml:"announcementsChannel"`
}

// ── Bot ───────────────────────────────────────────────────────────────────────

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
	Slash        *discord.SlashCommandManager

	lastSnapshot map[string]map[string]map[string]*ffxiv.Listing
	// channelID -> dataCentre -> listingID -> listing
}

func NewBot(discordToken, configPath, dbPath string) (*Bot, error) {
	logger.Info("initializing bot (db: %s)", dbPath)

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
		lastSnapshot: make(map[string]map[string]map[string]*ffxiv.Listing),
	}, nil
}

func (b *Bot) LoadConfig() error {
	logger.Info("loading config from %s", b.configPath)
	data, err := os.ReadFile(b.configPath)
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("parse config: %w", err)
	}

	b.cfg = cfg
	if b.discord != nil && cfg.AnnouncementsChannel != "" {
		b.discord.AnnouncementsChannel = cfg.AnnouncementsChannel
	}
	logger.Info("config loaded — %d channel(s), announcements: %q", len(cfg.Channels), cfg.AnnouncementsChannel)
	return nil
}

func (b *Bot) InitializeDiscord() error {
	logger.Info("connecting to Discord…")

	d := &discord.Discord{
		Token:                b.discordToken,
		Channels:             b.cfg.Channels,
		Store:                b.store,
		AnnouncementsChannel: b.cfg.AnnouncementsChannel,
		AnnouncedListings:    make(map[string]struct{}),
	}

	if err := d.Start(); err != nil {
		return fmt.Errorf("start discord: %w", err)
	}

	b.discord = d

	// Use the live session from the discord wrapper for handlers
	b.dg = d.Session
	b.dg.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if b.Slash != nil {
			b.Slash.HandleInteraction(i)
		}
	})

	logger.Info("Discord session ready")
	return nil
}

func (b *Bot) getSnapshot(channelID, dc string) map[string]*ffxiv.Listing {
	if b.lastSnapshot[channelID] == nil {
		b.lastSnapshot[channelID] = make(map[string]map[string]*ffxiv.Listing)
	}
	if b.lastSnapshot[channelID][dc] == nil {
		b.lastSnapshot[channelID][dc] = make(map[string]*ffxiv.Listing)
	}
	return b.lastSnapshot[channelID][dc]
}

func (b *Bot) setSnapshot(channelID, dc string, listings []*ffxiv.Listing) {
	if b.lastSnapshot[channelID] == nil {
		b.lastSnapshot[channelID] = make(map[string]map[string]*ffxiv.Listing)
	}
	newMap := make(map[string]*ffxiv.Listing)
	for i := range listings {
		l := listings[i]
		newMap[l.Id] = l
	}
	b.lastSnapshot[channelID][dc] = newMap
}

// ── Scraping loop ─────────────────────────────────────────────────────────────

func (b *Bot) StartScrapingLoop() {
	logger.Info("scraping loop started (interval: %s)", b.waitTime)
	for {
		b.runScrape()
		time.Sleep(b.waitTime)
	}
}

// runScrape performs a single scrape-and-post cycle.
func (b *Bot) runScrape() {
	logger.Info("scraping xivpf.com…")
	listings, err := b.scraper.Scrape()
	if err != nil {
		logger.Error("scrape failed: %v — retrying in 30s", err)
		time.Sleep(30 * time.Second)
		return
	}

	b.scraper.LastListings = listings
	count := len(listings.Listings)
	logger.Info("scraped %d listing(s)", count)

	if count == 0 {
		logger.Info("no listings — skipping post")
		return
	}

	// ── Merc / payment announcement ─────────────────────────────────────────
	b.discord.CheckAndAnnounceMercListings(listings)

	for _, c := range b.discord.Channels {
		if c == nil || !c.Enabled {
			continue
		}

		for _, dc := range c.DataCentres {
			msgID := b.discord.GetLastMessageID(c.ID, dc)

			// Build new snapshot (slice + map)
			newSnap := make(map[string]*ffxiv.Listing)
			var newSlice []*ffxiv.Listing
			for i := range listings.Listings {
				l := listings.Listings[i]
				if l.DataCentre != dc {
					continue
				}
				newSnap[l.Id] = l
				newSlice = append(newSlice, l)
			}

			// Log removed listings
			oldSnap := b.getSnapshot(c.ID, dc)
			for id := range oldSnap {
				if _, ok := newSnap[id]; !ok {
					logger.Debug("removed listing detected: %s (%s / %s)", id, c.Name, dc)
				}
			}

			// Update cache
			b.setSnapshot(c.ID, dc, newSlice)

			// Post / update Discord embed
			var postErr error
			if msgID != "" {
				postErr = b.discord.UpdateEmbedMessage(c.ID, msgID, listings, c.Duty, dc)
			} else {
				postErr = b.discord.PostEmbedMessage(c.ID, listings, c.Duty, dc)
			}

			if postErr != nil {
				logger.Error("discord post failed for %s (%s): %v", c.Name, dc, postErr)
			} else {
				logger.Info("updated embed — %s / %s", c.Name, dc)
			}
		}
	}
}

// ── Discord commands ──────────────────────────────────────────────────────────

func (b *Bot) MessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	const prefix = "*"
	content := m.Content

	if strings.HasPrefix(content, prefix) {
		logger.Debug("command from %s: %s", m.Author.Username, content)
	}

	switch {
	// ── *reload ──────────────────────────────────────────────────────────────
	case content == prefix+"reload":
		if err := b.LoadConfig(); err != nil {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("❌ Reload failed: %v", err))
			logger.Error("reload failed: %v", err)
			return
		}
		b.discord.Channels = b.cfg.Channels
		s.ChannelMessageSend(m.ChannelID, "✅ Config reloaded")

	// ── *status ───────────────────────────────────────────────────────────────
	case content == prefix+"status":
		enabled := 0
		for _, c := range b.discord.Channels {
			if c.Enabled {
				enabled++
			}
		}
		annCh := b.discord.AnnouncementsChannel
		if annCh == "" {
			annCh = "not set"
		}
		embed := &discordgo.MessageEmbed{
			Title: "Findingway Status",
			Fields: []*discordgo.MessageEmbedField{
				{Name: "Channels", Value: fmt.Sprintf("%d total, %d enabled", len(b.discord.Channels), enabled), Inline: true},
				{Name: "Scrape interval", Value: b.waitTime.String(), Inline: true},
				{Name: "Database", Value: b.dbPath, Inline: false},
				{Name: "Announcements channel", Value: annCh, Inline: false},
			},
			Color: 0x00ff99,
		}
		s.ChannelMessageSendEmbed(m.ChannelID, embed)

	// ── *help ─────────────────────────────────────────────────────────────────
	case content == prefix+"help":
		embed := &discordgo.MessageEmbed{
			Title: "Commands",
			Description: "```\n" +
				"*reload                 Reload config.yaml from disk\n" +
				"*status                 Show bot status\n" +
				"*scrape                 Trigger an immediate scrape\n" +
				"*interval <dur>         Set scrape interval (e.g. 2m, min 30s)\n" +
				"*enable  <name>         Enable a channel\n" +
				"*disable <name>         Disable a channel\n" +
				"*toggle  <name>         Toggle a channel on/off\n" +
				"*channels               List all configured channels\n" +
				"*lastmsg <name>         Show the stored embed message ID\n" +
				"*resetmsg <name>        Clear embed so a fresh one is posted\n" +
				"*forcepost <name>       Immediately post cached listings\n" +
				"*announce <channelID>   Set announcements channel (merc alerts)\n" +
				"*announce off           Disable merc announcements\n" +
				"*listingcount           Show how many listings are currently cached\n" +
				"*clearold               Remove announced-listing cache (merc tracker)\n" +
				"*help                   This message\n" +
				"```",
			Color: 0xff9900,
		}
		s.ChannelMessageSendEmbed(m.ChannelID, embed)

	// ── *interval ─────────────────────────────────────────────────────────────
	case strings.HasPrefix(content, prefix+"interval "):
		val := strings.TrimPrefix(content, prefix+"interval ")
		d, err := time.ParseDuration(val)
		if err != nil || d < 30*time.Second {
			s.ChannelMessageSend(m.ChannelID, "❌ Invalid duration (minimum 30s, e.g. `*interval 2m`)")
			return
		}
		b.waitTime = d
		logger.Info("scrape interval changed to %s by %s", d, m.Author.Username)
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("✅ Scrape interval set to **%s**", d))

	// ── *scrape ───────────────────────────────────────────────────────────────
	case content == prefix+"scrape":
		logger.Info("manual scrape triggered by %s", m.Author.Username)
		go b.runScrape()
		s.ChannelMessageSend(m.ChannelID, "✅ Scrape triggered")

	// ── *enable / *disable / *toggle ─────────────────────────────────────────
	case strings.HasPrefix(content, prefix+"enable "):
		name := strings.TrimPrefix(content, prefix+"enable ")
		b.discord.EnableChannel(name)
		logger.Info("channel %q enabled by %s", name, m.Author.Username)
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("✅ Channel **%s** enabled", name))

	case strings.HasPrefix(content, prefix+"disable "):
		name := strings.TrimPrefix(content, prefix+"disable ")
		b.discord.DisableChannel(name)
		logger.Info("channel %q disabled by %s", name, m.Author.Username)
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("✅ Channel **%s** disabled", name))

	case strings.HasPrefix(content, prefix+"toggle "):
		name := strings.TrimPrefix(content, prefix+"toggle ")
		ch := b.discord.GetChannelByName(name)
		if ch == nil {
			s.ChannelMessageSend(m.ChannelID, "❌ Channel not found")
			return
		}
		if ch.Enabled {
			b.discord.DisableChannel(name)
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("✅ Channel **%s** disabled", name))
		} else {
			b.discord.EnableChannel(name)
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("✅ Channel **%s** enabled", name))
		}
		logger.Info("channel %q toggled by %s (now: enabled=%v)", name, m.Author.Username, !ch.Enabled)

	// ── *channels ─────────────────────────────────────────────────────────────
	case content == prefix+"channels":
		if len(b.discord.Channels) == 0 {
			s.ChannelMessageSend(m.ChannelID, "No channels configured")
			return
		}
		var fields []*discordgo.MessageEmbedField
		for _, c := range b.discord.Channels {
			status := "🔴 disabled"
			if c.Enabled {
				status = "🟢 enabled"
			}
			fields = append(fields, &discordgo.MessageEmbedField{
				Name:   c.Name,
				Value:  fmt.Sprintf("%s\n%s\nDC: %s", status, c.Duty, strings.Join(c.DataCentres, ", ")),
				Inline: true,
			})
		}
		s.ChannelMessageSendEmbed(m.ChannelID, &discordgo.MessageEmbed{
			Title:  "Configured Channels",
			Fields: fields,
			Color:  0x6600ff,
		})

	// ── *lastmsg ──────────────────────────────────────────────────────────────
	case strings.HasPrefix(content, prefix+"lastmsg "):
		name := strings.TrimPrefix(content, prefix+"lastmsg ")
		ch := b.discord.GetChannelByName(name)
		if ch == nil {
			s.ChannelMessageSend(m.ChannelID, "❌ Channel not found")
			return
		}
		if ch.MessageID == "" {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("No stored message for **%s**", name))
		} else {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Last embed ID for **%s**: `%s`", name, ch.MessageID))
		}

	// ── *resetmsg ─────────────────────────────────────────────────────────────
	case strings.HasPrefix(content, prefix+"resetmsg "):
		name := strings.TrimPrefix(content, prefix+"resetmsg ")
		ch := b.discord.GetChannelByName(name)
		if ch == nil {
			s.ChannelMessageSend(m.ChannelID, "❌ Channel not found")
			return
		}
		if err := b.discord.ResetChannelMessage(ch); err != nil {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("❌ Reset failed: %v", err))
			logger.Error("resetmsg failed for %s: %v", name, err)
			return
		}
		logger.Info("embed reset for %s by %s", name, m.Author.Username)
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("✅ Embed reset for **%s** — fresh post next cycle", name))

	// ── *forcepost ────────────────────────────────────────────────────────────
	case strings.HasPrefix(content, prefix+"forcepost "):
		name := strings.TrimPrefix(content, prefix+"forcepost ")
		ch := b.discord.GetChannelByName(name)
		if ch == nil {
			s.ChannelMessageSend(m.ChannelID, "❌ Channel not found")
			return
		}
		latest := b.scraper.LatestListings()
		if latest == nil {
			s.ChannelMessageSend(m.ChannelID, "❌ No cached listings yet — wait for the first scrape or run `*scrape`")
			return
		}
		go func() {
			for _, dc := range ch.DataCentres {
				if err := b.discord.PostListings(ch.ID, latest, ch.Duty, dc); err != nil {
					logger.Error("forcepost for %s (%s): %v", name, dc, err)
				}
			}
		}()
		logger.Info("force-post triggered for %s by %s", name, m.Author.Username)
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("✅ Force-posting cached listings for **%s**", name))

	// ── *announce ─────────────────────────────────────────────────────────────
	// *announce <channelID>   — set the announcements channel
	// *announce off           — disable merc announcements
	case strings.HasPrefix(content, prefix+"announce"):
		arg := strings.TrimSpace(strings.TrimPrefix(content, prefix+"announce"))
		if arg == "" {
			cur := b.discord.AnnouncementsChannel
			if cur == "" {
				cur = "not set"
			}
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("ℹ️ Current announcements channel: **%s**\nUse `*announce <channelID>` to set or `*announce off` to disable.", cur))
			return
		}
		if strings.EqualFold(arg, "off") {
			b.discord.AnnouncementsChannel = ""
			logger.Info("merc announcements disabled by %s", m.Author.Username)
			s.ChannelMessageSend(m.ChannelID, "✅ Merc/payment announcements **disabled**")
			return
		}
		b.discord.AnnouncementsChannel = arg
		logger.Info("announcements channel set to %s by %s", arg, m.Author.Username)
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("✅ Announcements channel set to <#%s>", arg))

	// ── *listingcount ─────────────────────────────────────────────────────────
	case content == prefix+"listingcount":
		latest := b.scraper.LatestListings()
		if latest == nil {
			s.ChannelMessageSend(m.ChannelID, "❌ No cached listings yet")
			return
		}
		// Count per-channel
		var fields []*discordgo.MessageEmbedField
		for _, c := range b.discord.Channels {
			if !c.Enabled {
				continue
			}
			for _, dc := range c.DataCentres {
				scoped := latest.ForDutyAndDataCentre(c.Duty, dc)
				// count active only
				active := 0
				now := time.Now()
				for _, l := range scoped.Listings {
					exp, err := l.ExpiresAt()
					if err != nil || exp.After(now) {
						active++
					}
				}
				fields = append(fields, &discordgo.MessageEmbedField{
					Name:   fmt.Sprintf("%s / %s", c.Name, dc),
					Value:  fmt.Sprintf("%d active listings", active),
					Inline: true,
				})
			}
		}
		embed := &discordgo.MessageEmbed{
			Title:  "📊 Listing Counts (active, not expired)",
			Fields: fields,
			Color:  0x00ccff,
			Footer: &discordgo.MessageEmbedFooter{Text: fmt.Sprintf("Total in cache: %d", len(latest.Listings))},
		}
		s.ChannelMessageSendEmbed(m.ChannelID, embed)

	// ── *clearold ─────────────────────────────────────────────────────────────
	case content == prefix+"clearold":
		b.discord.AnnouncedListings = make(map[string]struct{})
		logger.Info("merc-announcement cache cleared by %s", m.Author.Username)
		s.ChannelMessageSend(m.ChannelID, "✅ Merc announcement cache cleared — already-seen listings will be re-evaluated next scrape")
	}
}

// ── main ──────────────────────────────────────────────────────────────────────

func main() {
	logPath := os.Getenv("LOG_PATH")
	if logPath == "" {
		logPath = "logs/findingway.log"
	}
	if err := logger.Init(logPath); err != nil {
		fmt.Fprintln(os.Stderr, "FATAL: could not init logger:", err)
		os.Exit(1)
	}
	defer logger.Close()

	logger.Section("Findingway")

	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		logger.Fatal("DISCORD_TOKEN env var is not set")
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
		logger.Fatal("init failed: %v", err)
	}

	if err := bot.LoadConfig(); err != nil {
		logger.Fatal("config load failed: %v", err)
	}

	if err := bot.InitializeDiscord(); err != nil {
		logger.Fatal("discord init failed: %v", err)
	}

	// Initialize slash commands manager and register commands (global registration)
	bot.Slash = &discord.SlashCommandManager{Session: bot.dg, Bot: bot}
	if err := bot.Slash.RegisterCommands(""); err != nil {
		logger.Warn("could not register slash commands: %v", err)
	} else {
		logger.Info("slash commands registered (global)")
	}

	bot.dg.AddHandler(bot.MessageCreate)

	go bot.StartScrapingLoop()

	logger.Info("bot is running — send SIGINT/SIGTERM to stop")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	bot.GracefulShutdown()
}

// ── Bot helpers ───────────────────────────────────────────────────────────────

func (b *Bot) EnableChannel(name string) {
	if b.discord != nil {
		b.discord.EnableChannel(name)
	}
}

func (b *Bot) DisableChannel(name string) {
	if b.discord != nil {
		b.discord.DisableChannel(name)
	}
}

func (b *Bot) GetChannelByName(name string) *discord.Channel {
	if b.discord == nil {
		return nil
	}
	return b.discord.GetChannelByName(name)
}

func (b *Bot) GetChannels() []*discord.Channel {
	if b.discord == nil {
		return nil
	}
	return b.discord.Channels
}

func (b *Bot) GetWaitTime() time.Duration {
	return b.waitTime
}

func (b *Bot) SetWaitTime(d time.Duration) {
	b.waitTime = d
}

func (b *Bot) GracefulShutdown() {
	logger.Info("graceful shutdown initiated")

	if b.discord != nil {
		if err := b.discord.Close(); err != nil {
			logger.Error("failed to close discord wrapper: %v", err)
		} else {
			logger.Info("discord wrapper closed")
		}
	}

	if b.dg != nil {
		if err := b.dg.Close(); err != nil {
			logger.Error("failed to close discord session: %v", err)
		} else {
			logger.Info("discord session closed")
		}
	}

	logger.Info("graceful shutdown complete")
}

func (b *Bot) GetAnnouncementsChannel() string {
	if b.discord == nil {
		return ""
	}
	return b.discord.AnnouncementsChannel
}

func (b *Bot) SetAnnouncementsChannel(id string) {
	if b.discord != nil {
		b.discord.AnnouncementsChannel = id
	}
}

func (b *Bot) GetCachedListingCount() int {
	latest := b.scraper.LatestListings()
	if latest == nil {
		return -1
	}
	return len(latest.Listings)
}

func (b *Bot) ClearMercCache() {
	if b.discord != nil {
		b.discord.AnnouncedListings = make(map[string]struct{})
	}
}

func (b *Bot) ForceScrape() error {
	go func() {
		listings, err := b.scraper.Scrape()
		if err != nil {
			logger.Error("force scrape failed: %v", err)
			return
		}

		b.scraper.LastListings = listings
		b.discord.CheckAndAnnounceMercListings(listings)

		for _, c := range b.discord.Channels {
			if c == nil || !c.Enabled {
				continue
			}
			for _, dc := range c.DataCentres {
				if err := b.discord.PostListings(c.ID, listings, c.Duty, dc); err != nil {
					logger.Error("forcepost for %s (%s): %v", c.Name, dc, err)
				}
			}
		}
	}()

	return nil
}
