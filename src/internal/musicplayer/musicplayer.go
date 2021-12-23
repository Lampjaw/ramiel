package musicplayer

import (
	"fmt"
	"log"
	"time"

	"ramiel/internal/youtube"

	"github.com/bwmarrin/discordgo"
)

type MusicPlayer struct {
	session   *discordgo.Session
	lavalink  *LavalinkManager
	channelID string
	queue     *MusicPlayerQueue
	isPlaying bool
	skip      chan bool
}

func New(session *discordgo.Session, voiceState *discordgo.VoiceState) (*MusicPlayer, error) {
	lavalink, err := NewLavalinkManager("lavalink:2333", "youshallnotpass", session)
	if err != nil {
		return nil, err
	}

	err = session.ChannelVoiceJoinManual(voiceState.GuildID, voiceState.ChannelID, false, true)
	if err != nil {
		return nil, err
	}

	return &MusicPlayer{
		session:   session,
		lavalink:  lavalink,
		channelID: voiceState.ChannelID,
		isPlaying: false,
		queue:     NewMusicPlayerQueue(),
		skip:      make(chan bool),
	}, nil
}

func (p *MusicPlayer) VoiceServerUpdate(s *discordgo.Session, event *discordgo.VoiceServerUpdate) error {
	return p.lavalink.VoiceServerUpdate(s, event)
}

func (p *MusicPlayer) GetChannelID() string {
	return p.channelID
}

func (p *MusicPlayer) AddPlaylistToQueue(member *discordgo.Member, url string) (*PlaylistInfo, error) {
	playlist, err := youtube.ResolvePlaylistData(url)
	if err != nil {
		return nil, fmt.Errorf("Failed to get playlist info: %v", err)
	}

	if playlist == nil {
		return nil, nil
	}

	requestedBy := fmt.Sprintf("%s#%s", member.User.Username, member.User.Discriminator)
	playlistInfo := newPlaylistInfo(requestedBy, playlist)

	p.queue.AddItems(playlistInfo.Items)

	return playlistInfo, nil
}

func (p *MusicPlayer) AddSongToQueue(member *discordgo.Member, url string) (*PlayerQueueItem, error) {
	video, err := youtube.ResolveVideoData(url)
	if err != nil {
		return nil, err
	}

	requestedBy := fmt.Sprintf("%s#%s", member.User.Username, member.User.Discriminator)
	queueItem := newPlayerQueueItem(requestedBy, video)

	p.queue.AddItem(queueItem)

	return queueItem, nil
}

func (p *MusicPlayer) Play() {
	if p.isPlaying {
		return
	}

	p.isPlaying = true

	for {
		item := p.queue.ActiveItem()

		if err := p.playItem(item); err != nil {
			log.Printf("Failed to play item: %s: %s", item.Url, err)
		}

		if p.queue.NextItem() == nil {
			break
		}
	}

	p.isPlaying = false
}

func (p *MusicPlayer) Stop() {
	p.lavalink.Player.Pause(true)
}

func (p *MusicPlayer) Resume() {
	p.lavalink.Player.Pause(false)
}

func (p *MusicPlayer) ToggleLoopingState(s LoopState) bool {
	return p.queue.ToggleLoopingState(s)
}

func (p *MusicPlayer) QueueLoopState() LoopState {
	return p.queue.LoopState()
}

func (p *MusicPlayer) TrackPosition() time.Duration {
	return time.Duration(p.lavalink.Player.Position()) * time.Millisecond
}

func (p *MusicPlayer) Queue() []*PlayerQueueItem {
	return p.queue.Queue()
}

func (p *MusicPlayer) Shuffle() {
	p.queue.Shuffle()
}

func (p *MusicPlayer) ClearQueue() {
	p.queue.Clear()
}

func (p *MusicPlayer) RemoveDuplicates() {
	p.queue.RemoveDuplicates()
}

func (p *MusicPlayer) NowPlaying() *PlayerQueueItem {
	return p.queue.ActiveItem()
}

func (p *MusicPlayer) Skip() {
	if p.queue.LoopState() == ItemLooping {
		p.queue.ToggleLoopingState(LoopingDisabled)
	}

	p.skip <- true
}

func (p *MusicPlayer) GetTotalQueueTime() time.Duration {
	return p.queue.QueueDuration()
}

func (p *MusicPlayer) Close() error {
	return p.lavalink.Close()
}

func (p *MusicPlayer) playItem(item *PlayerQueueItem) error {
	if item.Video == nil {
		video, err := youtube.ResolveVideoData(item.Url)
		if err != nil {
			return err
		}

		item.Video = video
		item.ThumbnailURL = video.Thumbnails[0].URL
	}

	err := p.lavalink.Play(item.Url)
	if err != nil {
		return err
	}

	for {
		select {
		case <-p.lavalink.isTrackEnded:
			return nil
		case <-p.skip:
			p.lavalink.Player.Stop()
			return nil
		}
	}
}
