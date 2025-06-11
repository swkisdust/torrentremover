package qbitorrentx

import (
	"slices"

	"github.com/autobrr/go-qbittorrent"

	"github.com/swkisdust/torrentremover/internal/utils"
	"github.com/swkisdust/torrentremover/model"
)

type Qbitorrent struct {
	client *qbittorrent.Client
}

func NewQbittorrent(config model.ClientConfig) *Qbitorrent {
	client := qbittorrent.NewClient(qbittorrent.Config{
		Host:          config.Host,
		Username:      config.Username,
		Password:      config.Password,
		BasicUser:     config.BasicUser,
		BasicPass:     config.BasicPass,
		TLSSkipVerify: config.InsecureTLS,
	})

	return &Qbitorrent{client}
}

func (qb *Qbitorrent) GetTorrents() ([]model.Torrent, error) {
	torrents, err := qb.client.GetTorrents(qbittorrent.TorrentFilterOptions{IncludeTrackers: true})
	if err != nil {
		return nil, err
	}

	return slices.Collect(utils.IterMap(slices.Values(torrents),
		func(qt qbittorrent.Torrent) model.Torrent {
			prop, _ := qb.client.GetTorrentProperties(qt.Hash)
			if len(qt.Trackers) == 0 {
				qt.Trackers, _ = qb.client.GetTorrentTrackers(qt.Hash)
			}
			return model.FromQbit(qt, prop)
		})), nil
}

func (qb *Qbitorrent) DeleteTorrents(torrents []model.Torrent, deleteFiles bool) error {
	return qb.client.DeleteTorrents(slices.Collect(utils.IterMap(slices.Values(torrents),
		func(t model.Torrent) string {
			return t.Hash
		})), deleteFiles)
}

func (qb *Qbitorrent) Reannounce(torrents []model.Torrent) error {
	return qb.client.ReAnnounceTorrents(slices.Collect(utils.IterMap(slices.Values(torrents),
		func(t model.Torrent) string {
			return t.Hash
		})))
}

func (qb *Qbitorrent) GetFreeSpaceOnDisk() int64 {
	val, err := qb.client.GetFreeSpaceOnDisk()
	if err != nil {
		return -1
	}
	return val
}
