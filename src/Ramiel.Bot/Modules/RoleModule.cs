using Discord;
using Discord.Interactions;
using Discord.WebSocket;
using Ramiel.Bot.Services;
using System.Text;

namespace Ramiel.Bot.Modules
{
    [Group("reactionrole", "Reaction role commands")]
    [EnabledInDm(false)]
    [DefaultMemberPermissions(GuildPermission.ManageRoles)]
    public class RoleModule : InteractionModuleBase<SocketInteractionContext>
    {
        private readonly ReactionRoleStore _reactionRoleStore;
        private readonly SemaphoreSlim _reactionSemaphore = new SemaphoreSlim(1, 1);

        public RoleModule(DiscordSocketClient client, ReactionRoleStore reactionRoleStore)
        {
            client.ReactionAdded += OnReactionAdded;
            client.ReactionRemoved += OnReactionRemoved;

            _reactionRoleStore = reactionRoleStore;
        }

        public async Task OnReactionAdded(Cacheable<IUserMessage, ulong> message, Cacheable<IMessageChannel, ulong> channel, SocketReaction reaction)
        {
            var guildChannel = await channel.GetOrDownloadAsync() as SocketGuildChannel;

            if (!await _reactionRoleStore.IsGuildReactionMessageAsync(guildChannel.Guild.Id, message.Id))
            {
                return;
            }

            var guildUser = guildChannel.GetUser(reaction.UserId);
            
            var emoteId = reaction.Emote.ToString();
            if (reaction.Emote is Emote emoteValue)
            {
                emoteId = emoteValue.Id.ToString();
            }

            if (guildChannel == null || guildUser == null || emoteId == null)
            {
                return;
            }

            var guildReaction = await _reactionRoleStore.TryGetAsync(guildChannel.Guild.Id, message.Id, emoteId);
            if (guildReaction == null || guildUser.Roles.Any(a => a.Id == guildReaction.RoleId))
            {
                return;
            }

            var role = guildChannel.Guild.GetRole(guildReaction.RoleId);
            if (role == null)
            {
                await _reactionRoleStore.RemoveAsync(guildChannel.Guild.Id, message.Id, emoteId);
                return;
            }

            try
            {
                await _reactionSemaphore.WaitAsync();

                await guildUser.AddRoleAsync(role);
            }
            finally
            {
                _reactionSemaphore.Release();
            }
        }

        public async Task OnReactionRemoved(Cacheable<IUserMessage, ulong> message, Cacheable<IMessageChannel, ulong> channel, SocketReaction reaction)
        {
            var guildChannel = await channel.GetOrDownloadAsync() as SocketGuildChannel;

            if (!await _reactionRoleStore.IsGuildReactionMessageAsync(guildChannel.Guild.Id, message.Id))
            {
                return;
            }

            var guildUser = guildChannel.GetUser(reaction.UserId);

            var emoteId = reaction.Emote.ToString();
            if (reaction.Emote is Emote emoteValue)
            {
                emoteId = emoteValue.Id.ToString();
            }

            if (guildChannel == null || guildUser == null || emoteId == null)
            {
                return;
            }

            var guildReaction = await _reactionRoleStore.TryGetAsync(guildChannel.Guild.Id, message.Id, emoteId);
            if (guildReaction == null)
            {
                return;
            }

            var role = guildChannel.Guild.GetRole(guildReaction.RoleId);
            if (role == null)
            {
                await _reactionRoleStore.RemoveAsync(guildChannel.Guild.Id, message.Id, emoteId);
                return;
            }

            try
            {
                await _reactionSemaphore.WaitAsync();

                await guildUser.RemoveRoleAsync(role);
            }
            finally
            {
                _reactionSemaphore.Release();
            }
        }

        [SlashCommand("add", "Register a reaction role emote")]
        public async Task AddReactionRoleAsync(string messageId, string emote, SocketRole role)
        {
            if (!ulong.TryParse(messageId, out var messageIdValue))
            {
                await RespondAsync("Please use a valid message id.");
                return;
            }

            var emoteId = emote;
            if (Emote.TryParse(emote, out var emoteValue))
            {
                emoteId = emoteValue.Id.ToString();
            }

            await _reactionRoleStore.AddOrUpdateRoleIdAsync(Context.Guild.Id, messageIdValue, emoteId, emote, role.Id);

            await RespondAsync("Registered!");
        }

        [SlashCommand("list", "List registered reaction roles")]
        public async Task ListReactionRolesAsync()
        {
            var roles = await _reactionRoleStore.GetAllForGuildIdAsync(Context.Guild.Id);
            if (roles == null || !roles.Any())
            {
                await RespondAsync("None found.");
                return;
            }

            var sb = new StringBuilder();
            sb.AppendLine("MessageId, EmoteId, Role");
            foreach (var role in roles)
            {
                sb.AppendLine($"{role.MessageId}, {role.EmoteDisplay}, <@&{role.RoleId}>");
            }

            var embed = new EmbedBuilder()
                .WithTitle("Reaction roles")
                .WithDescription(sb.ToString())
                .Build();

            await RespondAsync(embed: embed);
        }

        [SlashCommand("remove", "Remove a reaction role")]
        public async Task RemoveReactionRoleAsync(string messageId, string emote)
        {
            if (!ulong.TryParse(messageId, out var messageIdValue))
            {
                await RespondAsync("Please use a valid message id.");
                return;
            }

            var emoteId = emote;
            if (Emote.TryParse(emote, out var emoteValue))
            {
                emoteId = emoteValue.Id.ToString();
            }

            await _reactionRoleStore.RemoveAsync(Context.Guild.Id, messageIdValue, emoteId);

            await RespondAsync("Removed!");
        }

        [SlashCommand("clear", "Remove all reaction roles")]
        public async Task RemoveAllReactionRolesAsync()
        {
            await _reactionRoleStore.RemoveAllForGuildIdAsync(Context.Guild.Id);

            await RespondAsync("Removed!");
        }
    }
}