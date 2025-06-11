package client

import "github.com/swkisdust/torrentremover/model"

type Client interface {
	GetTorrents() ([]model.Torrent, error)
	DeleteTorrents(torrents []model.Torrent, deleteFiles bool) error
	Reannounce(torrents []model.Torrent) error
	GetFreeSpaceOnDisk() int64
}
