package discordutils

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

func IsBotUser(s *discordgo.Session, userID string) bool {
	return s.State.User.ID == userID
}

func LeaveVoiceChannel(s *discordgo.Session, guildID string) error {
	if err := s.ChannelVoiceJoinManual(guildID, "", false, true); err != nil {
		return fmt.Errorf("Failed to leave voice channel for %s: %s", guildID, err)
	}

	return nil
}

func SendMessageResponse(s *discordgo.Session, i *discordgo.Interaction, message string) error {
	responseData := &discordgo.InteractionResponseData{
		Content: message}

	return sendInteractionResponseContent(s, i, responseData)
}

func SendEmbedResponse(s *discordgo.Session, i *discordgo.Interaction, message *discordgo.MessageEmbed) error {
	responseData := &discordgo.InteractionResponseData{
		Embeds: []*discordgo.MessageEmbed{
			message,
		}}

	return sendInteractionResponseContent(s, i, responseData)
}

func sendInteractionResponseContent(s *discordgo.Session, i *discordgo.Interaction, d *discordgo.InteractionResponseData) error {
	response := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: d}

	if err := s.InteractionRespond(i, response); err != nil {
		return fmt.Errorf("Failed to send response: %s: %s", i.GuildID, err)
	}

	return nil
}
