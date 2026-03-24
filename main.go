package main

import (
	"fmt"
	"os"
	"time"

	"github.com/Veraticus/findingway/internal/discord"
	"github.com/Veraticus/findingway/internal/scraper"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Channels []*discord.Channel `yaml:"channels"`
}

func main() {
	discordToken := "MTQ4NTg2MTUwNDg1MDA3MTU3Mg.Gg1-Nn.jsRFIXCtBLecJ9Jn1Q8IoQx7H4BExifOU69JXw"

	// Load channels from config.yaml
	configData, err := os.ReadFile("config.yaml")
	if err != nil {
		panic(fmt.Errorf("Could not read config.yaml: %v", err))
	}

	var cfg Config
	if err := yaml.Unmarshal(configData, &cfg); err != nil {
		panic(fmt.Errorf("Could not parse config.yaml: %v", err))
	}

	d := &discord.Discord{
		Token:    discordToken,
		Channels: cfg.Channels,
	}

	fmt.Println("Initializing Discord...")
	if len(d.Channels) == 0 {
		panic("No channels provided in the configuration!")
	}
	fmt.Printf("Loaded %d channels from config.yaml\n", len(d.Channels))
	for _, c := range d.Channels {
		fmt.Printf("  - %s (%s) @ %v\n", c.Name, c.Duty, c.DataCentres)
	}

	err = d.Start()
	if err != nil {
		panic(fmt.Errorf("Could not instantiate Discord: %v", err))
	}
	defer func() {
		if d.Session != nil {
			d.Session.Close()
		}
	}()

	scraperInstance := &scraper.Scraper{Url: "https://xivpf.com"}

	fmt.Printf("Starting findingway...\n")

	for {
		totalWait := 3 * time.Minute
		fmt.Printf("Scraping source...\n")

		listings, err := scraperInstance.Scrape()
		if err != nil {
			fmt.Printf("Scraper error: %v\n", err)
			time.Sleep(30 * time.Second)
			continue
		}

		if len(listings.Listings) == 0 {
			fmt.Println("No listings found, skipping this cycle.")
			time.Sleep(totalWait)
			continue
		}

		fmt.Printf("Got %v listings.\n", len(listings.Listings))
		fmt.Printf("Sending to %v channels...\n", len(d.Channels))

		for _, c := range d.Channels {
			if c == nil {
				fmt.Println("Channel is nil!")
				continue
			}

			startTime := time.Now()
			fmt.Printf("Updating Discord for %v (%v)...\n", c.Name, c.Duty)

			for _, dc := range c.DataCentres {
				err = d.PostListings(c.ID, listings, c.Duty, dc)
				if err != nil {
					fmt.Printf("Discord error updating messages: %v\n", err)
				}
			}

			duration := time.Since(startTime)
			totalWait -= duration
		}

		fmt.Printf("Sleeping for %v...\n", totalWait)
		time.Sleep(totalWait)
	}
}