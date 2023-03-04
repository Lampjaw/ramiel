using Discord;
using Microsoft.Extensions.Logging;
using Ramiel.Bot.Models;
using Victoria.Node;
using Victoria.Node.EventArgs;
using Victoria.Player;
using Victoria.Responses.Search;

namespace Ramiel.Bot.Services
{
    public class MusicService
    {
        private readonly LavaNode _lavaNode;
        private readonly ILogger _logger;

        private readonly Dictionary<ulong, LoopTypeEnum> _guildLoopType = new();

        public MusicService(LavaNode lavaNode, ILogger<MusicService> logger)
        {
            _lavaNode = lavaNode;
            _logger = logger;

            _lavaNode.OnTrackEnd += OnTrackEndAsync;
            _lavaNode.OnWebSocketClosed += OnWebSocketClosedAsync;
            _lavaNode.OnTrackStuck += OnTrackStuckAsync;
            _lavaNode.OnTrackException += OnTrackExceptionAsync;
        }

        public bool IsPlaying(IGuild guild)
        {
            return _lavaNode.HasPlayer(guild);
        }

        public async Task JoinAsync(IVoiceState voiceState, ITextChannel textChannel)
        {
            var player = await _lavaNode.JoinAsync(voiceState.VoiceChannel, textChannel);
            await player.SetVolumeAsync(30);
        }

        public TimeSpan GetPlayerDuration(IGuild guild)
        {
            if (!_lavaNode.TryGetPlayer(guild, out var player))
            {
                return TimeSpan.Zero;
            }

            var activeRemaining = player.Track.Duration - player.Track.Position;
            return TimeSpan.FromSeconds(activeRemaining.TotalSeconds + player.Vueue.Sum(a => a.Duration.Seconds));
        }

        public async Task<string> PlayAsync(IGuild guild, ITextChannel textChannel, string searchQuery)
        {
            if (!_lavaNode.TryGetPlayer(guild, out var player))
            {
                return "Player error.";
            }

            var searchType = Uri.IsWellFormedUriString(searchQuery, UriKind.Absolute) ? SearchType.Direct : SearchType.YouTube;
            var searchResponse = await _lavaNode.SearchAsync(searchType, searchQuery);

            if (searchResponse.Status is SearchStatus.LoadFailed or SearchStatus.NoMatches)
            {
                return $"I wasn't able to find anything for `{searchQuery}`.";
            }

            string returnMessage;
            if (!string.IsNullOrWhiteSpace(searchResponse.Playlist.Name))
            {
                player.Vueue.Enqueue(searchResponse.Tracks);
                returnMessage = $"Added {searchResponse.Tracks.Count} tracks from the '{searchResponse.Playlist.Name}' playlist to the queue.";
            }
            else
            {
                var track = searchResponse.Tracks.FirstOrDefault();
                player.Vueue.Enqueue(track);

                returnMessage = $"Added '{track?.Title}' to the queue.";
            }

            if (player.PlayerState is PlayerState.Playing or PlayerState.Paused)
            {
                return returnMessage;
            }

            player.Vueue.TryDequeue(out var lavaTrack);
            await player.PlayAsync(lavaTrack);

            return returnMessage;
        }

        public async Task StopAsync(IGuild guild)
        {
            if (!_lavaNode.TryGetPlayer(guild, out var player))
            {
                return;
            }

            await player.PauseAsync();
        }

        public async Task ResumeAsync(IGuild guild)
        {
            if (!_lavaNode.TryGetPlayer(guild, out var player))
            {
                return;
            }

            await player.ResumeAsync();
        }

        public async Task<IEnumerable<LavaTrack>> QueueAsync(IGuild guild)
        {
            if (!_lavaNode.TryGetPlayer(guild, out var player))
            {
                return Enumerable.Empty<LavaTrack>();
            }

            return player.Vueue.ToList();
        }

        public async Task ShuffleAsync(IGuild guild)
        {
            if (!_lavaNode.TryGetPlayer(guild, out var player))
            {
                return;
            }

            player.Vueue.Shuffle();
        }

        public async Task SkipAsync(IGuild guild)
        {
            if (!_lavaNode.TryGetPlayer(guild, out var player))
            {
                return;
            }

            if (player.PlayerState != PlayerState.Playing)
            {
                return;
            }

            SetLoopType(guild, LoopTypeEnum.Off);

            await player.SkipAsync();
        }

        public async Task ClearAsync(IGuild guild)
        {
            if (!_lavaNode.TryGetPlayer(guild, out var player))
            {
                return;
            }

            player.Vueue.Clear();
        }

        public async Task<LavaTrack> NowPlayingAsync(IGuild guild)
        {
            if (!_lavaNode.TryGetPlayer(guild, out var player))
            {
                return null;
            }

            return player.Track;
        }

        public async Task DisconnectAsync(IGuild guild)
        {
            if (!_lavaNode.TryGetPlayer(guild, out var player))
            {
                return;
            }

            var voiceChannel = player.VoiceChannel;

            await _lavaNode.LeaveAsync(voiceChannel);

            if (_guildLoopType.ContainsKey(guild.Id))
            {
                _guildLoopType.Remove(guild.Id);
            }
        }

        public async Task RemoveDuplicatesAsync(IGuild guild)
        {
            if (!_lavaNode.TryGetPlayer(guild, out var player))
            {
                return;
            }

            for (var i = 0; i < player.Vueue.Count - 1; i++)
            {
                for (var k = i + 1; k < player.Vueue.Count; k++)
                {
                    if (player.Vueue.ElementAt(i).Id == player.Vueue.ElementAt(k).Id)
                    {
                        player.Vueue.RemoveAt(k);
                        k--;
                    }
                }
            }
        }

        public async Task LoopAsync(IGuild guild, LoopTypeEnum loopType)
        {
            if (!_lavaNode.TryGetPlayer(guild, out var player))
            {
                return;
            }

            SetLoopType(guild, loopType);
        }

        public LoopTypeEnum GetLoopType(IGuild guild)
        {
            return _guildLoopType.GetValueOrDefault(guild.Id, LoopTypeEnum.Off);
        }

        private void SetLoopType(IGuild guild, LoopTypeEnum loopType)
        {
            if (_guildLoopType.ContainsKey(guild.Id) && loopType == LoopTypeEnum.Off)
            {
                _guildLoopType.Remove(guild.Id);
            }
            else
            {
                _guildLoopType[guild.Id] = loopType;
            }
        }

        private static Task OnTrackExceptionAsync(TrackExceptionEventArg<LavaPlayer<LavaTrack>, LavaTrack> arg)
        {
            return arg.Player.TextChannel.SendMessageAsync($"{arg.Track} failed to play. Skipping.");
        }

        private static Task OnTrackStuckAsync(TrackStuckEventArg<LavaPlayer<LavaTrack>, LavaTrack> arg)
        {
            return arg.Player.TextChannel.SendMessageAsync($"{arg.Track} is stuck! Skipping.");
        }

        private Task OnWebSocketClosedAsync(WebSocketClosedEventArg arg)
        {
            SetLoopType(arg.Guild, LoopTypeEnum.Off);

            _logger.LogCritical($"{arg.Code} {arg.Reason}");
            return Task.CompletedTask;
        }

        private async Task OnTrackEndAsync(TrackEndEventArg<LavaPlayer<LavaTrack>, LavaTrack> arg)
        {
            if (arg.Reason is TrackEndReason.Stopped or TrackEndReason.Replaced)
            {
                return;
            }

            var loopType = GetLoopType(arg.Player.VoiceChannel.Guild);
            if (loopType == LoopTypeEnum.Track)
            {
                await arg.Player.PlayAsync(arg.Track);
                return;
            }
            else if (loopType == LoopTypeEnum.Queue)
            {
                arg.Player.Vueue.Enqueue(arg.Track);
            }

            if (arg.Player.Vueue.TryDequeue(out var track))
            {
                await arg.Player.PlayAsync(track);
            }
        }
    }
}
