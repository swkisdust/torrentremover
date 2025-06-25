package exprx

import (
	"cmp"
	"context"
	"fmt"
	"log/slog"
	"reflect"
	"time"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"

	"github.com/swkisdust/torrentremover/internal/client"
	"github.com/swkisdust/torrentremover/internal/utils"
	"github.com/swkisdust/torrentremover/model"
)

type RemoveExpr struct {
	prog *vm.Program
	c    client.Client
}

type env struct {
	Torrents []model.Torrent               `expr:"torrents"`
	Bytes    func(s string) (int64, error) `expr:"bytes"`
	Cmp      func(a, b int64) int          `expr:"cmp"`
	CmpFloat func(a, b float64) int        `expr:"cmpFloat"`
}

func Compile(raw string, client client.Client) (*RemoveExpr, error) {
	prog, err := expr.Compile(raw, expr.Env(env{}), expr.AsKind(reflect.Slice))
	if err != nil {
		return nil, err
	}

	return &RemoveExpr{prog, client}, nil
}

func (r *RemoveExpr) Run(ctx context.Context, torrents []model.Torrent, name string, raInt time.Duration, dryRun, reannounce, deleteFiles bool) error {
	env := env{torrents, utils.ParseBytes, cmp.Compare[int64], cmp.Compare[float64]}
	fti, err := expr.Run(r.prog, env)
	if err != nil {
		return err
	}

	rawFt, ok := fti.([]any)
	if !ok {
		return fmt.Errorf("expr returned an unexpected type: %T, expected []any", fti)
	}

	ft := make([]model.Torrent, 0, len(rawFt))
	for _, item := range rawFt {
		t, ok := item.(model.Torrent)
		if !ok {
			return fmt.Errorf("element in filtered list is not model.Torrent, got %T, value %v", item, item)
		}
		ft = append(ft, t)
	}

	for _, t := range ft {
		slog.Info("found deletable torrent",
			"strategy", name,
			"hash", t.Hash,
			"name", t.Name,
			"progress", t.Progress,
			"status", t.Status,
			"size", t.Size,
			"ratio", t.Ratio,
			"added_time", t.AddedTime,
			"last_activity", t.LastActivity,
			"time_elapsed", t.TimeElapsed,
			"seeding_time", t.SeedingTime,
			"category", t.Category,
			"tags", t.Tags,
			"trackers", t.Trackers,
		)
	}

	if dryRun {
		slog.Debug("dry-run ended", "strategy", name)
		return nil
	}

	if len(ft) < 1 {
		slog.Debug("no matching torrents found", "strategy", name)
		return nil
	}

	if err := r.c.DeleteTorrents(ctx, ft, name, reannounce, deleteFiles, raInt); err != nil {
		return fmt.Errorf("c.DeleteTorrents: %v", err)
	}

	slog.Info("torrents deleted", "strategy", name, "deleteFiles", deleteFiles)
	return nil
}
