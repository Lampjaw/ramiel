package musicplayer

import (
	"context"
	"errors"
	"fmt"
	"ramiel/internal/discordutils"

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

func (p *MusicPlayer) Destroy() {
	p.lavalinkManager.Destroy()
}

func (p *MusicPlayer) PlayerExists(guildID string) bool {
	_, exists := p.PlayerManagers[guildID]
	return exists
}

func (p *MusicPlayer) GetPlayerManager(guildID string) *PlayerManager {
	manager, exists := p.PlayerManagers[guildID]
	if !exists {
		panic(errors.New("Not running in a channel! Try playing something first."))
	}
	return manager
}

func (p *MusicPlayer) PlayQuery(guildID string, voiceChannelID string, channelID string, requestedBy string, query string) {
	if err := p.discordSession.ChannelVoiceJoinManual(guildID, voiceChannelID, false, false); err != nil {
		_, _ = p.discordSession.ChannelMessageSend(channelID, "error while joining voice channel: "+err.Error())
		return
	}

	manager, ok := p.PlayerManagers[guildID]
	if !ok {
		manager = NewPlayerManager(p.lavalinkManager.Link, guildID)
		p.PlayerManagers[guildID] = manager
	}

	p.lavalinkManager.Link.BestRestClient().LoadItemHandler(context.TODO(), query, lavalink.NewResultHandler(
		func(track lavalink.AudioTrack) {
			track.SetUserData(requestedBy)

			manager.Queue.AddTracks(track)

			_, _ = p.discordSession.ChannelMessageSend(channelID, fmt.Sprintf("Added %s to the queue.", track.Info().Title))
		},
		func(playlist lavalink.AudioPlaylist) {
			for _, t := range playlist.Tracks() {
				t.SetUserData(requestedBy)
			}

			manager.Queue.AddTracks(playlist.Tracks()...)

			_, _ = p.discordSession.ChannelMessageSend(channelID, fmt.Sprintf("Loaded %d tracks from the `%s` playlist.", len(playlist.Tracks()), playlist.Name()))
		},
		func(tracks []lavalink.AudioTrack) {
			for _, t := range tracks {
				t.SetUserData(requestedBy)
			}

			manager.Queue.AddTracks(tracks...)

			_, _ = p.discordSession.ChannelMessageSend(channelID, fmt.Sprintf("Loaded %d tracks from search results.", len(tracks)))
		},
		func() {
			_, _ = p.discordSession.ChannelMessageSend(channelID, "no matches found for: "+query)
			return
		},
		func(ex lavalink.FriendlyException) {
			_, _ = p.discordSession.ChannelMessageSend(channelID, "error while loading track: "+ex.Message)
			return
		},
	))

	manager.Play()
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
