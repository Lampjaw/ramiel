using Discord;
using Microsoft.Extensions.Logging;

namespace Ramiel.Discord
{
    public static class LoggerExtensions
    {
        private static readonly Dictionary<LogSeverity, LogLevel> _logLevelMap = new()
        {
            { LogSeverity.Critical, LogLevel.Critical },
            { LogSeverity.Error, LogLevel.Error },
            { LogSeverity.Warning, LogLevel.Warning },
            { LogSeverity.Info, LogLevel.Information },
            { LogSeverity.Verbose, LogLevel.Trace },
            { LogSeverity.Debug, LogLevel.Debug }
        };

        public static async Task LogAsync(this ILogger logger, LogMessage message)
        {
            logger.Log(_logLevelMap[message.Severity], message.Exception, message.Message, message.Source);
        }
    }
}
