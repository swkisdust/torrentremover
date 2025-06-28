package client

import (
	"context"
	"time"

	"github.com/swkisdust/torrentremover/model"
)

type Client interface {
	GetTorrents(ctx context.Context) ([]model.Torrent, error)
	PauseTorrents(ctx context.Context, torrents []model.Torrent) error
	ResumeTorrents(ctx context.Context, torrents []model.Torrent) error
	ThrottleTorrents(ctx context.Context, torrents []model.Torrent, limit model.Bytes) error
	DeleteTorrents(ctx context.Context, torrents []model.Torrent, name string, reannounce, deleteFiles bool, interval time.Duration) error
	GetFreeSpaceOnDisk(ctx context.Context, path string) (model.Bytes, error)
	SessionStats(ctx context.Context) (model.SessionStats, error)
}
