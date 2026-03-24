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
	"github.com/bwmarrin/discordgo"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Channels []*discord.Channel `yaml:"channels"`
}

type Bot struct {
	discordToken string
	configPath   string
	discord      *discord.Discord
	scraper      *scraper.Scraper
	cfg          Config
	dg           *discordgo.Session

	waitTime time.Duration
}

func NewBot(discordToken, configPath string) (*Bot, error) {
	fmt.Println("[LOG] Initializing bot")
	dg, err := discordgo.New("Bot " + discordToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create Discord session: %w", err)
	}

	return &Bot{
		discordToken: discordToken,
		configPath:   configPath,
		scraper:      &scraper.Scraper{Url: "https://xivpf.com"},
		dg:           dg,
		waitTime:     3 * time.Minute,
	}, nil
}

func (b *Bot) LoadConfig() error {
	fmt.Println("[LOG] Loading config from", b.configPath)
	data, err := os.ReadFile(b.configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
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
	}

	if err := d.Start(); err != nil {
		return fmt.Errorf("failed to start Discord session: %w", err)
	}

	b.discord = d
	fmt.Println("[LOG] Discord session started")
	return nil
}

func (b *Bot) GracefulShutdown() {
	fmt.Println("[LOG] Shutting down bot")
	if b.discord.Session != nil {
		b.discord.Session.Close()
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

		// Store latest listings
		b.scraper.LastListings = listings

		if len(listings.Listings) == 0 {
			fmt.Println("[LOG] No listings found, sleeping", b.waitTime)
			time.Sleep(b.waitTime)
			continue
		}

		for _, c := range b.discord.Channels {
			if c == nil {
				continue
			}

			for _, dc := range c.DataCentres {
				msgID := b.discord.GetLastMessageID(c.ID, dc)
				if msgID != "" {
					fmt.Println("[LOG] Updating embed for", c.Name, dc)
					err = b.discord.UpdateEmbedMessage(c.ID, msgID, listings, c.Duty, dc)
				} else {
					fmt.Println("[LOG] Posting new embed for", c.Name, dc)
					err = b.discord.PostEmbedMessage(c.ID, listings, c.Duty, dc)
				}

				if err != nil {
					fmt.Println("[ERROR] Discord error:", err)
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

	prefix := "*"
	content := m.Content
	fmt.Println("[LOG] Received message:", content, "from", m.Author.Username)

	switch {
	case content == prefix+"reload":
		fmt.Println("[LOG] Reload command triggered")
		if err := b.LoadConfig(); err != nil {
			s.ChannelMessageSend(m.ChannelID, "Reload failed")
			fmt.Println("[ERROR] Reload failed:", err)
		} else {
			s.ChannelMessageSend(m.ChannelID, "Reloaded")
		}

	case content == prefix+"status":
		msg := fmt.Sprintf("Channels: %d\nScrape interval: %s", len(b.discord.Channels), b.waitTime)
		embed := &discordgo.MessageEmbed{
			Title:       "Bot Status",
			Description: msg,
			Color:       0x00ff99,
		}
		s.ChannelMessageSendEmbed(m.ChannelID, embed)

	case content == prefix+"help":
		embed := &discordgo.MessageEmbed{
			Title: "Commands",
			Description: `
!reload                     → Reload config
!status                     → Show bot status
!scrape                     → Start scrape loop
!interval <time>            → Set scraping interval (e.g., 2m)
!enable <channel>           → Enable a channel
!disable <channel>          → Disable a channel
!toggle <channel>           → Toggle enable/disable
!channels                   → List all configured channels
!lastmsg <channel>          → Show last bot message ID for a channel
!resetmsg <channel>         → Reset the embed message for a channel
!forcepost <channel>        → Force post listings immediately
!help                       → Show this help
`,
			Color: 0xff9900,
		}
		s.ChannelMessageSendEmbed(m.ChannelID, embed)

	case strings.HasPrefix(content, prefix+"interval "):
		val := strings.TrimPrefix(content, prefix+"interval ")
		d, err := time.ParseDuration(val)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Invalid duration")
			fmt.Println("[ERROR] Invalid interval:", val)
			return
		}
		b.waitTime = d
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Scrape interval updated to %s", d))
		fmt.Println("[LOG] Interval updated to", d)

	case content == prefix+"scrape":
		fmt.Println("[LOG] Manual scrape triggered")
		go b.StartScrapingLoop()
		s.ChannelMessageSend(m.ChannelID, "Started scrape loop")

	case strings.HasPrefix(content, prefix+"enable "):
		name := strings.TrimPrefix(content, prefix+"enable ")
		b.discord.EnableChannel(name)
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Channel **%s** enabled", name))
		fmt.Println("[LOG] Channel enabled:", name)

	case strings.HasPrefix(content, prefix+"disable "):
		name := strings.TrimPrefix(content, prefix+"disable ")
		b.discord.DisableChannel(name)
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Channel **%s** disabled", name))
		fmt.Println("[LOG] Channel disabled:", name)

	case strings.HasPrefix(content, prefix+"toggle "):
		name := strings.TrimPrefix(content, prefix+"toggle ")
		ch := b.discord.GetChannelByName(name)
		if ch != nil {
			if ch.Enabled {
				b.discord.DisableChannel(name)
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Channel **%s** disabled", name))
				fmt.Println("[LOG] Channel toggled off:", name)
			} else {
				b.discord.EnableChannel(name)
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Channel **%s** enabled", name))
				fmt.Println("[LOG] Channel toggled on:", name)
			}
		} else {
			s.ChannelMessageSend(m.ChannelID, "Channel not found")
			fmt.Println("[WARN] Channel not found:", name)
		}

	case content == prefix+"channels":
		var list []string
		for _, c := range b.discord.Channels {
			status := "disabled"
			if c.Enabled {
				status = "enabled"
			}
			list = append(list, fmt.Sprintf("%s → %s", c.Name, status))
		}
		s.ChannelMessageSend(m.ChannelID, "Configured channels:\n"+strings.Join(list, "\n"))
		fmt.Println("[LOG] Listed all channels")

	case strings.HasPrefix(content, prefix+"lastmsg "):
		name := strings.TrimPrefix(content, prefix+"lastmsg ")
		ch := b.discord.GetChannelByName(name)
		if ch != nil {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Last message ID for **%s**: %s", name, ch.MessageID))
			fmt.Println("[LOG] Last message for", name, "is", ch.MessageID)
		} else {
			s.ChannelMessageSend(m.ChannelID, "Channel not found")
			fmt.Println("[WARN] Channel not found:", name)
		}

	case strings.HasPrefix(content, prefix+"resetmsg "):
		name := strings.TrimPrefix(content, prefix+"resetmsg ")
		ch := b.discord.GetChannelByName(name)
		if ch != nil {
			if err := b.discord.ResetChannelMessage(ch); err != nil {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Failed to reset message: %v", err))
				fmt.Println("[ERROR] Reset message failed for", name, err)
			} else {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Message for **%s** reset", name))
				fmt.Println("[LOG] Message reset for channel:", name)
			}
		} else {
			s.ChannelMessageSend(m.ChannelID, "Channel not found")
			fmt.Println("[WARN] Channel not found:", name)
		}

	case strings.HasPrefix(content, prefix+"forcepost "):
		name := strings.TrimPrefix(content, prefix+"forcepost ")
		ch := b.discord.GetChannelByName(name)
		if ch != nil {
			fmt.Println("[LOG] Force post triggered for channel:", name)
			go func() {
				for _, dc := range ch.DataCentres {
					_ = b.discord.PostListings(ch.ID, b.scraper.LatestListings(), ch.Duty, dc)
				}
			}()
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Force posted listings for **%s**", name))
		} else {
			s.ChannelMessageSend(m.ChannelID, "Channel not found")
			fmt.Println("[WARN] Channel not found for forcepost:", name)
		}
	}
}

func main() {
	token := ""
	config := "config.yaml"

	bot, err := NewBot(token, config)
	if err != nil {
		panic(err)
	}

	err = bot.dg.Open()
	if err != nil {
		panic(err)
	}

	bot.dg.AddHandler(bot.MessageCreate)

	if err := bot.LoadConfig(); err != nil {
		panic(err)
	}

	if err := bot.InitializeDiscord(); err != nil {
		panic(err)
	}

	go bot.StartScrapingLoop()

	fmt.Println("[LOG] Bot is running, press CTRL+C to exit")
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	bot.GracefulShutdown()
}