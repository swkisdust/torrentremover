package delugex

import (
	"context"
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

type Deluge struct {
	Host     string `mapstructure:"host"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	V2       bool   `mapstructure:"v2"`

	client deluge.DelugeClient
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

	var client deluge.DelugeClient
	if !d.V2 {
		client = deluge.NewV1(settings)
	} else {
		client = deluge.NewV2(settings)
	}

	if err := client.Connect(ctx); err != nil {
		return nil, err
	}

	return &d, nil
}

func (d *Deluge) GetTorrents(ctx context.Context) ([]model.Torrent, error) {
	torrents, err := d.client.TorrentsStatus(ctx, "", nil)
	if err != nil {
		return nil, err
	}

	return slices.Collect(utils.Seq2To1(maps.All(torrents),
		func(id string, ds *deluge.TorrentStatus) model.Torrent {
			return model.FromDeluge(id, ds)
		})), nil
}

func (d *Deluge) DeleteTorrents(ctx context.Context, torrents []model.Torrent, name string, reannounce, deleteFiles bool, interval time.Duration) error {
	hashes := slices.Collect(utils.IterMap(slices.Values(torrents), func(t model.Torrent) string {
		return t.Hash
	}))

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

	if reannounce {
		slog.Debug("reannouncing torrents", "strategy", name)
		if err := d.client.ForceReannounce(ctx, hashes); err != nil {
			return err
		}

		// Waiting for reannounce (might not needed)
		time.Sleep(utils.IfOr(interval != 0, interval, time.Second*10))
	}

	_, err := d.client.RemoveTorrents(ctx, hashes, deleteFiles)
	return err
}

func (d *Deluge) GetFreeSpaceOnDisk(ctx context.Context, path string) model.Bytes {
	size, err := d.client.GetFreeSpace(ctx, path)
	if err != nil {
		return -1
	}
	return model.Bytes(size)
}
