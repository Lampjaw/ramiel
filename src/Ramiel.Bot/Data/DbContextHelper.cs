using Microsoft.Extensions.DependencyInjection;

namespace Ramiel.Bot.Data
{
    public class DbContextHelper
    {
        private IServiceScopeFactory _scopeFactory;

        public DbContextHelper(IServiceScopeFactory serviceScopeFactory)
        {
            _scopeFactory = serviceScopeFactory;
        }

        public BotDbContext GetDbContext()
        {
            var scope = _scopeFactory.CreateScope();
            return scope.ServiceProvider.GetRequiredService<BotDbContext>();
        }
    }
}
