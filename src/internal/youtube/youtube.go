package youtube

import (
	"net/url"
	"strings"

	yt "github.com/kkdai/youtube/v2"
)

var (
	client = yt.Client{}
)

func ResolvePlaylistData(surl string) (*yt.Playlist, error) {
	playlistURL, err := url.ParseRequestURI(surl)
	if err != nil {
		return nil, err
	}

	plID := extractPlaylistID(playlistURL)

	if plID == "" {
		return nil, nil
	}

	playlist, err := client.GetPlaylist(plID)
	if err != nil {
		return nil, err
	}

	return playlist, nil
}

func ResolveVideoData(surl string) (*yt.Video, error) {
	videoURL, err := url.ParseRequestURI(surl)
	if err != nil {
		return nil, err
	}

	vID := extractVideoID(videoURL)

	if vID == "" {
		return nil, nil
	}

	video, err := client.GetVideo(vID)
	if err != nil {
		return nil, err
	}

	return video, nil
}

func extractVideoID(u *url.URL) string {
	switch u.Host {
	case "www.youtube.com", "youtube.com", "m.youtube.com":
		if u.Path == "/watch" {
			return u.Query().Get("v")
		}
		if strings.HasPrefix(u.Path, "/embed/") {
			return u.Path[7:]
		}
	case "youtu.be":
		if len(u.Path) > 1 {
			return u.Path[1:]
		}
	}
	return ""
}

func extractPlaylistID(u *url.URL) string {
	switch u.Host {
	case "www.youtube.com", "youtube.com", "m.youtube.com":
		if u.Path == "/playlist" {
			return u.Query().Get("list")
		}
	}
	return ""
}
