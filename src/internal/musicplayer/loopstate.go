package musicplayer

type LoopState int64

const (
	LoopingDisabled LoopState = iota
	QueueLooping
	ItemLooping
)
