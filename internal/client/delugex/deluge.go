package delugex

import (
	"context"
	"fmt"
	"maps"
	"net"
	"slices"
	"strconv"
	"time"

	"github.com/autobrr/go-deluge"

	"github.com/swkisdust/torrentremover/internal/utils"
	"github.com/swkisdust/torrentremover/model"
)

type Deluge struct {
	client deluge.DelugeClient
}

func NewDeluge(config model.ClientConfig) (*Deluge, error) {
	host, port, err := net.SplitHostPort(config.Host)
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
		Login:            config.Username,
		Password:         config.Password,
		ReadWriteTimeout: time.Second * 5,
	}

	clientV2 := deluge.NewV2(settings)
	if err := clientV2.Connect(context.Background()); err != nil {
		return nil, err
	}

	return &Deluge{clientV2}, nil
}

func (d *Deluge) GetTorrents() ([]model.Torrent, error) {
	torrents, err := d.client.TorrentsStatus(context.Background(), "", nil)
	if err != nil {
		return nil, err
	}

	return slices.Collect(utils.Seq2To1(maps.All(torrents),
		func(id string, ds *deluge.TorrentStatus) model.Torrent {
			return model.FromDeluge(id, ds)
		})), nil
}

func (d *Deluge) DeleteTorrents(torrents []model.Torrent, deleteFiles bool) error {
	_, err := d.client.RemoveTorrents(context.Background(),
		slices.Collect(utils.IterMap(slices.Values(torrents), func(t model.Torrent) string {
			return t.ID.(string)
		})), deleteFiles)

	return err
}

func (d *Deluge) Reannounce(torrents []model.Torrent) error {
	return d.client.ForceReannounce(context.Background(),
		slices.Collect(utils.IterMap(slices.Values(torrents), func(t model.Torrent) string {
			return t.ID.(string)
		})))
}

func (d *Deluge) GetFreeSpaceOnDisk(path string) model.Bytes {
	size, err := d.client.GetFreeSpace(context.Background(), path)
	if err != nil {
		return -1
	}
	return model.Bytes(size)
}
