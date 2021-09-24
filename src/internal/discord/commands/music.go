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
				musicPlayer[i.GuildID].Close()
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
							Value:  getDurationString(musicPlayer[i.GuildID].GetTotalQueueTime() - item.Duration),
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
						Color:  0x0345fc,
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
				return
			}

			musicPlayer[i.GuildID].Stop()

			sendMessageResponse(s, i, "Stopped playing")
		},
		"resume": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if musicPlayer[i.GuildID] == nil {
				sendMessageResponse(s, i, "Not running in a channel!")
				return
			}

			musicPlayer[i.GuildID].Resume()

			sendMessageResponse(s, i, "Resumed playing")
		},
		"queue": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if musicPlayer[i.GuildID] == nil {
				sendMessageResponse(s, i, "Not running in a channel!")
				return
			}

			guild, _ := s.Guild(i.GuildID)

			embed := &discordgo.MessageEmbed{
				Title:       fmt.Sprintf("Queue for %s", guild.Name),
				Description: getQueueListString(musicPlayer[i.GuildID]),
				Color:       0x0345fc,
			}
			sendEmbedResponse(s, i, embed)
		},
		"shuffle": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if musicPlayer[i.GuildID] == nil {
				sendMessageResponse(s, i, "Not running in a channel!")
				return
			}

			musicPlayer[i.GuildID].Shuffle()

			sendMessageResponse(s, i, "Queue shuffled")
		},
		"skip": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if musicPlayer[i.GuildID] == nil {
				sendMessageResponse(s, i, "Not running in a channel!")
				return
			}

			musicPlayer[i.GuildID].Skip()

			sendMessageResponse(s, i, "Skipped currently playing audio")
		},
		"clear": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if musicPlayer[i.GuildID] == nil {
				sendMessageResponse(s, i, "Not running in a channel!")
				return
			}

			musicPlayer[i.GuildID].ClearQueue()

			sendMessageResponse(s, i, "Cleared queue")
		},
		"nowplaying": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if musicPlayer[i.GuildID] == nil {
				sendMessageResponse(s, i, "Not running in a channel!")
				return
			}

			np := musicPlayer[i.GuildID].Queue()[0]

			seglength := np.Duration / 30
			elapsed := musicPlayer[i.GuildID].TrackPosition()
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
		},
		"disconnect": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if musicPlayer[i.GuildID] == nil {
				sendMessageResponse(s, i, "Not running in a channel!")
				return
			}

			err := musicPlayer[i.GuildID].Close()
			if err != nil {
				log.Printf("Error when closing music player: %v", err)
			}

			delete(musicPlayer, i.GuildID)
		},
	},
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
	b.WriteString("\n")
	b.WriteString("__Up Next:__\n")

	if len(queue) > 1 {
		cap := 10
		if len(queue) < cap {
			cap = len(queue)
		}

		for i, item := range queue[1:cap] {
			fmt.Fprintf(&b, "`%d.` %s", i+1, getQueueTrackString(item))
		}
	}

	queueDurationString := getDurationString(p.GetTotalQueueTime())
	fmt.Fprintf(&b, "**%d songs in queue | %s total length**", len(queue), queueDurationString)

	return b.String()
}

func getQueueTrackString(item *musicplayer.PlayerQueueItem) string {
	return fmt.Sprintf("[%s](%s) | `%s Requested by: %s`\n\n", item.Title, item.Url, getDurationString(item.Duration), item.RequestedBy)
}
