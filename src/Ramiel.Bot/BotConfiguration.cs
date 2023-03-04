namespace Ramiel.Bot
{
    public class BotConfiguration
    {
        public string DiscordToken { get; set; }
        public ulong? GuildId { get; set; }
        public string LavalinkHostname { get; set; }
        public ushort LavalinkPort { get; set; }
        public string LavalinkPassword { get; set; }
    }
}
