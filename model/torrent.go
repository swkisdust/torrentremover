package model

import (
	"slices"
	"strings"
	"time"

	"github.com/autobrr/go-qbittorrent"
	"github.com/hekmon/transmissionrpc/v3"

	"github.com/swkisdust/torrentremover/internal/utils"
)

type Torrent struct {
	ID           int64         `json:"id" expr:"-"`
	Hash         string        `json:"hash" expr:"-"`
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

func FromQbit(torrent qbittorrent.Torrent, prop qbittorrent.TorrentProperties) Torrent {
	return Torrent{
		AddedTime:    time.Unix(torrent.AddedOn, 0),
		LastActivity: utils.IfOr(torrent.LastActivity == 0, time.Time{}, time.Unix(torrent.LastActivity, 0)),
		TimeElapsed:  time.Duration(prop.TimeElapsed) * time.Second,
		SeedingTime:  time.Duration(prop.SeedingTime) * time.Second,
		Hash:         torrent.Hash,
		Name:         torrent.Name,
		Status:       GetStatus(string(torrent.State)),
		Ratio:        torrent.Ratio,
		Progress:     torrent.Progress,
		Category:     torrent.Category,
		Tags:         strings.Split(torrent.Tags, ","),
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

func FromTrans(torrent transmissionrpc.Torrent) Torrent {
	return Torrent{
		AddedTime:    *torrent.AddedDate,
		LastActivity: utils.IfOr(torrent.ActivityDate.Unix() == 0, time.Time{}, *torrent.ActivityDate),
		TimeElapsed:  time.Now().Sub(*torrent.AddedDate),
		SeedingTime:  *torrent.TimeSeeding,
		ID:           *torrent.ID,
		Hash:         *torrent.HashString,
		Name:         *torrent.Name,
		Status:       TRStatusToQbStatus(*torrent.Status),
		Ratio:        *torrent.UploadRatio,
		Progress:     *torrent.PercentDone,
		Tags:         torrent.Labels,
		Size:         int64(torrent.TotalSize.Byte()),
		Leecher: utils.Reduce(func(sum int64, v transmissionrpc.TrackerStats) int64 {
			return sum + v.LeecherCount
		}, 0, slices.Values(torrent.TrackerStats)),
		Seeder: utils.Reduce(func(sum int64, v transmissionrpc.TrackerStats) int64 {
			return sum + v.SeederCount
		}, 0, slices.Values(torrent.TrackerStats)),
		DlSpeed:    *torrent.RateDownload,
		UpSpeed:    *torrent.RateUpload,
		AvgDlSpeed: utils.SafeDivide(*torrent.DownloadedEver, int64(torrent.TimeDownloading.Seconds())),
		AvgUpSpeed: utils.SafeDivide(*torrent.UploadedEver, int64(torrent.TimeSeeding.Seconds())),
		Downloaded: *torrent.DownloadedEver,
		Uploaded:   *torrent.UploadedEver,
		Trackers: slices.Collect(utils.IterMap(slices.Values(torrent.Trackers),
			func(tt transmissionrpc.Tracker) string {
				return tt.Announce
			})),
	}
}

func FilterTorrents(f Filter, torrents []Torrent) []Torrent {
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
