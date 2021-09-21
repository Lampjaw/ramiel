package discord

import "github.com/bwmarrin/discordgo"

type CommandDefinition struct {
	BotCommands        []*discordgo.ApplicationCommand
	BotCommandHandlers map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate)
}
