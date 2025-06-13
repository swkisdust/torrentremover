package transmissionx

import (
	"context"
	"net/url"
	"slices"

	"github.com/hekmon/transmissionrpc/v3"

	"github.com/swkisdust/torrentremover/internal/utils"
	"github.com/swkisdust/torrentremover/model"
)

type Transmission struct {
	client *transmissionrpc.Client
}

func NewTransmission(config model.ClientConfig) (*Transmission, error) {
	endpoint, err := url.Parse(config.Host)
	if err != nil {
		return nil, err
	}
	endpoint.User = url.UserPassword(config.Username, config.Password)

	client, err := transmissionrpc.New(endpoint, nil)
	if err != nil {
		return nil, err
	}

	return &Transmission{client}, nil
}

func (tr *Transmission) GetTorrents() ([]model.Torrent, error) {
	torrents, err := tr.client.TorrentGetAll(context.Background())
	if err != nil {
		return nil, err
	}

	return slices.Collect(utils.IterMap(slices.Values(torrents),
		func(tt transmissionrpc.Torrent) model.Torrent {
			return model.FromTrans(tt)
		})), nil
}

func (tr *Transmission) DeleteTorrents(torrents []model.Torrent, deleteFiles bool) error {
	return tr.client.TorrentRemove(context.Background(), transmissionrpc.TorrentRemovePayload{
		IDs: slices.Collect(utils.IterMap(slices.Values(torrents),
			func(t model.Torrent) int64 {
				return t.ID.(int64)
			})),
		DeleteLocalData: deleteFiles,
	})
}

func (tr *Transmission) Reannounce(torrents []model.Torrent) error {
	return tr.client.TorrentReannounceHashes(context.Background(), slices.Collect(utils.IterMap(slices.Values(torrents),
		func(t model.Torrent) string {
			return t.Hash
		})))
}

func (tr *Transmission) GetFreeSpaceOnDisk(path string) int64 {
	free, _, err := tr.client.FreeSpace(context.Background(), path)
	if err != nil {
		return -1
	}
	return int64(free.Byte())
}
