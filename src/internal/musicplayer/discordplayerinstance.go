package musicplayer

import (
	"fmt"
	"log"
	"math"
	"ramiel/internal/discordutils"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

type DiscordPlayerInstance struct {
	session             *discordgo.Session
	musicplayer         *MusicPlayer
	sourceTextChannelID string
	guildID             string
}

func NewDiscordPlayerInstance(s *discordgo.Session, guildID string, voiceChannelID string, sourceTextChannelID string) (*DiscordPlayerInstance, error) {
	instance := &DiscordPlayerInstance{
		session:             s,
		guildID:             guildID,
		sourceTextChannelID: sourceTextChannelID,
	}

	if err := instance.connectMusicPlayer(voiceChannelID); err != nil {
		return nil, err
	}

	return instance, nil
}

func (p *DiscordPlayerInstance) SetPlayerChannel(channelID string) error {
	if p.session.VoiceConnections[p.guildID] != nil && p.session.VoiceConnections[p.guildID].ChannelID != channelID {
		if err := p.destroyMusicPlayer(); err != nil {
			return err
		}

		if err := p.connectMusicPlayer(channelID); err != nil {
			return err
		}
	}

	return nil
}

func (p *DiscordPlayerInstance) VoiceServerUpdate(s *discordgo.Session, event *discordgo.VoiceServerUpdate) error {
	return p.musicplayer.VoiceServerUpdate(s, event)
}

func (p *DiscordPlayerInstance) Destroy() error {
	return p.destroyMusicPlayer()
}

func (p *DiscordPlayerInstance) Play(sourceUrl string, interaction *discordgo.InteractionCreate) error {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("Panic occured: %v", err)
		}
	}()

	if sourceUrl != "" {
		if strings.Contains(sourceUrl, "playlist") {
			playlist, err := p.musicplayer.AddPlaylistToQueue(interaction.Member, sourceUrl)
			if err != nil {
				log.Printf("[%s] Failed to add playlist: %v", p.guildID, err)
				return fmt.Errorf("Failed to add playlist")
			}

			p.sendMessageResponse(interaction, fmt.Sprintf("Adding %v songs to the queue from `%s`", len(playlist.Items), playlist.Title))
		} else {
			item, err := p.musicplayer.AddSongToQueue(interaction.Member, sourceUrl)
			if err != nil {
				log.Printf("[%s] Failed to add track: %v", p.guildID, err)
				return fmt.Errorf("Failed to add track")
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
					Value:  getDurationString(p.musicplayer.GetTotalQueueTime() - item.Duration),
					Inline: true,
				},
				&discordgo.MessageEmbedField{
					Name:   "Position in queue",
					Value:  fmt.Sprintf("%v", len(p.musicplayer.Queue())),
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
			p.sendEmbedResponse(interaction, embed)
		}
	}

	go p.musicplayer.Play()

	return nil
}

func (p *DiscordPlayerInstance) Stop(interaction *discordgo.InteractionCreate) {
	p.musicplayer.Stop()

	p.sendMessageResponse(interaction, "Stopping playback")
}

func (p *DiscordPlayerInstance) Resume(interaction *discordgo.InteractionCreate) {
	p.musicplayer.Resume()

	p.sendMessageResponse(interaction, "Resuming playback")
}

func (p *DiscordPlayerInstance) Queue(interaction *discordgo.InteractionCreate) {
	guild, _ := p.session.Guild(p.guildID)

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("Queue for %s", guild.Name),
		Description: p.getQueueListString(),
		Color:       0x0345fc,
	}

	p.sendEmbedResponse(interaction, embed)
}

func (p *DiscordPlayerInstance) Shuffle(interaction *discordgo.InteractionCreate) {
	p.musicplayer.Shuffle()

	p.sendMessageResponse(interaction, "Queue shuffled")
}

func (p *DiscordPlayerInstance) Skip(interaction *discordgo.InteractionCreate) {
	p.musicplayer.Skip()

	p.sendMessageResponse(interaction, "Skipped currently playing audio")
}

func (p *DiscordPlayerInstance) Clear(interaction *discordgo.InteractionCreate) {
	p.musicplayer.ClearQueue()

	p.sendMessageResponse(interaction, "Cleared queue")
}

func (p *DiscordPlayerInstance) NowPlaying(interaction *discordgo.InteractionCreate) {
	np := p.musicplayer.NowPlaying()

	seglength := np.Duration / 30
	elapsed := p.musicplayer.TrackPosition()
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

	p.sendEmbedResponse(interaction, embed)
}

func (p *DiscordPlayerInstance) Loop(interaction *discordgo.InteractionCreate) {
	isEnabled := p.musicplayer.ToggleLoopingState(QueueLooping)

	loopState := ""
	if isEnabled {
		loopState = "enabled"
	} else {
		loopState = "disabled"
	}

	p.sendMessageResponse(interaction, fmt.Sprintf("Queue looping **%s** 🔁", loopState))
}

func (p *DiscordPlayerInstance) RemoveDuplicates(interaction *discordgo.InteractionCreate) {
	p.musicplayer.RemoveDuplicates()

	p.sendMessageResponse(interaction, "Duplicates removed")
}

func (p *DiscordPlayerInstance) InitPlayerErrorListener(player *MusicPlayer) {
	go func() {
		for {
			if player == nil {
				break
			}
			select {
			case err := <-player.Errors:
				p.sendMessage(p.sourceTextChannelID, err.Error())
			}
		}
	}()
}

func (p *DiscordPlayerInstance) getQueueListString() string {
	var b strings.Builder
	active := p.musicplayer.NowPlaying()
	queue := p.musicplayer.Queue()

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
	if p.musicplayer.QueueLoopState() == QueueLooping {
		loopQueueState = " | 🔁 Enabled"
	}

	queueDurationString := getDurationString(p.musicplayer.GetTotalQueueTime())
	fmt.Fprintf(&b, "**%d songs in queue | %s total length%s**", len(queue), queueDurationString, loopQueueState)

	return b.String()
}

func (p *DiscordPlayerInstance) connectMusicPlayer(channelID string) error {
	if err := p.session.ChannelVoiceJoinManual(p.guildID, channelID, false, true); err != nil {
		return err
	}

	musicplayer, err := NewMusicPlayer(p.session)
	if err != nil {
		return err
	}

	p.musicplayer = musicplayer

	p.InitPlayerErrorListener(p.musicplayer)

	return nil
}

func (p *DiscordPlayerInstance) destroyMusicPlayer() error {
	if err := p.musicplayer.Destroy(); err != nil {
		return err
	}

	if err := discordutils.LeaveVoiceChannel(p.session, p.guildID); err != nil {
		return err
	}

	p.musicplayer = nil
	return nil
}

func (p *DiscordPlayerInstance) sendMessageResponse(i *discordgo.InteractionCreate, message string) {
	if err := discordutils.SendMessageResponse(p.session, i.Interaction, message); err != nil {
		log.Println(err)
	}
}

func (p *DiscordPlayerInstance) sendEmbedResponse(i *discordgo.InteractionCreate, message *discordgo.MessageEmbed) {
	if err := discordutils.SendEmbedResponse(p.session, i.Interaction, message); err != nil {
		log.Println(err)
	}
}

func (p *DiscordPlayerInstance) sendMessage(channelID string, message string) {
	if err := discordutils.SendMessage(p.session, channelID, message); err != nil {
		log.Println(err)
	}
}

func getQueueTrackString(item *PlayerQueueItem) string {
	return fmt.Sprintf("[%s](%s) | `%s` `Requested by: %s`\n\n", item.Title, item.Url, getDurationString(item.Duration), item.RequestedBy)
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
