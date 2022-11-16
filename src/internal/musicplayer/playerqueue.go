package musicplayer

import (
	"math/rand"
	"sync"
	"time"

	"github.com/disgoorg/disgolink/lavalink"
)

type PlayerQueue struct {
	queue   []lavalink.AudioTrack
	queueMu sync.Mutex
}

func NewPlayerQueue() *PlayerQueue {
	return &PlayerQueue{
		queue: make([]lavalink.AudioTrack, 0),
	}
}

func (q *PlayerQueue) Get() []lavalink.AudioTrack {
	q.queueMu.Lock()
	defer q.queueMu.Unlock()

	return q.queue
}

func (q *PlayerQueue) AddTracks(tracks ...lavalink.AudioTrack) {
	q.queueMu.Lock()
	defer q.queueMu.Unlock()

	q.queue = append(q.queue, tracks...)
}

func (q *PlayerQueue) PopTrack() lavalink.AudioTrack {
	q.queueMu.Lock()
	defer q.queueMu.Unlock()

	if len(q.queue) == 0 {
		return nil
	}

	var track lavalink.AudioTrack
	track, q.queue = q.queue[0], q.queue[1:]

	return track
}

func (q *PlayerQueue) Duration() time.Duration {
	var sum time.Duration = 0

	for _, v := range q.queue {
		sum += time.Duration(v.Info().Length) * time.Millisecond
	}

	return sum
}

func (q *PlayerQueue) Shuffle() {
	q.queueMu.Lock()
	defer q.queueMu.Unlock()

	if len(q.queue) < 2 {
		return
	}

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(q.queue), func(i, j int) { q.queue[i], q.queue[j] = q.queue[j], q.queue[i] })
}

func (q *PlayerQueue) RemoveDuplicates() {
	q.queueMu.Lock()
	defer q.queueMu.Unlock()

	keys := make(map[string]bool)
	list := make([]lavalink.AudioTrack, 0)
	for _, entry := range q.queue {
		if _, value := keys[entry.Info().Identifier]; !value {
			keys[entry.Info().Identifier] = true
			list = append(list, entry)
		}
	}
	q.queue = list
}

func (q *PlayerQueue) Clear() {
	q.queueMu.Lock()
	defer q.queueMu.Unlock()

	q.queue = []lavalink.AudioTrack{}
}
