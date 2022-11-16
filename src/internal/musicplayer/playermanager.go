package musicplayer

import (
	"fmt"
	"sync"
	"time"

	"github.com/disgoorg/disgolink/dgolink"
	"github.com/disgoorg/disgolink/lavalink"
	"github.com/disgoorg/snowflake/v2"
)

type PlayerManager struct {
	lavalink.PlayerEventAdapter
	Player        lavalink.Player
	Queue         *PlayerQueue
	PlayerMu      sync.Mutex
	RepeatingMode RepeatingMode
}

func NewPlayerManager(link *dgolink.Link, guildID string) *PlayerManager {
	manager := &PlayerManager{
		Player:        link.Player(snowflake.MustParse(guildID)),
		Queue:         NewPlayerQueue(),
		RepeatingMode: RepeatingModeOff,
	}

	manager.Player.AddListener(manager)
	manager.Player.SetVolume(30)

	return manager
}

func (m *PlayerManager) Destroy() {
	m.Player.Destroy()
}

func (m *PlayerManager) Play() error {
	if m.Player.PlayingTrack() == nil {
		m.playNextTrack()
	}

	return nil
}

func (m *PlayerManager) SkipTrack() error {
	if m.RepeatingMode == RepeatingModeSong {
		m.RepeatingMode = RepeatingModeOff
	}

	return m.playNextTrack()
}

func (m *PlayerManager) ToggleLoopingState() {
	if m.RepeatingMode == RepeatingModeOff {
		m.RepeatingMode = RepeatingModeSong
	} else {
		m.RepeatingMode = RepeatingModeOff
	}
}

func (m *PlayerManager) TrackPosition() time.Duration {
	return time.Duration(m.Player.Position()) * time.Millisecond
}

func (m *PlayerManager) OnTrackException(player lavalink.Player, track lavalink.AudioTrack, exception lavalink.FriendlyException) {
	fmt.Println("error playing track:", exception.Message)
}

func (m *PlayerManager) OnTrackStuck(player lavalink.Player, track lavalink.AudioTrack) {
	fmt.Println("track stuck:", track.Info().Title)
}

func (m *PlayerManager) OnTrackEnd(player lavalink.Player, track lavalink.AudioTrack, endReason lavalink.AudioTrackEndReason) {
	if !endReason.MayStartNext() {
		return
	}

	switch m.RepeatingMode {
	case RepeatingModeOff:
		m.playNextTrack()
	case RepeatingModeSong:
		if err := player.Play(track.Clone()); err != nil {
			fmt.Println("error playing next track:", err)
		}

	case RepeatingModeQueue:
		m.Queue.AddTracks(track)
		if nextTrack := m.Queue.PopTrack(); nextTrack != nil {
			if err := player.Play(nextTrack); err != nil {
				fmt.Println("error playing next track:", err)
			}
		}
	}
}

func (m *PlayerManager) playNextTrack() error {
	if nextTrack := m.Queue.PopTrack(); nextTrack != nil {
		if err := m.Player.Play(nextTrack); err != nil {
			return fmt.Errorf("error while playing track: %s", err.Error())
		}
	}

	return nil
}
