package delugex

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"net"
	"slices"
	"strconv"
	"time"

	"github.com/autobrr/go-deluge"
	"github.com/go-viper/mapstructure/v2"

	"github.com/swkisdust/torrentremover/internal/utils"
	"github.com/swkisdust/torrentremover/model"
)

type delugeClient interface {
	deluge.DelugeClient
	LabelPlugin(ctx context.Context) (*deluge.LabelPlugin, error)
}

type Deluge struct {
	Host     string `mapstructure:"host"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	V2       bool   `mapstructure:"v2"`

	lp     *deluge.LabelPlugin
	client delugeClient
}

func NewDeluge(ctx context.Context, config map[string]any) (*Deluge, error) {
	var d Deluge
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		WeaklyTypedInput:     true,
		IgnoreUntaggedFields: true,
		Result:               &d,
	})
	if err != nil {
		return nil, err
	}

	if err := decoder.Decode(config); err != nil {
		return nil, err
	}

	host, port, err := net.SplitHostPort(d.Host)
	if err != nil {
		return nil, fmt.Errorf("net.SplitHostPort: %v", err)
	}

	portInt, err := strconv.ParseUint(port, 10, 16)
	if err != nil {
		return nil, fmt.Errorf("cannot parse port %s to uint: %v", port, err)
	}

	settings := deluge.Settings{
		Hostname:         host,
		Port:             uint(portInt),
		Login:            d.Username,
		Password:         d.Password,
		ReadWriteTimeout: time.Second * 5,
	}

	var client delugeClient
	if !d.V2 {
		client = deluge.NewV1(settings)
	} else {
		client = deluge.NewV2(settings)
	}

	if lp, err := client.LabelPlugin(ctx); err == nil {
		d.lp = lp
	} else {
		slog.Warn("failed to retrieve deluge label plugin", "error", err)
	}

	if err := client.Connect(ctx); err != nil {
		return nil, err
	}

	return &d, nil
}

func (d *Deluge) GetTorrents(ctx context.Context) ([]*model.Torrent, error) {
	torrents, err := d.client.TorrentsStatus(ctx, "", nil)
	if err != nil {
		return nil, err
	}

	var labels map[string]string
	if d.lp != nil {
		if labels, err = d.lp.GetTorrentsLabels(deluge.StateUnspecified, nil); err != nil {
			return nil, err
		}
	} else {
		labels = make(map[string]string)
	}

	return slices.Collect(utils.Seq2To1(maps.All(torrents),
		func(id string, ds *deluge.TorrentStatus) *model.Torrent {
			return model.FromDeluge(ds, labels[id])
		})), nil
}

func (d *Deluge) PauseTorrents(ctx context.Context, torrents []*model.Torrent) error {
	hashes := utils.SlicesMap(torrents, func(t *model.Torrent) string {
		return t.Hash
	})

	return d.client.PauseTorrents(ctx, hashes...)
}

func (d *Deluge) ResumeTorrents(ctx context.Context, torrents []*model.Torrent) error {
	hashes := utils.SlicesMap(torrents, func(t *model.Torrent) string {
		return t.Hash
	})

	return d.client.ResumeTorrents(ctx, hashes...)
}

func (d *Deluge) ThrottleTorrents(ctx context.Context, torrents []*model.Torrent, limit model.Bytes) error {
	hashes := utils.SlicesMap(torrents, func(t *model.Torrent) string {
		return t.Hash
	})

	var uploadSpeed int
	if limit == -1 {
		uploadSpeed = -1
	} else {
		uploadSpeed = int(limit.KiB())
	}

	opts := deluge.Options{
		MaxUploadSpeed: &uploadSpeed,
	}

	var wrapErr error
	for _, id := range hashes {
		if err := d.client.SetTorrentOptions(ctx, id, &opts); err != nil {
			wrapErr = errors.Join(wrapErr, err)
			continue
		}
	}

	return wrapErr
}

func (d *Deluge) DeleteTorrents(ctx context.Context, torrents []*model.Torrent, name string, reannounce, deleteFiles bool, interval time.Duration) error {
	hashes := utils.SlicesMap(torrents, func(t *model.Torrent) string {
		return t.Hash
	})

	if reannounce {
		slog.Debug("pausing torrents", "strategy", name)
		if err := d.client.PauseTorrents(ctx, hashes...); err != nil {
			return err
		}

		// Waiting to pause torrents
		time.Sleep(time.Second * 2)

		slog.Debug("resuming torrents", "strategy", name)
		if err := d.client.ResumeTorrents(ctx, hashes...); err != nil {
			return err
		}

		// Waiting to resume torrents
		time.Sleep(time.Second * 2)

		slog.Debug("reannouncing torrents", "strategy", name)
		if err := d.client.ForceReannounce(ctx, hashes); err != nil {
			return err
		}

		// Waiting for reannounce (might not needed)
		time.Sleep(utils.IfOr(interval != 0, interval, time.Second*4))
	}

	_, err := d.client.RemoveTorrents(ctx, hashes, deleteFiles)
	return err
}

func (d *Deluge) GetFreeSpaceOnDisk(ctx context.Context, path string) (model.Bytes, error) {
	size, err := d.client.GetFreeSpace(ctx, path)
	if err != nil {
		return -1, err
	}
	return model.Bytes(size), nil
}

func (d *Deluge) SessionStats(ctx context.Context) (model.SessionStats, error) {
	var stats model.SessionStats
	dstats, err := d.client.GetSessionStatus(ctx)
	if err != nil {
		return stats, err
	}

	stats.TotalDlSpeed = int64(dstats.DownloadRate)
	stats.TotalUpSpeed = int64(dstats.UploadRate)
	return stats, nil
}
