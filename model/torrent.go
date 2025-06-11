package model

import (
	"slices"
	"time"

	"github.com/autobrr/go-qbittorrent"

	"github.com/swkisdust/torrentremover/internal/utils"
)

type Torrent struct {
	Hash         string        `json:"hash" expr:"hash"`
	Name         string        `json:"name" expr:"name"`
	Ratio        float64       `json:"ratio" expr:"ratio"`
	Progress     float64       `json:"progress" expr:"progress"`
	Category     string        `json:"category" expr:"category"`
	Tag          string        `json:"tag" expr:"tag"`
	Status       Status        `json:"status" expr:"status"`
	Size         int64         `json:"size" expr:"size"`
	Leecher      int64         `json:"leecher" expr:"leecher"`
	Seeder       int64         `json:"seeder" expr:"seeder"`
	DlSpeed      int64         `json:"dl_speed" expr:"dl_speed"`
	UpSpeed      int64         `json:"up_speed" expr:"up_speed"`
	AvgDlSpeed   int64         `json:"avg_dl_speed" expr:"avg_dl_speed"`
	AvgUpSpeed   int64         `json:"avg_up_speed" expr:"avg_up_speed"`
	Downloaded   int64         `json:"downloaded" expr:"downloaded"`
	Uploaded     int64         `json:"uploaded" expr:"uploaded"`
	AddedTime    time.Time     `json:"added_time" expr:"added_time"`
	LastActivity time.Time     `json:"last_activity" expr:"last_activity"`
	SeedingTime  time.Duration `json:"seeding_time" expr:"seeding_time"`
	TimeElapsed  time.Duration `json:"time_elapsed" expr:"time_elapsed"`

	Trackers []string `json:"trackers" expr:"-"`
}

func (t Torrent) String() string {
	return t.Hash
}

func FromQbit(torrent qbittorrent.Torrent, prop qbittorrent.TorrentProperties) Torrent {
	return Torrent{
		AddedTime:    time.Unix(torrent.AddedOn, 0),
		LastActivity: time.Unix(torrent.LastActivity, 0),
		TimeElapsed:  time.Duration(prop.TimeElapsed) * time.Second,
		SeedingTime:  time.Duration(prop.SeedingTime) * time.Second,
		Hash:         torrent.Hash,
		Name:         torrent.Name,
		Status:       GetStatus(string(torrent.State)),
		Ratio:        torrent.Ratio,
		Progress:     torrent.Progress,
		Category:     torrent.Category,
		Tag:          torrent.Tags,
		Size:         torrent.Size,
		Leecher:      torrent.NumLeechs,
		Seeder:       torrent.NumSeeds,
		DlSpeed:      torrent.DlSpeed,
		UpSpeed:      torrent.UpSpeed,
		AvgDlSpeed:   int64(prop.DlSpeedAvg),
		AvgUpSpeed:   int64(prop.UpSpeedAvg),
		Downloaded:   torrent.Downloaded,
		Uploaded:     torrent.Uploaded,
		Trackers: slices.Collect(utils.IterMap(slices.Values(torrent.Trackers),
			func(qt qbittorrent.TorrentTracker) string {
				return qt.Url
			})),
	}
}

func FilterTorrents(f Filter, torrents []Torrent) []Torrent {
	dup := make([]Torrent, len(torrents))
	copy(dup, torrents)

	return slices.DeleteFunc(dup, func(torrent Torrent) bool {
		if f.ExcludedCategories != nil && slices.Contains(f.ExcludedCategories, torrent.Category) {
			return true
		}
		if f.ExcludedTags != nil && slices.Contains(f.ExcludedTags, torrent.Tag) {
			return true
		}
		if f.ExcludedStatus != nil && ContainStatus(f.ExcludedStatus, torrent.Status) {
			return true
		}
		if f.ExcludedTrackers != nil && torrent.Trackers != nil && utils.SlicesHaveSubstrings(torrent.Trackers, f.ExcludedTrackers...) {
			return true
		}
		if f.Categories != nil && !slices.Contains(f.Categories, torrent.Category) {
			return true
		}
		if f.Tags != nil && !slices.Contains(f.Tags, torrent.Tag) {
			return true
		}
		if f.Status != nil && !ContainStatus(f.Status, torrent.Status) {
			return true
		}
		if f.Trackers != nil && torrent.Trackers != nil && !utils.SlicesHaveSubstrings(torrent.Trackers, f.Trackers...) {
			return true
		}
		return false
	})
}
