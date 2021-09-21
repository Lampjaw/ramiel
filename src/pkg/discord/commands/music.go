package discord

import (
	"fmt"
	"log"
	"strings"
	"time"

	"ramiel/pkg/musicplayer"

	"github.com/bwmarrin/discordgo"
)

var musicPlayer = map[string]*musicplayer.MusicPlayer{}

var MusicCommands = &CommandDefinition{
	BotCommands: []*discordgo.ApplicationCommand{
		{
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
			Name:        "stop",
			Description: "Stop playback",
		},
		{
			Name:        "resume",
			Description: "Resume playback",
		},
		{
			Name:        "queue",
			Description: "Get queue",
		},
		{
			Name:        "shuffle",
			Description: "Shuffle queue",
		},
		{
			Name:        "skip",
			Description: "Skip playing item",
		},
		{
			Name:        "clear",
			Description: "Clear queue",
		},
		{
			Name:        "nowplaying",
			Description: "See information on playing item",
		},
		{
			Name:        "disconnect",
			Description: "Disconnect from the voice channel",
		},
	},
	BotCommandHandlers: map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"play": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			voiceState, _ := s.State.VoiceState(i.GuildID, i.Member.User.ID)

			if voiceState == nil {
				sendMessageResponse(s, i, "You must be in a voice channel to use this command.")
				return
			}

			if musicPlayer[i.GuildID] != nil && musicPlayer[i.GuildID].GetChannelID() != voiceState.ChannelID {
				musicPlayer[i.GuildID].Exit()
				delete(musicPlayer, i.GuildID)
			}

			if musicPlayer[i.GuildID] == nil {
				var err error
				musicPlayer[i.GuildID], err = musicplayer.New(s, voiceState)
				if err != nil {
					sendMessageResponse(s, i, "Unable to join voice channel")
					log.Printf("[%s] Unable to join voice channel: %v", i.GuildID, err)
					return
				}
			}

			url := i.ApplicationCommandData().Options[0].StringValue()
			if url != "" {
				if strings.Contains(url, "playlist") {
					playlist, err := musicPlayer[i.GuildID].AddPlaylistToQueue(i.Member, url)
					if err != nil {
						sendMessageResponse(s, i, "Failed to add playlist")
						log.Printf("[%s] Failed to add playlist: %v", i.GuildID, err)
						return
					}

					sendMessageResponse(s, i, fmt.Sprintf("Adding %v songs to the queue from `%s`", len(playlist.Items), playlist.Title))
				} else {
					item, err := musicPlayer[i.GuildID].AddSongToQueue(i.Member, url)
					if err != nil {
						sendMessageResponse(s, i, "Failed to add track")
						log.Printf("[%s] Failed to add track: %v", i.GuildID, err)
						return
					}

					fields := make([]*discordgo.MessageEmbedField, 0)
					durr := item.Duration.Round(time.Minute)
					durr -= (durr / time.Hour) * time.Hour
					durrMins := durr / time.Minute
					durrSecs := durr / time.Second
					fields = append(fields,
						&discordgo.MessageEmbedField{
							Name:   "Channel",
							Value:  item.Author,
							Inline: true,
						},
						&discordgo.MessageEmbedField{
							Name:   "Song Duration",
							Value:  fmt.Sprintf("%02d:%02d", durrMins, durrSecs),
							Inline: true,
						},
						&discordgo.MessageEmbedField{
							Name:   "Estimated time until playing",
							Value:  fmt.Sprintf("%02d:%02d", durrMins, durrSecs),
							Inline: true,
						},
						&discordgo.MessageEmbedField{
							Name:   "Position in queue",
							Value:  fmt.Sprintf("%v", len(musicPlayer[i.GuildID].Queue())),
							Inline: true,
						})
					embed := &discordgo.MessageEmbed{
						Author: &discordgo.MessageEmbedAuthor{
							Name: "Added to queue",
						},
						Title: item.Title,
						URL:   item.Url,
						Thumbnail: &discordgo.MessageEmbedThumbnail{
							URL: item.ThumbnailURL,
						},
						Color:  0x070707,
						Fields: fields,
					}
					sendEmbedResponse(s, i, embed)
				}
			}

			go musicPlayer[i.GuildID].Play()

		},
		"stop": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if musicPlayer[i.GuildID] == nil {
				sendMessageResponse(s, i, "Not running in a channel!")
			}

			musicPlayer[i.GuildID].Stop()

			sendMessageResponse(s, i, "Stopped playing")
		},
		"resume": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if musicPlayer[i.GuildID] == nil {
				sendMessageResponse(s, i, "Not running in a channel!")
			}

			musicPlayer[i.GuildID].Resume()

			sendMessageResponse(s, i, "Resumed playing")
		},
		"queue": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("You entered %s!", i.Interaction.ApplicationCommandData().Name),
				},
			})
		},
		"shuffle": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if musicPlayer[i.GuildID] == nil {
				sendMessageResponse(s, i, "Not running in a channel!")
			}

			musicPlayer[i.GuildID].Shuffle()

			sendMessageResponse(s, i, "Queue shuffled")
		},
		"skip": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if musicPlayer[i.GuildID] == nil {
				sendMessageResponse(s, i, "Not running in a channel!")
			}

			musicPlayer[i.GuildID].Skip()

			sendMessageResponse(s, i, "Skipped currently playing audio")
		},
		"clear": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if musicPlayer[i.GuildID] == nil {
				sendMessageResponse(s, i, "Not running in a channel!")
			}

			musicPlayer[i.GuildID].ClearQueue()

			sendMessageResponse(s, i, "Cleared queue")
		},
		"nowplaying": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("You entered %s!", i.Interaction.ApplicationCommandData().Name),
				},
			})
		},
		"disconnect": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if musicPlayer[i.GuildID] == nil {
				sendMessageResponse(s, i, "Not running in a channel!")
			}

			musicPlayer[i.GuildID].Exit()
		},
	},
}

func sendMessageResponse(s *discordgo.Session, i *discordgo.InteractionCreate, message string) error {
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: message,
		},
	})
}

func sendEmbedResponse(s *discordgo.Session, i *discordgo.InteractionCreate, message *discordgo.MessageEmbed) error {
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				message,
			},
		},
	})
}
