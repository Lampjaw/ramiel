using Microsoft.EntityFrameworkCore;
using Ramiel.Bot.Data;
using Ramiel.Bot.Data.Models;
using System.Collections.Concurrent;

namespace Ramiel.Bot.Services
{
    public class ReactionRoleStore
    {
        private readonly DbContextHelper _dbContextHelper;

        private readonly ConcurrentDictionary<string, ReactionRole> _storeCache = new ();

        private Func<ulong, ulong, string, string> _getCacheKey = (guildId, messageId, emoteId) => $"g:{guildId}-m:{messageId}-e:{emoteId}";

        public ReactionRoleStore(DbContextHelper dbContextHelper)
        {
            _dbContextHelper = dbContextHelper;

            PopulateCacheAsync().Wait();
        }

        public async Task PopulateCacheAsync()
        {
            using (var dbContext = _dbContextHelper.GetDbContext())
            {
                var reactionRoles = await dbContext.ReactionRoles.ToListAsync();

                foreach (var reactionRole in reactionRoles)
                {
                    var cacheKey = _getCacheKey(reactionRole.GuildId, reactionRole.MessageId, reactionRole.EmoteId);
                    _storeCache.TryAdd(cacheKey, reactionRole);
                }
            }
        }

        public async Task<bool> IsGuildReactionMessageAsync(ulong guildId, ulong messageId)
        {
            return _storeCache.Keys.Any(a => a.Contains($"g:{guildId}-m:{messageId}"));
        }

        public async Task<ReactionRole> TryGetAsync(ulong guildId, ulong messageId, string emoteId)
        {
            var cacheKey = _getCacheKey(guildId, messageId, emoteId);

            if (_storeCache.TryGetValue(cacheKey, out var value))
            {
                return value;
            }

            return null;
        }

        public async Task AddOrUpdateRoleIdAsync(ulong guildId, ulong messageId, string emoteId, string emote, ulong roleId)
        {
            ReactionRole? reactionRole;

            using (var dbContext = _dbContextHelper.GetDbContext())
            {
                reactionRole = await dbContext.ReactionRoles.FindAsync(guildId, messageId, emoteId);
                if (reactionRole != null)
                {
                    reactionRole.RoleId = roleId;

                    dbContext.Update(reactionRole);
                    await dbContext.SaveChangesAsync();
                }
                else
                {
                    reactionRole = new ReactionRole
                    {
                        GuildId = guildId,
                        MessageId = messageId,
                        EmoteId = emoteId,
                        EmoteDisplay = emote,
                        RoleId = roleId
                    };

                    dbContext.Add(reactionRole);
                    await dbContext.SaveChangesAsync();
                }
            }

            var cacheKey = _getCacheKey(guildId, messageId, emoteId);

            _storeCache.AddOrUpdate(cacheKey, reactionRole, (key, oldValue) =>
            {
                oldValue.RoleId = roleId;
                return oldValue;
            });
        }

        public async Task RemoveAsync(ulong guildId, ulong messageId, string emoteId)
        {
            using (var dbContext = _dbContextHelper.GetDbContext())
            {
                var existing = await dbContext.ReactionRoles.FindAsync(guildId, messageId, emoteId);
                if (existing != null)
                {
                    dbContext.Remove(existing);

                    await dbContext.SaveChangesAsync();
                }
            }

            var cacheKey = _getCacheKey(guildId, messageId, emoteId);

            _storeCache.TryRemove(cacheKey, out _);
        }

        public async Task<IEnumerable<ReactionRole>> GetAllForGuildIdAsync(ulong guildId)
        {
            using (var dbContext = _dbContextHelper.GetDbContext())
            {
                var roles = await dbContext.ReactionRoles.Where(a => a.GuildId == guildId).ToListAsync();
                if (roles == null || !roles.Any())
                {
                    return Enumerable.Empty<ReactionRole>();
                }

                return roles;
            }
        }

        public async Task RemoveAllForGuildIdAsync(ulong guildId)
        {
            using (var dbContext = _dbContextHelper.GetDbContext())
            {
                var allRoles = await dbContext.ReactionRoles.Where(a => a.GuildId == guildId).ToListAsync();
                if (allRoles != null && allRoles.Any())
                {
                    dbContext.RemoveRange(allRoles);

                    await dbContext.SaveChangesAsync();
                }
            }

            _storeCache.Keys
                .Where(a => a.Contains($"g:{guildId}")).ToList()
                .ForEach(k => _storeCache.Remove(k, out _));
        }
    }
}
