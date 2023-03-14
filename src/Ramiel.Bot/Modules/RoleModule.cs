using Discord;
using Discord.Interactions;
using Discord.WebSocket;
using Microsoft.EntityFrameworkCore;
using Ramiel.Bot.Data;
using Ramiel.Bot.Data.Models;
using System.Data;
using System.Text;

namespace Ramiel.Bot.Modules
{
    [Group("reactionrole", "Reaction role commands")]
    [EnabledInDm(false)]
    [DefaultMemberPermissions(GuildPermission.ManageRoles)]
    public class RoleModule : InteractionModuleBase<SocketInteractionContext>
    {
        private readonly DbContextHelper _dbContextHelper;

        public RoleModule(DiscordSocketClient client, DbContextHelper dbContextHelper)
        {
            client.ReactionAdded += OnReactionAdded;
            client.ReactionRemoved += OnReactionRemoved;

            _dbContextHelper = dbContextHelper;
        }

        public async Task OnReactionAdded(Cacheable<IUserMessage, ulong> message, Cacheable<IMessageChannel, ulong> channel, SocketReaction reaction)
        {
            var guildChannel = await channel.GetOrDownloadAsync() as SocketGuildChannel;
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

            using (var dbContext = _dbContextHelper.GetDbContext())
            {
                var guildReaction = await dbContext.ReactionRoles.FindAsync(guildChannel.Guild.Id, message.Id, emoteId);
                if (guildReaction == null || guildUser.Roles.Any(a => a.Id == guildReaction.RoleId))
                {
                    return;
                }

                var role = guildChannel.Guild.GetRole(guildReaction.RoleId);
                if (role == null)
                {
                    dbContext.ReactionRoles.Remove(guildReaction);
                    await dbContext.SaveChangesAsync();
                    return;
                }

                await guildUser.AddRoleAsync(role);
            }
        }

        public async Task OnReactionRemoved(Cacheable<IUserMessage, ulong> message, Cacheable<IMessageChannel, ulong> channel, SocketReaction reaction)
        {
            var guildChannel = await channel.GetOrDownloadAsync() as SocketGuildChannel;
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

            using (var dbContext = _dbContextHelper.GetDbContext())
            {
                var guildReaction = await dbContext.ReactionRoles.FindAsync(guildChannel.Guild.Id, message.Id, emoteId);
                if (guildReaction == null)
                {
                    return;
                }

                var role = guildChannel.Guild.GetRole(guildReaction.RoleId);
                if (role == null)
                {
                    dbContext.ReactionRoles.Remove(guildReaction);
                    await dbContext.SaveChangesAsync();
                    return;
                }

                await guildUser.RemoveRoleAsync(role);
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

            using (var dbContext = _dbContextHelper.GetDbContext())
            {
                var existing = await dbContext.ReactionRoles.FindAsync(Context.Guild.Id, messageIdValue, emoteId);
                if (existing != null)
                {
                    existing.RoleId = role.Id;

                    dbContext.Update(existing);
                    await dbContext.SaveChangesAsync();
                }
                else
                {
                    var reactionRole = new ReactionRole
                    {
                        GuildId = Context.Guild.Id,
                        MessageId = messageIdValue,
                        EmoteId = emoteId,
                        EmoteDisplay = emote,
                        RoleId = role.Id
                    };

                    dbContext.Add(reactionRole);
                    await dbContext.SaveChangesAsync();
                }
            }

            await RespondAsync("Registered!");
        }

        [SlashCommand("list", "List registered reaction roles")]
        public async Task ListReactionRolesAsync()
        {
            using (var dbContext = _dbContextHelper.GetDbContext())
            {
                var roles = await dbContext.ReactionRoles.Where(a => a.GuildId == Context.Guild.Id).ToListAsync();
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

            using (var dbContext = _dbContextHelper.GetDbContext())
            {
                var existing = await dbContext.ReactionRoles.FindAsync(Context.Guild.Id, messageIdValue, emoteId);
                if (existing != null)
                {
                    dbContext.Remove(existing);

                    await dbContext.SaveChangesAsync();
                }
                else
                {
                    await RespondAsync("Failed to find a matching reaction role.");
                    return;
                }
            }

            await RespondAsync("Removed!");
        }

        [SlashCommand("clear", "Remove all reaction roles")]
        public async Task RemoveAllReactionRolesAsync()
        {
            using (var dbContext = _dbContextHelper.GetDbContext())
            {
                var allRoles = await dbContext.ReactionRoles.Where(a => a.GuildId == Context.Guild.Id).ToListAsync();
                if (allRoles != null && allRoles.Any())
                {
                    dbContext.RemoveRange(allRoles);

                    await dbContext.SaveChangesAsync();
                }
            }

            await RespondAsync("Removed!");
        }
    }
}