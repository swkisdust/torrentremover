package model

import (
	"slices"
	"time"

	"github.com/swkisdust/torrentremover/internal/utils"
)

type Torrent struct {
	Hash         string        `json:"hash" expr:"hash"`
	Name         string        `json:"name" expr:"name"`
	Ratio        float64       `json:"ratio" expr:"ratio"`
	Progress     float64       `json:"progress" expr:"progress"`
	Category     string        `json:"category" expr:"category"`
	Tags         []string      `json:"tags" expr:"tags"`
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

	Trackers []TorrentTracker `json:"trackers" expr:"trackers"`

	ClientData any `json:"-" expr:"-"` // optional field for client-specific usage
}

func (t Torrent) String() string {
	return t.Hash
}

func FilterTorrents(f *Filters, freeSpace Bytes, torrents []*Torrent) []*Torrent {
	if f.Disk != 0 && freeSpace > f.Disk {
		return nil
	}

	accesser := func(tt TorrentTracker) string {
		return tt.URL
	}

	return utils.SlicesFilter(func(t *Torrent) bool {
		if len(f.ExcludedCategories) > 0 && slices.Contains(f.ExcludedCategories, t.Category) {
			return false
		}
		if len(f.ExcludedTags) > 0 && len(t.Tags) > 0 && utils.SlicesHas(f.ExcludedTags, t.Tags...) {
			return false
		}
		if len(f.ExcludedStatus) > 0 && ContainStatus(f.ExcludedStatus, t.Status) {
			return false
		}
		if len(f.ExcludedTrackers) > 0 && len(t.Trackers) > 0 && utils.SlicesHasSubstringsFunc(t.Trackers, accesser, f.ExcludedTrackers...) {
			return false
		}
		if len(f.Categories) > 0 && !slices.Contains(f.Categories, t.Category) {
			return false
		}
		if len(f.Tags) > 0 && !utils.SlicesHas(f.Tags, t.Tags...) {
			return false
		}
		if len(f.Status) > 0 && !ContainStatus(f.Status, t.Status) {
			return false
		}
		if len(f.Trackers) > 0 && !utils.SlicesHasSubstringsFunc(t.Trackers, accesser, f.Trackers...) {
			return false
		}

		return true
	}, torrents)
}

type TorrentTracker struct {
	URL     string `json:"url" expr:"url"`
	Status  int    `json:"status" expr:"status"`
	Message string `json:"message" expr:"message"`
}
