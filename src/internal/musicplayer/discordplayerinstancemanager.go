package musicplayer

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
)

type DiscordPlayerInstanceManager struct {
	session   *discordgo.Session
	instances map[string]*DiscordPlayerInstance
}

func NewDiscordPlayerInstanceManager(s *discordgo.Session) *DiscordPlayerInstanceManager {
	return &DiscordPlayerInstanceManager{
		session:   s,
		instances: make(map[string]*DiscordPlayerInstance),
	}
}

func (m *DiscordPlayerInstanceManager) VerifyPlayerInstanceExists(guildID string) bool {
	return m.instances[guildID] != nil
}

func (m *DiscordPlayerInstanceManager) GetPlayerInstance(guildID string) *DiscordPlayerInstance {
	return m.instances[guildID]
}

func (m *DiscordPlayerInstanceManager) GetOrCreatePlayerInstance(guildID string, channelID string, sourceTextChannelID string) (*DiscordPlayerInstance, error) {
	if m.instances[guildID] == nil {
		var err error
		m.instances[guildID], err = NewDiscordPlayerInstance(m.session, guildID, channelID, sourceTextChannelID)
		if err != nil {
			log.Printf("[%s] Unable to join voice channel: %v", guildID, err)
			return nil, fmt.Errorf("Unable to join voice channel")
		}
	}

	if err := m.instances[guildID].SetPlayerChannel(channelID); err != nil {
		log.Printf("[%s] Unable to set voice channel: %v", guildID, err)
		return nil, fmt.Errorf("Unable to join voice channel")
	}

	return m.instances[guildID], nil
}

func (m *DiscordPlayerInstanceManager) DestroyMusicPlayerInstance(guildID string) error {
	if m.instances[guildID] != nil {
		if err := m.instances[guildID].Destroy(); err != nil {
			return fmt.Errorf("Error destroying player instance for %s: %s", guildID, err)
		}
		delete(m.instances, guildID)
	}

	return nil
}
