package model

import (
	"slices"
	"time"

	"github.com/swkisdust/torrentremover/internal/utils"
)

type Torrent struct {
	ID           any           `json:"id" expr:"-"`
	Hash         string        `json:"hash" expr:"hash"`
	Name         string        `json:"name" expr:"name"`
	Ratio        float64       `json:"ratio" expr:"ratio"`
	Progress     float64       `json:"progress" expr:"progress"`
	Category     string        `json:"category" expr:"-"`
	Tags         []string      `json:"tags" expr:"-"`
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

func FilterTorrents(f Filter, freeSpace Bytes, torrents []Torrent) []Torrent {
	if f.Disk != 0 && freeSpace > f.Disk {
		return nil
	}

	dup := make([]Torrent, len(torrents))
	copy(dup, torrents)

	return slices.DeleteFunc(dup, func(torrent Torrent) bool {
		if len(f.ExcludedCategories) > 0 && slices.Contains(f.ExcludedCategories, torrent.Category) {
			return true
		}
		if len(f.ExcludedTags) > 0 && len(torrent.Tags) > 0 && utils.SlicesHave(f.ExcludedTags, torrent.Tags...) {
			return true
		}
		if len(f.ExcludedStatus) > 0 && ContainStatus(f.ExcludedStatus, torrent.Status) {
			return true
		}
		if len(f.ExcludedTrackers) > 0 && len(torrent.Trackers) > 0 && utils.SlicesHaveSubstrings(torrent.Trackers, f.ExcludedTrackers...) {
			return true
		}
		if len(f.Categories) > 0 && !slices.Contains(f.Categories, torrent.Category) {
			return true
		}
		if len(f.Tags) > 0 && !utils.SlicesHave(f.Tags, torrent.Tags...) {
			return true
		}
		if len(f.Status) > 0 && !ContainStatus(f.Status, torrent.Status) {
			return true
		}
		if len(f.Trackers) > 0 && !utils.SlicesHaveSubstrings(torrent.Trackers, f.Trackers...) {
			return true
		}
		return false
	})
}
