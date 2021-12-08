package musicplayer

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"ramiel/internal/youtube"

	"github.com/bwmarrin/discordgo"
)

type MusicPlayer struct {
	session    *discordgo.Session
	lavalink   *LavalinkManager
	channelID  string
	queue      []*PlayerQueueItem
	activeSong *PlayerQueueItem
	isPlaying  bool
	loopQueue  bool
	skip       chan bool
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
		queue:     make([]*PlayerQueueItem, 0),
		loopQueue: false,
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

	p.queue = append(p.queue, playlistInfo.Items...)

	return playlistInfo, nil
}

func (p *MusicPlayer) AddSongToQueue(member *discordgo.Member, url string) (*PlayerQueueItem, error) {
	video, err := youtube.ResolveVideoData(url)
	if err != nil {
		return nil, err
	}

	requestedBy := fmt.Sprintf("%s#%s", member.User.Username, member.User.Discriminator)
	queueItem := newPlayerQueueItem(requestedBy, video)

	p.queue = append(p.queue, queueItem)

	return queueItem, nil
}

func (p *MusicPlayer) Play() error {
	if p.isPlaying {
		return nil
	}

	p.isPlaying = true

	var err error
	for len(p.queue) > 0 && p.isPlaying {
		err = p.playCurrentSong()
		if err != nil {
			break
		}
		if p.activeSong != nil {
			p.postSongHandling(p.activeSong)
		}
	}

	p.isPlaying = false

	return err
}

func (p *MusicPlayer) Stop() {
	p.lavalink.Player.Pause(true)
}

func (p *MusicPlayer) Resume() {
	p.lavalink.Player.Pause(false)
}

func (p *MusicPlayer) LoopQueue() {
	p.loopQueue = !p.loopQueue
}

func (p *MusicPlayer) LoopQueueState() bool {
	return p.loopQueue
}

func (p *MusicPlayer) TrackPosition() time.Duration {
	return time.Duration(p.lavalink.Player.Position()) * time.Millisecond
}

func (p *MusicPlayer) Queue() []*PlayerQueueItem {
	return p.queue
}

func (p *MusicPlayer) Shuffle() {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(p.queue), func(i, j int) { p.queue[i], p.queue[j] = p.queue[j], p.queue[i] })
}

func (p *MusicPlayer) ClearQueue() {
	for j := 1; j < len(p.queue); j++ {
		p.queue[j] = nil
	}
	p.queue = p.queue[:1]
}

func (p *MusicPlayer) RemoveDuplicates() {
	keys := make(map[string]bool)
	list := make([]*PlayerQueueItem, 0)
	for _, entry := range p.queue {
		if _, value := keys[entry.VideoID]; !value {
			keys[entry.VideoID] = true
			list = append(list, entry)
		}
	}
	p.queue = list
}

func (p *MusicPlayer) NowPlaying() *PlayerQueueItem {
	return p.activeSong
}

func (p *MusicPlayer) Skip() {
	p.skip <- true
}

func (p *MusicPlayer) RemoveSongFromQueue(item *PlayerQueueItem) {
	if len(p.queue) == 0 {
		return
	}

	sIdx := p.findSongIndex(item.VideoID)
	p.queue = append(p.queue[:sIdx], p.queue[sIdx+1:]...)
}

func (p *MusicPlayer) GetTotalQueueTime() time.Duration {
	var sum time.Duration = 0
	for _, v := range p.Queue() {
		sum += v.Duration
	}
	return sum
}

func (p *MusicPlayer) Close() error {
	return p.lavalink.Close()
}

func (p *MusicPlayer) playCurrentSong() error {
	if p.queue[0].Video == nil {
		video, err := youtube.ResolveVideoData(p.queue[0].Url)
		if err != nil {
			return err
		}

		p.queue[0] = newPlayerQueueItem(p.queue[0].RequestedBy, video)
	}

	p.activeSong = p.queue[0]

	log.Printf("Playing %s", p.activeSong.Title)

	err := p.lavalink.Play(p.activeSong.Url)
	if err != nil {
		return err
	}

	log.Println("Play started")

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

func (p *MusicPlayer) postSongHandling(item *PlayerQueueItem) {
	p.activeSong = nil

	if p.loopQueue {
		sIdx := p.findSongIndex(item.VideoID)
		p.queue = append(p.queue[:sIdx], p.queue[sIdx+1:]...)
		p.queue = append(p.queue, item)
	} else {
		p.RemoveSongFromQueue(item)
	}
}

func (p *MusicPlayer) findSongIndex(videoID string) int {
	for i := range p.queue {
		if p.queue[i].VideoID == videoID {
			return i
		}
	}
	return -1
}
