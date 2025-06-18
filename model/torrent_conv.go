package model

import (
	"slices"
	"strings"
	"time"

	"github.com/autobrr/go-deluge"
	"github.com/autobrr/go-qbittorrent"
	"github.com/hekmon/transmissionrpc/v3"

	"github.com/swkisdust/torrentremover/internal/utils"
)

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
		Leecher:      int64(prop.PeersTotal),
		Seeder:       int64(prop.SeedsTotal),
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
		TimeElapsed:  time.Since(*torrent.AddedDate),
		SeedingTime:  *torrent.TimeSeeding,
		ID:           *torrent.ID,
		Hash:         *torrent.HashString,
		Name:         *torrent.Name,
		Status:       FromTrStatus(*torrent.Status),
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

func FromDeluge(id string, ts *deluge.TorrentStatus) Torrent {
	addedTime := time.Unix(int64(ts.TimeAdded), 0)

	return Torrent{
		AddedTime:    addedTime,
		LastActivity: utils.IfOr(ts.LastSeenComplete == 0, time.Time{}, time.Unix(ts.LastSeenComplete, 0)),
		TimeElapsed:  time.Since(addedTime),
		SeedingTime:  time.Duration(ts.SeedingTime) * time.Second,
		ID:           id,
		Hash:         ts.Hash,
		Name:         ts.Name,
		Status:       GetStatus(ts.State),
		Ratio:        float64(ts.Ratio),
		Progress:     float64(ts.Progress),
		Size:         ts.TotalSize,
		Leecher:      ts.TotalPeers,
		Seeder:       ts.TotalSeeds,
		DlSpeed:      ts.DownloadPayloadRate,
		UpSpeed:      ts.UploadPayloadRate,
		AvgUpSpeed:   utils.SafeDivide(ts.TotalUploaded, ts.ActiveTime),
		AvgDlSpeed:   utils.SafeDivide(ts.AllTimeDownload, (ts.ActiveTime - ts.CompletedTime)),
		Downloaded:   ts.AllTimeDownload,
		Uploaded:     ts.TotalUploaded,
		Trackers:     []string{ts.TrackerHost},
	}
}
