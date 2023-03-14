using Microsoft.EntityFrameworkCore;

namespace Ramiel.Bot.Data.Models
{
    [PrimaryKey("GuildId", "MessageId", "EmoteId")]
    public class ReactionRole
    {
        public ulong GuildId { get; set; }
        public ulong MessageId { get; set; }
        public string EmoteId { get; set; }
        public string EmoteDisplay { get; set; }
        public ulong RoleId { get; set; }
    }
}
