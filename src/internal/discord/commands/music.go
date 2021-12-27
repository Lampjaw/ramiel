package discord

import (
	"log"

	"ramiel/internal/discordutils"
	"ramiel/internal/musicplayer"

	"github.com/bwmarrin/discordgo"
)

var playerInstanceManager *musicplayer.DiscordPlayerInstanceManager

var MusicCommands = &CommandDefinition{
	BotCommandInitializer: func(s *discordgo.Session) {
		playerInstanceManager = musicplayer.NewDiscordPlayerInstanceManager(s)

		s.AddHandler(func(s *discordgo.Session, event *discordgo.VoiceStateUpdate) {
			if discordutils.IsBotUser(s, event.UserID) && playerInstanceManager.VerifyPlayerInstanceExists(event.GuildID) && event.ChannelID == "" {
				if err := playerInstanceManager.DestroyMusicPlayerInstance(event.GuildID); err != nil {
					log.Printf("Error when destroying instance: %v", err)
				}
			}
		})

		s.AddHandler(func(s *discordgo.Session, event *discordgo.VoiceServerUpdate) {
			instance := playerInstanceManager.GetPlayerInstance(event.GuildID)
			if instance != nil {
				instance.VoiceServerUpdate(s, event)
			}
		})
	},
	BotCommands: []*discordgo.ApplicationCommand{
		{
			Name:        "music",
			Description: "Music commands",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "play",
					Description: "Play music",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "url",
							Description: "YouTube video or playlist URL",
							Required:    true,
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "stop",
					Description: "Stop playback",
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "resume",
					Description: "Resume playback",
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "queue",
					Description: "Get queue",
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "shuffle",
					Description: "Shuffle queue",
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "skip",
					Description: "Skip playing item",
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "clear",
					Description: "Clear queue",
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "nowplaying",
					Description: "See information on playing item",
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "disconnect",
					Description: "Disconnect from the voice channel",
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "loop",
					Description: "Toggle queue looping",
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "removeduplicates",
					Description: "Remove duplicate items",
				},
			},
		},
	},
	BotCommandHandlers: map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"music": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			voiceState, _ := s.State.VoiceState(i.GuildID, i.Member.User.ID)
			if voiceState == nil {
				sendMessageResponse(s, i, "You must be in a voice channel to use this command")
				return
			}

			if i.ApplicationCommandData().Options[0].Name == "play" {
				playerChannelID := discordutils.GetInteractionUserVoiceChannelID(s, i.Interaction)
				instance, err := playerInstanceManager.GetOrCreatePlayerInstance(i.GuildID, playerChannelID, i.ChannelID)
				if err != nil {
					sendMessageResponse(s, i, err.Error())
					return
				}

				url := i.ApplicationCommandData().Options[0].Options[0].StringValue()

				if err := instance.Play(url, i); err != nil {
					sendMessageResponse(s, i, err.Error())
					return
				}

				return
			}

			instance := playerInstanceManager.GetPlayerInstance(i.GuildID)
			if instance == nil {
				sendMessageResponse(s, i, "Not running in a channel! Try playing something first.")
				return
			}

			switch i.ApplicationCommandData().Options[0].Name {
			case "stop":
				instance.Stop(i)
			case "resume":
				instance.Resume(i)
			case "queue":
				instance.Queue(i)
			case "shuffle":
				instance.Shuffle(i)
			case "skip":
				instance.Skip(i)
			case "clear":
				instance.Clear(i)
			case "nowplaying":
				instance.NowPlaying(i)
			case "loop":
				instance.Loop(i)
			case "removeduplicates":
				instance.RemoveDuplicates(i)
			case "disconnect":
				if err := playerInstanceManager.DestroyMusicPlayerInstance(i.GuildID); err != nil {
					log.Printf("Error when disconnecting: %v", err)
					sendMessageResponse(s, i, "Unable to disconnect player. Try disconnecting manually.")
				} else {
					sendMessageResponse(s, i, "See ya!")
				}
				return
			}
		},
	},
}

func sendMessageResponse(s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
	if err := discordutils.SendMessageResponse(s, i.Interaction, message); err != nil {
		log.Println(err)
	}
}
