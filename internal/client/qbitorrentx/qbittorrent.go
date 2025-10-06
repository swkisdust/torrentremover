package qbitorrentx

import (
	"context"
	"log/slog"
	"slices"
	"strings"
	"time"

	"github.com/autobrr/go-qbittorrent"
	"github.com/go-viper/mapstructure/v2"

	"github.com/swkisdust/torrentremover/internal/utils"
	"github.com/swkisdust/torrentremover/model"
)

type Qbitorrent struct {
	Host        string `mapstructure:"host"`
	Username    string `mapstructure:"username"`
	Password    string `mapstructure:"password"`
	BasicUser   string `mapstructure:"basic_user"`
	BasicPass   string `mapstructure:"basic_pass"`
	InsecureTLS bool   `mapstructure:"insecure_tls"`

	client *qbittorrent.Client
}

func NewQbittorrent(config map[string]any) (*Qbitorrent, error) {
	var qb Qbitorrent
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		WeaklyTypedInput:     true,
		IgnoreUntaggedFields: true,
		Result:               &qb,
	})
	if err != nil {
		return nil, err
	}

	if err := decoder.Decode(config); err != nil {
		return nil, err
	}

	qb.client = qbittorrent.NewClient(qbittorrent.Config{
		Host:          qb.Host,
		Username:      qb.Username,
		Password:      qb.Password,
		BasicUser:     qb.BasicUser,
		BasicPass:     qb.BasicPass,
		TLSSkipVerify: qb.InsecureTLS,
	})

	return &qb, nil
}

func (qb *Qbitorrent) GetTorrents(ctx context.Context) ([]*model.Torrent, error) {
	torrents, err := qb.client.GetTorrentsCtx(ctx, qbittorrent.TorrentFilterOptions{IncludeTrackers: true})
	if err != nil {
		return nil, err
	}

	return utils.SlicesMap(torrents,
		func(qt qbittorrent.Torrent) *model.Torrent {
			prop, _ := qb.client.GetTorrentPropertiesCtx(ctx, qt.Hash)
			if len(qt.Trackers) == 0 {
				trackers, _ := qb.client.GetTorrentTrackersCtx(ctx, qt.Hash)
				qt.Trackers = slices.DeleteFunc(trackers, func(tt qbittorrent.TorrentTracker) bool {
					return strings.Contains(tt.Url, "[DHT]") || strings.Contains(tt.Url, "[PeX]") || strings.Contains(tt.Url, "[LSD]")
				})
			}
			return model.FromQbit(&qt, &prop)
		}), nil
}

func (qb *Qbitorrent) PauseTorrents(ctx context.Context, torrents []*model.Torrent) error {
	hashes := utils.SlicesMap(torrents,
		func(t *model.Torrent) string {
			return t.Hash
		})

	return qb.client.PauseCtx(ctx, hashes)
}

func (qb *Qbitorrent) ResumeTorrents(ctx context.Context, torrents []*model.Torrent) error {
	hashes := utils.SlicesMap(torrents,
		func(t *model.Torrent) string {
			return t.Hash
		})

	return qb.client.ResumeCtx(ctx, hashes)
}

func (qb *Qbitorrent) ThrottleTorrents(ctx context.Context, torrents []*model.Torrent, limit model.Bytes) error {
	hashes := utils.SlicesMap(torrents,
		func(t *model.Torrent) string {
			return t.Hash
		})

	return qb.client.SetTorrentUploadLimitCtx(ctx, hashes, int64(limit))
}

func (qb *Qbitorrent) DeleteTorrents(ctx context.Context, torrents []*model.Torrent, name string, reannounce, deleteFiles bool, interval time.Duration) error {
	hashes := utils.SlicesMap(torrents,
		func(t *model.Torrent) string {
			return t.Hash
		})

	if reannounce {
		slog.Debug("pausing torrents", "strategy", name)
		if err := qb.client.PauseCtx(ctx, hashes); err != nil {
			return err
		}

		// Waiting to pause torrents
		time.Sleep(time.Second * 2)

		slog.Debug("resuming torrents", "strategy", name)
		if err := qb.client.ResumeCtx(ctx, hashes); err != nil {
			return err
		}

		// Waiting to resume torrents
		time.Sleep(time.Second * 2)

		slog.Debug("reannouncing torrents", "strategy", name)
		if err := qb.client.ReAnnounceTorrentsCtx(ctx, hashes); err != nil {
			return err
		}

		// Waiting for reannounce (might not needed)
		time.Sleep(utils.IfOr(interval != 0, interval, time.Second*4))
	}

	return qb.client.DeleteTorrents(utils.SlicesMap(torrents,
		func(t *model.Torrent) string {
			return t.Hash
		}), deleteFiles)
}

func (qb *Qbitorrent) GetFreeSpaceOnDisk(ctx context.Context, path string) (model.Bytes, error) {
	val, err := qb.client.GetFreeSpaceOnDiskCtx(ctx)
	if err != nil {
		return -1, err
	}
	return model.Bytes(val), nil
}

func (qb *Qbitorrent) SessionStats(ctx context.Context) (model.SessionStats, error) {
	var stats model.SessionStats
	maindata, err := qb.client.SyncMainDataCtx(ctx, 0)
	if err != nil {
		return stats, err
	}

	stats.TotalDlSpeed = maindata.ServerState.DlInfoSpeed
	stats.TotalUpSpeed = maindata.ServerState.UpInfoSpeed
	return stats, nil
}
