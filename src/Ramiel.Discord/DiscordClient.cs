using Discord.WebSocket;
using Microsoft.Extensions.Logging;
using Microsoft.Extensions.Options;

namespace Ramiel.Discord
{
    public class DiscordClient : DiscordSocketClient
    {
        public DiscordClient(IOptions<DiscordSocketConfig> config, ILogger<DiscordClient> logger)
            :base(config.Value)
        {
            Log += logger.LogAsync;
        }
    }
}
