using Discord.Interactions;
using Discord.WebSocket;
using Microsoft.Extensions.DependencyInjection;

namespace Ramiel.Discord
{
    public static class ServiceCollectionExtensions
    {
        public static IServiceCollection AddDiscordServices(this IServiceCollection services, Action<DiscordConfig> config = null)
        {
            var discordConfig = new DiscordConfig();
            config?.Invoke(discordConfig);

            services.AddOptions<DiscordSocketConfig>().Configure(a => a = discordConfig.DiscordSocket);
            services.AddOptions<InteractionServiceConfig>().Configure(a => a = discordConfig.InteractionService);

            services.AddSingleton<DiscordClient>();
            services.AddSingleton<DiscordSocketClient>(sp => sp.GetRequiredService<DiscordClient>());
            services.AddSingleton<DiscordInteractionService>();
            services.AddSingleton<InteractionService>(sp => sp.GetRequiredService<DiscordInteractionService>());

            return services;
        }
    }
}