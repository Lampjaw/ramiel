package musicplayer

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/disgoorg/disgolink/dgolink"
	"github.com/disgoorg/disgolink/lavalink"
)

type LavalinkManager struct {
	Link *dgolink.Link
}

func NewLavalinkManager(lavallinkHost string, lavalinkPort string, lavalinkPassword string, session *discordgo.Session) *LavalinkManager {
	manager := &LavalinkManager{
		Link: dgolink.New(session),
	}

	manager.Link.AddNode(context.TODO(), lavalink.NodeConfig{
		Name:     "node",
		Host:     lavallinkHost,
		Port:     lavalinkPort,
		Password: lavalinkPassword,
		Secure:   false,
	})

	return manager
}

func (m *LavalinkManager) Destroy() {
	m.Link.Close()
}
