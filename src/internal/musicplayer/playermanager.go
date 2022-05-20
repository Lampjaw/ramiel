package musicplayer

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/disgoorg/disgolink/dgolink"
	"github.com/disgoorg/disgolink/lavalink"
	"github.com/disgoorg/snowflake/v2"
)

type PlayerManager struct {
	lavalink.PlayerEventAdapter
	Player        lavalink.Player
	Queue         []lavalink.AudioTrack
	QueueMu       sync.Mutex
	RepeatingMode RepeatingMode
}

func NewPlayerManager(link *dgolink.Link, guildID string) *PlayerManager {
	manager := &PlayerManager{
		Player:        link.Player(snowflake.MustParse(guildID)),
		RepeatingMode: RepeatingModeOff,
	}
	manager.Player.AddListener(manager)
	manager.Player.SetVolume(30)
	return manager
}

func (m *PlayerManager) AddQueue(tracks ...lavalink.AudioTrack) {
	m.QueueMu.Lock()
	defer m.QueueMu.Unlock()

	m.Queue = append(m.Queue, tracks...)
}

func (m *PlayerManager) PopQueue() lavalink.AudioTrack {
	m.QueueMu.Lock()
	defer m.QueueMu.Unlock()
	if len(m.Queue) == 0 {
		return nil
	}
	var track lavalink.AudioTrack
	track, m.Queue = m.Queue[0], m.Queue[1:]
	return track
}

func (m *PlayerManager) QueueDuration() time.Duration {
	var sum time.Duration = 0
	for _, v := range m.Queue {
		sum += time.Duration(v.Info().Length) * time.Millisecond
	}
	return sum
}

func (m *PlayerManager) Shuffle() {
	m.QueueMu.Lock()
	defer m.QueueMu.Unlock()

	if len(m.Queue) < 2 {
		return
	}

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(m.Queue), func(i, j int) { m.Queue[i], m.Queue[j] = m.Queue[j], m.Queue[i] })
}

func (m *PlayerManager) RemoveDuplicates() {
	m.QueueMu.Lock()
	defer m.QueueMu.Unlock()

	keys := make(map[string]bool)
	list := make([]lavalink.AudioTrack, 0)
	for _, entry := range m.Queue {
		if _, value := keys[entry.Info().Identifier]; !value {
			keys[entry.Info().Identifier] = true
			list = append(list, entry)
		}
	}
	m.Queue = list
}

func (m *PlayerManager) Clear() {
	m.QueueMu.Lock()
	defer m.QueueMu.Unlock()

	m.Queue = []lavalink.AudioTrack{}
}

func (m *PlayerManager) Destroy() {
	m.Player.Destroy()
}

func (m *PlayerManager) OnTrackEnd(player lavalink.Player, track lavalink.AudioTrack, endReason lavalink.AudioTrackEndReason) {
	if !endReason.MayStartNext() {
		return
	}

	switch m.RepeatingMode {
	case RepeatingModeOff:
		if nextTrack := m.PopQueue(); nextTrack != nil {
			if err := player.Play(nextTrack); err != nil {
				fmt.Println("error playing next track:", err)
			}
		}
	case RepeatingModeSong:
		if err := player.Play(track.Clone()); err != nil {
			fmt.Println("error playing next track:", err)
		}

	case RepeatingModeQueue:
		m.AddQueue(track)
		if nextTrack := m.PopQueue(); nextTrack != nil {
			if err := player.Play(nextTrack); err != nil {
				fmt.Println("error playing next track:", err)
			}
		}
	}
}
