using Discord;
using Discord.Interactions;
using Microsoft.Extensions.Logging;
using Microsoft.Extensions.Options;
using System.Reflection;

namespace Ramiel.Discord
{
    public class DiscordInteractionService : InteractionService
    {
        private readonly IServiceProvider _serviceProvider;

        public DiscordInteractionService(IServiceProvider serviceProvider, DiscordClient client, IOptions<InteractionServiceConfig> config,
            ILogger<DiscordInteractionService> logger)
            : base (client, config.Value)
        {
            _serviceProvider = serviceProvider;

            Log += logger.LogAsync;
        }

        public Task AddModulesAsync(Assembly assembly)
        {
            return AddModulesAsync(assembly, _serviceProvider);
        }

        public Task<IResult> ExecuteCommandAsync(IInteractionContext context)
        {
            return ExecuteCommandAsync(context, _serviceProvider);
        }
    }
}
