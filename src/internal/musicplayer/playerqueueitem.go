package musicplayer

import (
	"fmt"
	"time"

	yt "github.com/kkdai/youtube/v2"
)

type PlaylistInfo struct {
	Title string
	Items []*PlayerQueueItem
}

type PlayerQueueItem struct {
	Url          string
	RequestedBy  string
	VideoID      string
	Title        string
	Author       string
	Duration     time.Duration
	ThumbnailURL string
	Video        *yt.Video
}

func newPlaylistInfo(member string, p *yt.Playlist) *PlaylistInfo {
	return &PlaylistInfo{
		Title: p.Title,
		Items: newPlaylistItems(member, p.Videos),
	}
}

func newPlaylistItems(member string, p []*yt.PlaylistEntry) []*PlayerQueueItem {
	items := make([]*PlayerQueueItem, 0)
	for _, v := range p {
		items = append(items, &PlayerQueueItem{
			Url:         fmt.Sprintf("https://www.youtube.com/watch?v=%s", v.ID),
			RequestedBy: member,
			VideoID:     v.ID,
			Title:       v.Title,
			Author:      v.Author,
			Duration:    v.Duration,
		})
	}
	return items
}

func newPlayerQueueItem(member string, v *yt.Video) *PlayerQueueItem {
	return &PlayerQueueItem{
		Url:          fmt.Sprintf("https://www.youtube.com/watch?v=%s", v.ID),
		RequestedBy:  member,
		VideoID:      v.ID,
		Title:        v.Title,
		Duration:     v.Duration,
		Author:       v.Author,
		ThumbnailURL: v.Thumbnails[0].URL,
		Video:        v,
	}
}
