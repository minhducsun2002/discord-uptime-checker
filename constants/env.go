package constants

import "os"

var (
	DiscordToken = os.Getenv("DISCORD_TOKEN")
	Port         = os.Getenv("PORT")
)
