using Discord.Interactions;
using Discord.WebSocket;

namespace Ramiel.Discord
{
    public class DiscordConfig
    {
        public DiscordSocketConfig DiscordSocket { get; set; } = new();
        public InteractionServiceConfig InteractionService { get; set; } = new();
    }
}
