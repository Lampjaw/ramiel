package discord

import (
	"errors"
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	"ramiel/internal/discordutils"
	"ramiel/internal/musicplayer"

	"github.com/bwmarrin/discordgo"
	"github.com/disgoorg/disgolink/lavalink"
)

var musicPlayer *musicplayer.MusicPlayer

var MusicCommands = &CommandDefinition{
	BotCommandInitializer: func(s *discordgo.Session) {
		musicPlayer = musicplayer.NewMusicPlayer(s)
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

			defer func() {
				if r := recover(); r != nil {
					sendMessageResponse(s, i, "r.(error)")
				}
			}()

			switch i.ApplicationCommandData().Options[0].Name {
			case "play":
				playerChannelID := discordutils.GetInteractionUserVoiceChannelID(s, i.Interaction)
				query := i.ApplicationCommandData().Options[0].Options[0].StringValue()

				requestedBy := fmt.Sprintf("%s#%s", i.Member.User.Username, i.Member.User.Discriminator)
				musicPlayer.PlayQuery(i.GuildID, playerChannelID, i.ChannelID, requestedBy, query)

			case "stop":
				pm := musicPlayer.GetPlayerManager(i.GuildID)
				pm.Player.Pause(true)

				sendMessageResponse(s, i, "Stopped playback")

			case "resume":
				pm := musicPlayer.GetPlayerManager(i.GuildID)
				pm.Player.Pause(false)

				sendMessageResponse(s, i, "Resuming playback")

			case "queue":
				pm := musicPlayer.GetPlayerManager(i.GuildID)

				displayQueue(s, i, pm)

			case "shuffle":
				pm := musicPlayer.GetPlayerManager(i.GuildID)
				pm.Queue.Shuffle()

				sendMessageResponse(s, i, "Queue shuffled")

			case "skip":
				pm := musicPlayer.GetPlayerManager(i.GuildID)
				pm.SkipTrack()

				sendMessageResponse(s, i, "Skipped currently playing audio")

			case "clear":
				pm := musicPlayer.GetPlayerManager(i.GuildID)
				pm.Queue.Clear()

				sendMessageResponse(s, i, "Cleared queue")

			case "nowplaying":
				pm := musicPlayer.GetPlayerManager(i.GuildID)

				displayNowPlaying(s, i, pm)

			case "loop":
				pm := musicPlayer.GetPlayerManager(i.GuildID)
				pm.ToggleLoopingState()

			case "removeduplicates":
				pm := musicPlayer.GetPlayerManager(i.GuildID)
				pm.Queue.RemoveDuplicates()

				sendMessageResponse(s, i, "Removed duplicate tracks")

			case "disconnect":
				_ = musicPlayer.GetPlayerManager(i.GuildID)

				musicPlayer.Disconnect(i.GuildID)

				sendMessageResponse(s, i, "See ya!")
			}
		},
	},
}

func verifyMusicPlayer(p *musicplayer.MusicPlayer, guildID string) {
	if !p.PlayerExists(guildID) {
		panic(errors.New("Not running in a channel! Try playing something first."))
	}
}

func displayQueue(s *discordgo.Session, i *discordgo.InteractionCreate, m *musicplayer.PlayerManager) {
	guild, _ := s.Guild(i.GuildID)

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("Queue for %s", guild.Name),
		Description: getQueueListString(m, i.GuildID),
		Color:       0x0345fc,
	}

	sendEmbedResponse(s, i, embed)
}

func displayNowPlaying(s *discordgo.Session, i *discordgo.InteractionCreate, m *musicplayer.PlayerManager) {
	np := m.Player.PlayingTrack()
	npLength := time.Duration(np.Info().Length) * time.Millisecond

	seglength := npLength / 30
	elapsed := m.TrackPosition()
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

	fmt.Fprintf(&b, "\n\n%s / %s", getDurationString(elapsed), getDurationString(npLength))
	fmt.Fprintf(&b, "\n\n`Requested by:` %s", np.UserData())

	embed := &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name: "Now Playing 🎵",
		},
		Title: np.Info().Title,
		URL:   *np.Info().URI,
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: getYoutubeThumbnailUrl(np),
		},
		Color:       0x0345fc,
		Description: b.String(),
	}

	sendEmbedResponse(s, i, embed)
}

func sendMessageResponse(s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
	if err := discordutils.SendMessageResponse(s, i.Interaction, message); err != nil {
		log.Println(err)
	}
}

func sendEmbedResponse(s *discordgo.Session, i *discordgo.InteractionCreate, message *discordgo.MessageEmbed) {
	if err := discordutils.SendEmbedResponse(s, i.Interaction, message); err != nil {
		log.Println(err)
	}
}

func getQueueTrackString(item lavalink.AudioTrack) string {
	sDuration := getDurationString(time.Duration(item.Info().Length) * time.Millisecond)
	return fmt.Sprintf("[%s](%s) | `%s` `Requested by: %s`\n\n", item.Info().Title, *item.Info().URI, sDuration, item.UserData())
}

func getQueueListString(m *musicplayer.PlayerManager, guildID string) string {
	var b strings.Builder
	active := m.Player.PlayingTrack()
	queue := m.Queue.Get()

	if len(queue) == 0 && active == nil {
		return "Queue empty! Add some music!"
	}

	b.WriteString("__Now Playing:__\n")
	fmt.Fprintf(&b, getQueueTrackString(active))

	if len(queue) > 0 {
		b.WriteString("\n")
		b.WriteString("__Up Next:__\n")

		cap := 10
		if len(queue) < cap {
			cap = len(queue)
		}

		for i, item := range queue[:cap] {
			fmt.Fprintf(&b, "`%d.` %s", i+1, getQueueTrackString(item))
		}
	}

	loopQueueState := ""
	if m.RepeatingMode == musicplayer.RepeatingModeQueue {
		loopQueueState = " | 🔁 Enabled"
	}

	queueDurationString := getDurationString(m.Queue.Duration())
	fmt.Fprintf(&b, "**%d songs in queue | %s total length%s**", len(queue), queueDurationString, loopQueueState)

	return b.String()
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

func getYoutubeThumbnailUrl(track lavalink.AudioTrack) string {
	return fmt.Sprintf("https://i.ytimg.com/vi/%s/hqdefault.jpg", track.Info().Identifier)
}
