package main

import (
	"discord-uptime-checker/constants"
	"discord-uptime-checker/service"
	config2 "discord-uptime-checker/tasks/config"
	"discord-uptime-checker/tasks/export"
	"github.com/bwmarrin/discordgo"
	"log"
)

func main() {
	discord, err := discordgo.New("Bot " + constants.DiscordToken)
	if err != nil {
		log.Fatalf("error creating Discord session: %v", err)
	}

	config, err := config2.LoadConfig()
	if err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	discord.Identify.Intents |= discordgo.IntentGuildMessages
	discord.Identify.Intents |= discordgo.IntentGuilds
	if err := discord.Open(); err != nil {
		log.Fatalf("error connecting to Discord: %v", err)
	}

	currentUser := discord.State.User
	log.Printf("Connected to Discord as %v#%v (%v)", currentUser.Username, currentUser.Discriminator, currentUser.ID)
	registry, serve, err := export.ServeMetrics()
	if err != nil {
		log.Fatalf("error creating metrics: %v", err)
	}

	checkService := service.NewCheckService(discord, config, registry)

	checkService.Start()
	if err := serve(); err != nil {
		log.Fatalf("error serving metrics: %v", err)
	}
}
