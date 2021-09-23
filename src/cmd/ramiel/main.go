package main

import (
	"log"
	"os"
	"os/signal"

	"ramiel/internal/discord"

	"github.com/namsral/flag"
)

var (
	DiscordToken          string
	DiscordGuildID        string
	DiscordRemoveCommands bool
)

func init() {
	flag.StringVar(&DiscordToken, "DISCORD_TOKEN", "", "Discord bot access token")
	flag.StringVar(&DiscordGuildID, "DISCORD_GUILD_ID", "", "Test guild ID. If not passed - bot registers commands globally")
	flag.BoolVar(&DiscordRemoveCommands, "DISCORD_REMOVE_COMMANDS", false, "Remove all commands after shutdowning or not")

	flag.Parse()
}

func main() {
	if DiscordToken == "" {
		log.Fatalf("No discord bot token provided")
	}

	log.Println("Running...")

	discordConfig := &discord.DiscordConfig{
		BotToken:       DiscordToken,
		GuildID:        DiscordGuildID,
		RemoveCommands: DiscordRemoveCommands,
	}

	discordClient := discord.New(discordConfig)

	err := discordClient.Connect()
	if err != nil {
		log.Fatalln(err)
	}

	defer discordClient.Close()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, os.Kill)
	<-stop
	log.Println("Shutting down...")
}
