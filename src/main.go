package main

import (
	"log"
	"os"
	"os/signal"

	"ramiel/pkg/discord"

	"github.com/namsral/flag"
)

var (
	DiscordToken          = flag.String("DiscordToken", "", "Discord bot access token")
	DiscordGuildID        = flag.String("DiscordGuildId", "", "Test guild ID. If not passed - bot registers commands globally")
	DiscordRemoveCommands = flag.Bool("DiscordRemoveCommands", false, "Remove all commands after shutdowning or not")
)

func init() { flag.Parse() }

func main() {
	if DiscordToken == nil {
		log.Fatalf("No discord bot token provided")
	}

	log.Println("Running...")

	discordConfig := &discord.DiscordConfig{
		BotToken:       *DiscordToken,
		GuildID:        *DiscordGuildID,
		RemoveCommands: *DiscordRemoveCommands,
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
