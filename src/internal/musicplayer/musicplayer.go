package musicplayer

import (
	"context"
	"ramiel/internal/discordutils"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/disgoorg/disgolink/lavalink"
)

type MusicPlayer struct {
	discordSession  *discordgo.Session
	lavalinkManager *LavalinkManager
	PlayerManagers  map[string]*PlayerManager
}

func NewMusicPlayer(session *discordgo.Session) *MusicPlayer {
	return &MusicPlayer{
		discordSession:  session,
		lavalinkManager: NewLavalinkManager("lavalink", "2333", "youshallnotpass", session),
		PlayerManagers:  map[string]*PlayerManager{},
	}
}

func (p *MusicPlayer) PlayerExists(guildID string) bool {
	_, exists := p.PlayerManagers[guildID]
	return exists
}

func (p *MusicPlayer) PlayQuery(guildID string, voiceChannelID string, channelID string, requestedBy string, query string) {
	p.lavalinkManager.Link.BestRestClient().LoadItemHandler(context.TODO(), query, lavalink.NewResultHandler(
		func(track lavalink.AudioTrack) {
			track.SetUserData(requestedBy)
			p.play(guildID, voiceChannelID, channelID, track)
		},
		func(playlist lavalink.AudioPlaylist) {
			for _, t := range playlist.Tracks() {
				t.SetUserData(requestedBy)
			}

			p.play(guildID, voiceChannelID, channelID, playlist.Tracks()...)
		},
		func(tracks []lavalink.AudioTrack) {
			for _, t := range tracks {
				t.SetUserData(requestedBy)
			}

			p.play(guildID, voiceChannelID, channelID, tracks[0])
		},
		func() {
			_, _ = p.discordSession.ChannelMessageSend(channelID, "no matches found for: "+query)
		},
		func(ex lavalink.FriendlyException) {
			_, _ = p.discordSession.ChannelMessageSend(channelID, "error while loading track: "+ex.Message)
		},
	))
}

func (p *MusicPlayer) play(guildID string, voiceChannelID string, channelID string, tracks ...lavalink.AudioTrack) {
	if err := p.discordSession.ChannelVoiceJoinManual(guildID, voiceChannelID, false, false); err != nil {
		_, _ = p.discordSession.ChannelMessageSend(channelID, "error while joining voice channel: "+err.Error())
		return
	}

	manager, ok := p.PlayerManagers[guildID]
	if !ok {
		manager = NewPlayerManager(p.lavalinkManager.Link, guildID)
		p.PlayerManagers[guildID] = manager
	}

	manager.AddQueue(tracks...)

	if manager.Player.PlayingTrack() != nil {
		return
	}

	p.nextTrack(guildID, channelID)
}

func (p *MusicPlayer) nextTrack(guildID string, channelID string) {
	manager, ok := p.PlayerManagers[guildID]
	if !ok {
		manager = NewPlayerManager(p.lavalinkManager.Link, guildID)
		p.PlayerManagers[guildID] = manager
	}

	if len(manager.Queue) == 0 {
		if err := manager.Player.Stop(); err != nil {
			_, _ = p.discordSession.ChannelMessageSend(channelID, "error while stopping track: "+err.Error())
		}
		return
	}

	track := manager.PopQueue()

	if err := manager.Player.Play(track); err != nil {
		_, _ = p.discordSession.ChannelMessageSend(channelID, "error while playing track: "+err.Error())
		return
	}
}

func (p *MusicPlayer) Stop(guildID string) {
	manager, ok := p.PlayerManagers[guildID]
	if !ok {
		return
	}

	manager.Player.Pause(true)
}

func (p *MusicPlayer) Resume(guildID string) {
	manager, ok := p.PlayerManagers[guildID]
	if !ok {
		return
	}

	manager.Player.Pause(false)
}

func (p *MusicPlayer) TrackPosition(guildID string) time.Duration {
	manager, ok := p.PlayerManagers[guildID]
	if !ok {
		return 0
	}

	return time.Duration(manager.Player.Position()) * time.Millisecond
}

func (p *MusicPlayer) ToggleLoopingState(guildID string) {
	manager, ok := p.PlayerManagers[guildID]
	if !ok {
		return
	}

	if manager.RepeatingMode == RepeatingModeOff {
		manager.RepeatingMode = RepeatingModeSong
	} else {
		manager.RepeatingMode = RepeatingModeOff
	}
}

func (p *MusicPlayer) QueueLoopState(guildID string) RepeatingMode {
	manager, ok := p.PlayerManagers[guildID]
	if !ok {
		return -1
	}

	return manager.RepeatingMode
}

func (p *MusicPlayer) Shuffle(guildID string) {
	manager, ok := p.PlayerManagers[guildID]
	if !ok {
		return
	}

	manager.Shuffle()
}

func (p *MusicPlayer) ClearQueue(guildID string) {
	manager, ok := p.PlayerManagers[guildID]
	if !ok {
		return
	}

	manager.Clear()
}

func (p *MusicPlayer) RemoveDuplicates(guildID string) {
	manager, ok := p.PlayerManagers[guildID]
	if !ok {
		return
	}

	manager.RemoveDuplicates()
}

func (p *MusicPlayer) Skip(guildID string) {
	manager, ok := p.PlayerManagers[guildID]
	if !ok {
		return
	}

	if manager.RepeatingMode == RepeatingModeSong {
		manager.RepeatingMode = RepeatingModeOff
	}

	p.nextTrack(guildID, manager.Player.ChannelID().String())
}

func (p *MusicPlayer) NowPlaying(guildID string) lavalink.AudioTrack {
	manager, ok := p.PlayerManagers[guildID]
	if !ok {
		return nil
	}

	return manager.Player.PlayingTrack()
}

func (p *MusicPlayer) GetQueue(guildID string) []lavalink.AudioTrack {
	manager, ok := p.PlayerManagers[guildID]
	if !ok {
		return nil
	}

	return manager.Queue
}

func (p *MusicPlayer) GetTotalQueueTime(guildID string) time.Duration {
	manager, ok := p.PlayerManagers[guildID]
	if !ok {
		return 0
	}

	return manager.QueueDuration()
}

func (p *MusicPlayer) Disconnect(guildID string) {
	manager, ok := p.PlayerManagers[guildID]
	if !ok {
		return
	}

	delete(p.PlayerManagers, guildID)
	manager.Destroy()
	_ = discordutils.LeaveVoiceChannel(p.discordSession, guildID)
}

func (p *MusicPlayer) Destroy() {
	p.lavalinkManager.Destroy()
}
