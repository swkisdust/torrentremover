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

func (qb *Qbitorrent) GetTorrents(ctx context.Context) ([]model.Torrent, error) {
	torrents, err := qb.client.GetTorrentsCtx(ctx, qbittorrent.TorrentFilterOptions{IncludeTrackers: true})
	if err != nil {
		return nil, err
	}

	return slices.Collect(utils.IterMap(slices.Values(torrents),
		func(qt qbittorrent.Torrent) model.Torrent {
			prop, _ := qb.client.GetTorrentPropertiesCtx(ctx, qt.Hash)
			if len(qt.Trackers) == 0 {
				trackers, _ := qb.client.GetTorrentTrackersCtx(ctx, qt.Hash)
				qt.Trackers = slices.DeleteFunc(trackers, func(tt qbittorrent.TorrentTracker) bool {
					return strings.Contains(tt.Url, "[DHT]") || strings.Contains(tt.Url, "[PeX]") || strings.Contains(tt.Url, "[LSD]")
				})
			}
			return model.FromQbit(qt, prop)
		})), nil
}

func (qb *Qbitorrent) DeleteTorrents(ctx context.Context, torrents []model.Torrent, name string, reannounce, deleteFiles bool, interval time.Duration) error {
	hashes := slices.Collect(utils.IterMap(slices.Values(torrents),
		func(t model.Torrent) string {
			return t.Hash
		}))

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

	if reannounce {
		slog.Debug("reannouncing torrents", "strategy", name)
		if err := qb.client.ReAnnounceTorrentsCtx(ctx, hashes); err != nil {
			return err
		}

		// Waiting for reannounce (might not needed)
		time.Sleep(utils.IfOr(interval != 0, interval, time.Second*10))
	}

	return qb.client.DeleteTorrents(slices.Collect(utils.IterMap(slices.Values(torrents),
		func(t model.Torrent) string {
			return t.Hash
		})), deleteFiles)
}

func (qb *Qbitorrent) GetFreeSpaceOnDisk(ctx context.Context, path string) model.Bytes {
	val, err := qb.client.GetFreeSpaceOnDiskCtx(ctx)
	if err != nil {
		return -1
	}
	return model.Bytes(val)
}
