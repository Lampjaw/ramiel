using Discord;
using Discord.Interactions;
using Discord.WebSocket;
using Microsoft.Extensions.Hosting;
using Microsoft.Extensions.Options;
using Ramiel.Discord;
using System.Reflection;
using Victoria;

namespace Ramiel.Bot
{
    public class DiscordHostedService : IHostedService
    {
        private readonly IServiceProvider _serviceProvider;
        private readonly DiscordClient _client;
        private readonly DiscordInteractionService _interactionService;
        private readonly BotConfiguration _options;

        public DiscordHostedService(IServiceProvider serviceProvider, DiscordClient client, DiscordInteractionService interactionService,
            IOptions<BotConfiguration> options)
        {
            _serviceProvider = serviceProvider;
            _client = client;
            _interactionService = interactionService;
            _options = options.Value;

            client.Ready += ReadyAsync;
            client.InteractionCreated += HandleInteractionAsync;
        }

        public async Task StartAsync(CancellationToken cancellationToken)
        {
            if (string.IsNullOrWhiteSpace(_options.DiscordToken))
            {
                throw new ArgumentException("Discord Token cannot be null!");
            }

            await _interactionService.AddModulesAsync(Assembly.GetEntryAssembly());

            await _client.LoginAsync(TokenType.Bot, _options.DiscordToken);
            await _client.StartAsync();
        }

        public Task StopAsync(CancellationToken cancellationToken)
        {
            return Task.CompletedTask;
        }

        private async Task ReadyAsync()
        {
            await _serviceProvider.UseLavaNodeAsync();

            if (_options.GuildId != null)
            {
                await _interactionService.RegisterCommandsToGuildAsync(_options.GuildId.Value, true);
            }
            else
            {
                await _interactionService.RegisterCommandsGloballyAsync(true);
            }

            await _client.SetGameAsync("*screaming geometrically*", type: ActivityType.Playing);
        }

        private async Task HandleInteractionAsync(SocketInteraction interaction)
        {
            try
            {
                var context = new SocketInteractionContext(_client, interaction);

                var result = await _interactionService.ExecuteCommandAsync(context);

                if (!result.IsSuccess)
                    switch (result.Error)
                    {
                        case InteractionCommandError.UnmetPrecondition:
                            await context.Interaction.RespondAsync(result.ErrorReason);
                            break;
                        default:
                            await context.Interaction.RespondAsync(result.ErrorReason);
                            break;
                    }
            }
            catch
            {
                if (interaction.Type is InteractionType.ApplicationCommand)
                    await interaction.GetOriginalResponseAsync().ContinueWith(async (msg) => await msg.Result.DeleteAsync());
            }
        }
    }
}
