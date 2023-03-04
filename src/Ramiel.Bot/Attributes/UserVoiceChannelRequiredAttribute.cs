using Discord;
using Discord.Interactions;

namespace Ramiel.Bot.Attributes
{
    public class UserVoiceChannelRequiredAttribute : PreconditionAttribute
    {
        public override async Task<PreconditionResult> CheckRequirementsAsync(IInteractionContext context, ICommandInfo commandInfo, IServiceProvider services)
        {
            var userId = context.User.Id;
            var voiceChannels = await context.Guild.GetVoiceChannelsAsync();
            var userVoiceChannel = voiceChannels.FirstOrDefault(a => a.GetUserAsync(userId) != null);

            if (userVoiceChannel == null)
            {
                return PreconditionResult.FromError("You need to be in a voice channel to use this command!");
            }

            return PreconditionResult.FromSuccess();
        }
    }
}
