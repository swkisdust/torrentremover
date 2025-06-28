package transmissionx

import (
	"context"
	"log/slog"
	"net/url"
	"slices"
	"time"

	"github.com/go-viper/mapstructure/v2"
	"github.com/hekmon/transmissionrpc/v3"

	"github.com/swkisdust/torrentremover/internal/utils"
	"github.com/swkisdust/torrentremover/model"
)

type Transmission struct {
	Host     string `mapstructure:"host"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`

	client *transmissionrpc.Client
}

func NewTransmission(config map[string]any) (*Transmission, error) {
	var tr Transmission
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		WeaklyTypedInput:     true,
		IgnoreUntaggedFields: true,
		Result:               &tr,
	})
	if err != nil {
		return nil, err
	}

	if err := decoder.Decode(config); err != nil {
		return nil, err
	}

	endpoint, err := url.Parse(tr.Host)
	if err != nil {
		return nil, err
	}
	endpoint.User = url.UserPassword(tr.Username, tr.Password)

	tr.client, err = transmissionrpc.New(endpoint, nil)
	if err != nil {
		return nil, err
	}

	return &tr, nil
}

func (tr *Transmission) GetTorrents(ctx context.Context) ([]model.Torrent, error) {
	torrents, err := tr.client.TorrentGetAll(ctx)
	if err != nil {
		return nil, err
	}

	return slices.Collect(utils.IterMap(slices.Values(torrents),
		func(tt transmissionrpc.Torrent) model.Torrent {
			return model.FromTrans(tt)
		})), nil
}

func (tr *Transmission) PauseTorrents(ctx context.Context, torrents []model.Torrent) error {
	ids := slices.Collect(utils.IterMap(slices.Values(torrents),
		func(t model.Torrent) int64 {
			return t.ID.(int64)
		}))

	return tr.client.TorrentStopIDs(ctx, ids)
}

func (tr *Transmission) ResumeTorrents(ctx context.Context, torrents []model.Torrent) error {
	ids := slices.Collect(utils.IterMap(slices.Values(torrents),
		func(t model.Torrent) int64 {
			return t.ID.(int64)
		}))

	return tr.client.TorrentStartIDs(ctx, ids)
}

func (tr *Transmission) ThrottleTorrents(ctx context.Context, torrents []model.Torrent, limit model.Bytes) error {
	ids := slices.Collect(utils.IterMap(slices.Values(torrents),
		func(t model.Torrent) int64 {
			return t.ID.(int64)
		}))

	var uploadSpeed int64
	if limit == -1 {
		uploadSpeed = -1
	} else {
		uploadSpeed = limit.KB()
	}
	limited := utils.IfOr(uploadSpeed < 1, false, true)

	return tr.client.TorrentSet(ctx, transmissionrpc.TorrentSetPayload{
		IDs:           ids,
		UploadLimit:   &uploadSpeed,
		UploadLimited: &limited,
	})
}

func (tr *Transmission) DeleteTorrents(ctx context.Context, torrents []model.Torrent, name string, reannounce, deleteFiles bool, interval time.Duration) error {
	ids := slices.Collect(utils.IterMap(slices.Values(torrents),
		func(t model.Torrent) int64 {
			return t.ID.(int64)
		}))

	if reannounce {
		slog.Debug("pausing torrents", "strategy", name)
		if err := tr.client.TorrentStopIDs(ctx, ids); err != nil {
			return err
		}

		// Waiting to pause torrents
		time.Sleep(time.Second * 2)

		slog.Debug("resuming torrents", "strategy", name)
		if err := tr.client.TorrentStartIDs(ctx, ids); err != nil {
			return err
		}

		// Waiting to resume torrents
		time.Sleep(time.Second * 2)

		slog.Debug("reannouncing torrents", "strategy", name)
		if err := tr.client.TorrentReannounceIDs(ctx, ids); err != nil {
			return err
		}

		// Waiting for reannounce (might not needed)
		time.Sleep(utils.IfOr(interval != 0, interval, time.Second*10))
	}

	return tr.client.TorrentRemove(ctx, transmissionrpc.TorrentRemovePayload{
		IDs:             ids,
		DeleteLocalData: deleteFiles,
	})
}

func (tr *Transmission) GetFreeSpaceOnDisk(ctx context.Context, path string) (model.Bytes, error) {
	free, _, err := tr.client.FreeSpace(ctx, path)
	if err != nil {
		return -1, err
	}
	return model.Bytes(free.Byte()), nil
}

func (tr *Transmission) SessionStats(ctx context.Context) (model.SessionStats, error) {
	var stats model.SessionStats
	tstats, err := tr.client.SessionStats(ctx)
	if err != nil {
		return stats, err
	}

	stats.TotalDlSpeed = tstats.DownloadSpeed
	stats.TotalUpSpeed = tstats.UploadSpeed
	return stats, nil
}
