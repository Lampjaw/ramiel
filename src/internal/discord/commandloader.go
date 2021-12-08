package discord

import (
	"fmt"

	commands "ramiel/internal/discord/commands"

	"github.com/bwmarrin/discordgo"
)

var (
	commandsets = []*commands.CommandDefinition{
		commands.MusicCommands,
	}
	botCommands        = []*discordgo.ApplicationCommand{}
	botCommandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){}
)

func init() {
	for _, set := range commandsets {
		for key := range set.BotCommandHandlers {
			botCommandHandlers[key] = set.BotCommandHandlers[key]
		}
		botCommands = append(botCommands, set.BotCommands...)
	}
}

func (d *DiscordClient) loadCommands() error {
	for _, set := range commandsets {
		set.BotCommandInitializer(d.session)
	}

	d.session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := botCommandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})

	for _, command := range botCommands {
		_, err := d.session.ApplicationCommandCreate(d.session.State.User.ID, *&d.config.GuildID, command)
		if err != nil {
			return fmt.Errorf("Cannot create '%v' command: %v", command.Name, err)
		}
	}

	return nil
}

func (d *DiscordClient) removeCommands() error {
	cmds, err := d.session.ApplicationCommands(d.session.State.User.ID, d.config.GuildID)
	if err != nil {
		return fmt.Errorf("Failed to retrieve commands list: %v", err)
	}

	for _, command := range cmds {
		err := d.session.ApplicationCommandDelete(d.session.State.User.ID, *&d.config.GuildID, command.ID)
		if err != nil {
			return fmt.Errorf("Cannot delete '%v' command: %v", command.Name, err)
		}
	}

	return nil
}
