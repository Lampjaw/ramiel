package musicplayer

import (
	"math/rand"
	"time"
)

type MusicPlayerQueue struct {
	queue     []*PlayerQueueItem
	loopState LoopState
}

func NewMusicPlayerQueue() *MusicPlayerQueue {
	return &MusicPlayerQueue{
		queue:     make([]*PlayerQueueItem, 0),
		loopState: LoopingDisabled,
	}
}

func (q *MusicPlayerQueue) ActiveItem() *PlayerQueueItem {
	if len(q.queue) > 0 {
		return q.queue[0]
	}
	return nil
}

func (q *MusicPlayerQueue) LoopState() LoopState {
	return q.loopState
}

func (q *MusicPlayerQueue) Queue() []*PlayerQueueItem {
	return q.queue
}

func (q *MusicPlayerQueue) Length() int {
	return len(q.queue)
}

func (q *MusicPlayerQueue) ToggleLoopingState(s LoopState) bool {
	if q.loopState == s {
		q.loopState = LoopingDisabled
		return false
	} else {
		q.loopState = s
		return true
	}
}

func (p *MusicPlayerQueue) RemoveDuplicates() {
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

func (p *MusicPlayerQueue) Clear() {
	for j := 1; j < len(p.queue); j++ {
		p.queue[j] = nil
	}
	p.queue = p.queue[:1]
}

func (p *MusicPlayerQueue) Shuffle() {
	if len(p.queue) < 3 {
		return
	}
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(p.queue)-1, func(i, j int) { p.queue[i+1], p.queue[j+1] = p.queue[j+1], p.queue[i+1] })
}

func (q *MusicPlayerQueue) NextItem() *PlayerQueueItem {
	switch q.loopState {
	case QueueLooping:
		q.queue = append(q.queue[1:], q.queue[0])
	case LoopingDisabled:
		q.queue = q.queue[1:]
	}

	return q.ActiveItem()
}

func (q *MusicPlayerQueue) QueueDuration() time.Duration {
	var sum time.Duration = 0
	for _, v := range q.queue {
		sum += v.Duration
	}
	return sum
}

func (q *MusicPlayerQueue) AddItem(item *PlayerQueueItem) {
	q.queue = append(q.queue, item)
}

func (q *MusicPlayerQueue) AddItems(items []*PlayerQueueItem) {
	q.queue = append(q.queue, items...)
}
