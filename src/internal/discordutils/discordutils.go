package discordutils

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

func GetInteractionUserVoiceChannelID(s *discordgo.Session, i *discordgo.Interaction) string {
	voiceState, _ := s.State.VoiceState(i.GuildID, i.Member.User.ID)
	return voiceState.ChannelID
}

func IsBotUser(s *discordgo.Session, userID string) bool {
	return s.State.User.ID == userID
}

func LeaveVoiceChannel(s *discordgo.Session, guildID string) error {
	if err := s.ChannelVoiceJoinManual(guildID, "", false, true); err != nil {
		return fmt.Errorf("Failed to leave voice channel for %s: %s", guildID, err)
	}

	return nil
}

func SendMessage(s *discordgo.Session, channelID string, message string) error {
	if _, err := s.ChannelMessageSend(channelID, message); err != nil {
		return err
	}

	return nil
}

func SendMessageEmbed(s *discordgo.Session, channelID string, message *discordgo.MessageEmbed) error {
	if _, err := s.ChannelMessageSendEmbed(channelID, message); err != nil {
		return err
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
