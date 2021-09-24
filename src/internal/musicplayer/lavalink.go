package musicplayer

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/foxbot/gavalink"
)

type LavalinkManager struct {
	lavalink     *gavalink.Lavalink
	Player       *gavalink.Player
	isTrackEnded chan bool
}

func NewLavalinkManager(lavallinkEndpoint string, lavalinkPassword string, discordSession *discordgo.Session) (*LavalinkManager, error) {
	lavalink := gavalink.NewLavalink("1", discordSession.State.User.ID)

	err := lavalink.AddNodes(gavalink.NodeConfig{
		REST:      fmt.Sprintf("http://%s", lavallinkEndpoint),
		WebSocket: fmt.Sprintf("ws://%s", lavallinkEndpoint),
		Password:  lavalinkPassword,
	})
	if err != nil {
		return nil, err
	}

	m := &LavalinkManager{
		lavalink:     lavalink,
		isTrackEnded: make(chan bool),
	}

	discordSession.AddHandler(func(s *discordgo.Session, event *discordgo.VoiceServerUpdate) {
		m.voiceServerUpdate(s, event)
	})

	return m, nil
}

func (m *LavalinkManager) Play(query string) error {
	node, err := m.lavalink.BestNode()
	if err != nil {
		return err
	}

	tracks, err := node.LoadTracks(query)
	if err != nil {
		return err
	}

	if tracks.Type != gavalink.TrackLoaded {
		return fmt.Errorf("Unhandled track type: %v", tracks.Type)
	}

	err = m.Player.Play(tracks.Tracks[0].Data)
	if err != nil {
		return err
	}

	return nil
}

func (m *LavalinkManager) Close() error {
	return m.Player.Destroy()
}

func (m *LavalinkManager) voiceServerUpdate(s *discordgo.Session, event *discordgo.VoiceServerUpdate) error {
	vsu := gavalink.VoiceServerUpdate{
		Endpoint: event.Endpoint,
		GuildID:  event.GuildID,
		Token:    event.Token,
	}

	if p, err := m.lavalink.GetPlayer(event.GuildID); err == nil {
		err = p.Forward(s.State.SessionID, vsu)
		if err != nil {
			return err
		}
	}

	node, err := m.lavalink.BestNode()
	if err != nil {
		return err
	}

	eventHandler := &LavalinkEventHandler{
		manager: m,
	}

	m.Player, err = node.CreatePlayer(event.GuildID, s.State.SessionID, vsu, eventHandler)
	if err != nil {
		return err
	}

	m.Player.Volume(50)

	return nil
}

type LavalinkEventHandler struct {
	manager *LavalinkManager
	gavalink.EventHandler
}

func (h *LavalinkEventHandler) OnTrackEnd(player *gavalink.Player, track string, reason string) error {
	if reason == "FINISHED" {
		go func() {
			h.manager.isTrackEnded <- true
		}()
	}

	return nil
}
