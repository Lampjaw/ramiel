using Microsoft.EntityFrameworkCore;
using Ramiel.Bot.Data.Models;

namespace Ramiel.Bot.Data
{
    public class BotDbContext : DbContext
    {
        public string DbPath { get; }

        public BotDbContext()
        {
            var currentPath = Path.GetDirectoryName(Environment.CurrentDirectory);
            var dataPath = Path.Join(currentPath, "/data");
            Directory.CreateDirectory(dataPath);

            DbPath = Path.Combine(dataPath, "bot.db");

            Database.EnsureCreated();
        }

        protected override void OnConfiguring(DbContextOptionsBuilder options)
            => options.UseSqlite($"Data Source={DbPath}");

        public DbSet<ReactionRole> ReactionRoles { get; set; }
    }
}
