package exprx

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/swkisdust/torrentremover/model"
)

type mockClient struct {
	t        *testing.T
	expected []model.Torrent
}

var testCases = []model.Torrent{
	{
		Hash:         "test1",
		Name:         "test1",
		Ratio:        1.234,
		Progress:     34.421,
		Size:         10240000,
		UpSpeed:      1048576,
		SeedingTime:  time.Hour * 2,
		LastActivity: time.Now().Add(-time.Hour * 2),
		Leecher:      12,
		Seeder:       128,
	},
	{
		Hash:         "test2",
		Name:         "test2",
		Ratio:        2.50,
		Progress:     100.00,
		Size:         5368709120,
		UpSpeed:      20971520,
		SeedingTime:  time.Hour * 24 * 7,
		LastActivity: time.Now().Add(-time.Minute * 30),
		Leecher:      52,
		Seeder:       1,
	},
	{
		Hash:         "test3",
		Name:         "test3",
		Ratio:        0.85,
		Progress:     75.12,
		Size:         2147483648,
		UpSpeed:      524288,
		SeedingTime:  time.Hour * 3,
		LastActivity: time.Now().Add(-time.Hour * 5),
		Leecher:      30,
		Seeder:       20,
	},
}

func (c *mockClient) GetTorrents(ctx context.Context) ([]model.Torrent, error) {
	return testCases, nil
}

func (c *mockClient) PauseTorrents(ctx context.Context, torrents []model.Torrent) error {
	c.t.Logf("received torrents %v", torrents)
	return nil
}

func (c *mockClient) ResumeTorrents(ctx context.Context, torrents []model.Torrent) error {
	c.t.Logf("received torrents %v", torrents)
	return nil
}

func (c *mockClient) ThrottleTorrents(ctx context.Context, torrents []model.Torrent, limit model.Bytes) error {
	c.t.Logf("received torrents %v", torrents)
	return nil
}

func (c *mockClient) DeleteTorrents(ctx context.Context, torrents []model.Torrent, name string, reannounce, deleteFiles bool, interval time.Duration) error {
	c.t.Logf("received torrents %v", torrents)
	if !reflect.DeepEqual(c.expected, torrents) {
		c.t.Errorf("excepted %v, got %v", c.expected, torrents)
	}
	return nil
}

func (c *mockClient) GetFreeSpaceOnDisk(ctx context.Context, path string) (model.Bytes, error) {
	return 2 * 1024 * 1024 * 1024, nil
}

func (c *mockClient) SessionStats(ctx context.Context) (model.SessionStats, error) {
	return model.SessionStats{}, nil
}

func TestRemoveExpr(t *testing.T) {
	t.Run("SimpleExpr", func(t *testing.T) {
		const exprStr = `filter(torrents, .size > 10240000 && .seeding_time > duration("1h"))`
		client := &mockClient{t, testCases[1:3]}

		expr, err := Compile(exprStr, client)
		if err != nil {
			t.Errorf("failed to compile expr: %v", err)
		}

		if err := expr.Run(context.Background(), testCases, "testSt", RunOptions{
			Reannounce:  true,
			DeleteFiles: true,
		}); err != nil {
			t.Errorf("failed to execute expr: %v", err)
		}
	})

	t.Run("ArithmeticExpr", func(t *testing.T) {
		const exprStr = `filter(torrents, .seeder / .leecher < 1 && now() - .last_activity > duration("1h"))`
		client := &mockClient{t, testCases[2:3]}

		expr, err := Compile(exprStr, client)
		if err != nil {
			t.Errorf("failed to compile expr: %v", err)
		}

		if err := expr.Run(context.Background(), testCases, "testSt", RunOptions{
			Reannounce:  true,
			DeleteFiles: true,
		}); err != nil {
			t.Errorf("failed to execute expr: %v", err)
		}
	})
}
