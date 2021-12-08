package discord

import (
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	"ramiel/internal/musicplayer"

	"github.com/bwmarrin/discordgo"
)

var guildMusicPlayers = map[string]*musicplayer.MusicPlayer{}

var MusicCommands = &CommandDefinition{
	BotCommandInitializer: func(s *discordgo.Session) {
		s.AddHandler(func(s *discordgo.Session, event *discordgo.VoiceStateUpdate) {
			if guildMusicPlayers[event.GuildID] != nil && event.ChannelID == "" {
				if err := disconnectPlayer(s, event.GuildID); err != nil {
					log.Printf("Error when disconnecting: %v", err)
				}
			}
		})

		s.AddHandler(func(s *discordgo.Session, event *discordgo.VoiceServerUpdate) {
			if guildMusicPlayers[event.GuildID] != nil {
				guildMusicPlayers[event.GuildID].VoiceServerUpdate(s, event)
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
			musicPlayer, err := getMusicPlayerInstance(s, i)
			if err != nil {
				sendMessageResponse(s, i, err.Error())
				return
			}

			switch i.ApplicationCommandData().Options[0].Name {
			case "play":
				url := i.ApplicationCommandData().Options[0].Options[0].StringValue()
				if url != "" {
					if strings.Contains(url, "playlist") {
						playlist, err := musicPlayer.AddPlaylistToQueue(i.Member, url)
						if err != nil {
							sendMessageResponse(s, i, "Failed to add playlist")
							log.Printf("[%s] Failed to add playlist: %v", i.GuildID, err)
							return
						}

						sendMessageResponse(s, i, fmt.Sprintf("Adding %v songs to the queue from `%s`", len(playlist.Items), playlist.Title))
					} else {
						item, err := musicPlayer.AddSongToQueue(i.Member, url)
						if err != nil {
							sendMessageResponse(s, i, "Failed to add track")
							log.Printf("[%s] Failed to add track: %v", i.GuildID, err)
							return
						}

						fields := make([]*discordgo.MessageEmbedField, 0)

						fields = append(fields,
							&discordgo.MessageEmbedField{
								Name:   "Channel",
								Value:  item.Author,
								Inline: true,
							},
							&discordgo.MessageEmbedField{
								Name:   "Song Duration",
								Value:  getDurationString(item.Duration),
								Inline: true,
							},
							&discordgo.MessageEmbedField{
								Name:   "Estimated time until playing",
								Value:  getDurationString(musicPlayer.GetTotalQueueTime() - item.Duration),
								Inline: true,
							},
							&discordgo.MessageEmbedField{
								Name:   "Position in queue",
								Value:  fmt.Sprintf("%v", len(musicPlayer.Queue())),
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
							Color:  0x0345fc,
							Fields: fields,
						}
						sendEmbedResponse(s, i, embed)
					}
				}

				go musicPlayer.Play()
			case "stop":
				musicPlayer.Stop()

				sendMessageResponse(s, i, "Stopped playing")
			case "resume":
				musicPlayer.Resume()

				sendMessageResponse(s, i, "Resumed playing")
			case "queue":
				guild, _ := s.Guild(i.GuildID)

				embed := &discordgo.MessageEmbed{
					Title:       fmt.Sprintf("Queue for %s", guild.Name),
					Description: getQueueListString(musicPlayer),
					Color:       0x0345fc,
				}
				sendEmbedResponse(s, i, embed)
			case "shuffle":
				musicPlayer.Shuffle()

				sendMessageResponse(s, i, "Queue shuffled")
			case "skip":
				musicPlayer.Skip()

				sendMessageResponse(s, i, "Skipped currently playing audio")
			case "clear":
				musicPlayer.ClearQueue()

				sendMessageResponse(s, i, "Cleared queue")
			case "nowplaying":
				np := musicPlayer.Queue()[0]

				seglength := np.Duration / 30
				elapsed := musicPlayer.TrackPosition()
				seekPosition := int(math.Round(elapsed.Seconds() / seglength.Seconds()))

				var b strings.Builder
				b.WriteString("`")
				for i := 0; i <= 30; i++ {
					if i == seekPosition {
						b.WriteString("🔘")
					} else {
						b.WriteString("▬")
					}

				}
				b.WriteString("`")

				fmt.Fprintf(&b, "\n\n%s / %s", getDurationString(elapsed), getDurationString(np.Duration))
				fmt.Fprintf(&b, "\n\n`Requested by:` %s", np.RequestedBy)

				embed := &discordgo.MessageEmbed{
					Author: &discordgo.MessageEmbedAuthor{
						Name: "Now Playing 🎵",
					},
					Title: np.Title,
					URL:   np.Url,
					Thumbnail: &discordgo.MessageEmbedThumbnail{
						URL: np.ThumbnailURL,
					},
					Color:       0x0345fc,
					Description: b.String(),
				}
				sendEmbedResponse(s, i, embed)
			case "loop":
				musicPlayer.LoopQueue()

				loopState := ""
				if musicPlayer.LoopQueueState() {
					loopState = "enabled"
				} else {
					loopState = "disabled"
				}
				sendMessageResponse(s, i, fmt.Sprintf("Queue looping **%s** 🔁", loopState))
			case "removeduplicates":
				musicPlayer.RemoveDuplicates()

				sendMessageResponse(s, i, "Duplicates removed")
			case "disconnect":
				if err := disconnectPlayer(s, i.GuildID); err != nil {
					log.Printf("Error when disconnecting: %v", err)
					sendMessageResponse(s, i, "Unable to disconnect player. Try disconnecting manually.")
				} else {
					sendMessageResponse(s, i, "See ya!")
				}
			}
		},
	},
}

func getMusicPlayerInstance(s *discordgo.Session, i *discordgo.InteractionCreate) (*musicplayer.MusicPlayer, error) {
	voiceState, _ := s.State.VoiceState(i.GuildID, i.Member.User.ID)
	if voiceState == nil {
		return nil, fmt.Errorf("You must be in a voice channel to use this command")
	}

	if i.ApplicationCommandData().Options[0].Name == "play" && guildMusicPlayers[i.GuildID] != nil && guildMusicPlayers[i.GuildID].GetChannelID() != voiceState.ChannelID {
		guildMusicPlayers[i.GuildID].Close()
		delete(guildMusicPlayers, i.GuildID)
	}

	if guildMusicPlayers[i.GuildID] == nil {
		if i.ApplicationCommandData().Options[0].Name == "play" {
			var err error
			guildMusicPlayers[i.GuildID], err = musicplayer.New(s, voiceState)
			if err != nil {
				log.Printf("[%s] Unable to join voice channel: %v", i.GuildID, err)
				return nil, fmt.Errorf("Unable to join voice channel")
			}
		} else {
			return nil, fmt.Errorf("Not running in a channel! Try playing something first.")
		}
	}

	return guildMusicPlayers[i.GuildID], nil
}

func disconnectPlayer(s *discordgo.Session, guildID string) error {
	if guildMusicPlayers[guildID] != nil {
		if err := guildMusicPlayers[guildID].Close(); err != nil {
			return fmt.Errorf("Error destroying player for %s: %s", guildID, err)
		}
		delete(guildMusicPlayers, guildID)
	}

	if err := s.ChannelVoiceJoinManual(guildID, "", false, true); err != nil {
		return fmt.Errorf("Failed to leave voice channel for %s: %s", guildID, err)
	}

	return nil
}

func sendMessageResponse(s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: message,
		},
	})
	if err != nil {
		log.Printf("Failed to send response: %v", err)
	}
}

func sendEmbedResponse(s *discordgo.Session, i *discordgo.InteractionCreate, message *discordgo.MessageEmbed) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				message,
			},
		},
	})
	if err != nil {
		log.Printf("Failed to send response: %v", err)
	}
}

func getDurationString(d time.Duration) string {
	d = d.Round(time.Second)
	d -= (d / time.Hour) * time.Hour
	hours := d / time.Hour
	d -= hours * time.Hour
	mins := d / time.Minute
	d -= mins * time.Minute
	secs := d / time.Second

	var b strings.Builder
	if hours > 0 {
		fmt.Fprintf(&b, "%d:%02d", hours, mins)
	} else {
		fmt.Fprintf(&b, "%d", mins)
	}
	fmt.Fprintf(&b, ":%02d", secs)
	return b.String()
}

func getQueueListString(p *musicplayer.MusicPlayer) string {
	var b strings.Builder
	active := p.NowPlaying()
	queue := p.Queue()

	if len(queue) == 0 {
		return "Queue empty! Add some music!"
	}

	b.WriteString("__Now Playing:__\n")
	fmt.Fprintf(&b, getQueueTrackString(active))

	if len(queue) > 1 {
		b.WriteString("\n")
		b.WriteString("__Up Next:__\n")

		cap := 10
		if len(queue) < cap {
			cap = len(queue)
		}

		for i, item := range queue[1:cap] {
			fmt.Fprintf(&b, "`%d.` %s", i+1, getQueueTrackString(item))
		}
	}

	loopQueueState := ""
	if p.LoopQueueState() {
		loopQueueState = " | 🔁 Enabled"
	}

	queueDurationString := getDurationString(p.GetTotalQueueTime())
	fmt.Fprintf(&b, "**%d songs in queue | %s total length%s**", len(queue), queueDurationString, loopQueueState)

	return b.String()
}

func getQueueTrackString(item *musicplayer.PlayerQueueItem) string {
	return fmt.Sprintf("[%s](%s) | `%s` `Requested by: %s`\n\n", item.Title, item.Url, getDurationString(item.Duration), item.RequestedBy)
}
