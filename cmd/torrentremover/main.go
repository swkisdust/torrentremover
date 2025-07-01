package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
	_ "time/tzdata"

	"github.com/robfig/cron/v3"
	"github.com/urfave/cli/v3"

	"github.com/swkisdust/torrentremover/internal/client"
	"github.com/swkisdust/torrentremover/internal/client/delugex"
	"github.com/swkisdust/torrentremover/internal/client/qbitorrentx"
	"github.com/swkisdust/torrentremover/internal/client/transmissionx"
	"github.com/swkisdust/torrentremover/internal/exprx"
	logx "github.com/swkisdust/torrentremover/internal/log"
	"github.com/swkisdust/torrentremover/internal/utils"
	"github.com/swkisdust/torrentremover/model"
)

func initConfig(path string) (*model.Config, error) {
	var config model.Config
	if err := config.Read(path); err != nil {
		return nil, fmt.Errorf("init config: %v", err)
	}

	return &config, nil
}

func loadDefaultConfigPath() string {
	var err error
	executablePath, err := os.Executable()
	if err != nil {
		return ""
	}
	return filepath.Join(filepath.Dir(executablePath), "config.yaml")
}

func main() {
	app := &cli.Command{
		Usage:   "torrentremover",
		Version: model.Version,
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "config", Aliases: []string{"c"}, Usage: "config path"},
			&cli.BoolFlag{Name: "dry-run", Aliases: []string{"n"}, Usage: "dry run"},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			var config *model.Config
			var err error
			if path := c.String("config"); path != "" {
				if config, err = initConfig(path); err != nil {
					return err
				}
			} else {
				if config, err = initConfig(loadDefaultConfigPath()); err != nil {
					return err
				}
			}
			dryRun := c.Bool("dry-run")
			setupLogger(config.Log)
			return setupDaemon(ctx, config, dryRun)
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		slog.Error("application error", "error", err)
		os.Exit(1)
	}
}

func setupLogger(c model.LogConfig) {
	logger := logx.NewLogger(!c.Disabled, c.Level)
	slog.SetDefault(logger)
}

func setupDaemon(ctx context.Context, c *model.Config, dryRun bool) error {
	clientMap := parseClients(ctx, c)
	if len(clientMap) == 0 {
		return errors.New("you didn't configure any client")
	}

	if len(c.Profiles) == 0 {
		return errors.New("you didn't configure any profile")
	}

	if c.Daemon.Disabled {
		slog.Info("running in oneshot mode")
		return run(ctx, c, clientMap, dryRun)
	}

	slog.Info("running in daemon mode", "cronexp", c.Daemon.CronExp)
	cronLogger := logx.NewCronLogger(slog.Default())
	cronScheduler := cron.New(cron.WithLogger(cronLogger), cron.WithSeconds(), cron.WithChain(
		cron.Recover(cronLogger),
		cron.SkipIfStillRunning(cronLogger),
	))

	_, err := cronScheduler.AddFunc(c.Daemon.CronExp, func() {
		if err := run(ctx, c, clientMap, dryRun); err != nil {
			slog.Error("run() error", "error", err)
		}
	})

	if err != nil {
		slog.Error("failed to schedule cron job", "cronexp", c.Daemon.CronExp, "error", err)
		os.Exit(1)
	}

	cronScheduler.Start()
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-stopChan
	slog.Info("received shutdown signal", "signal", sig.String())
	slog.Info("exiting...")

	cronScheduler.Stop().Done()
	return nil
}

func parseClients(ctx context.Context, c *model.Config) map[string]client.Client {
	clientMap := make(map[string]client.Client)

	for name, config := range c.Clients {
		switch config.Type {
		case "qbittorrent":
			if client, err := qbitorrentx.NewQbittorrent(config.Config); err == nil {
				clientMap[name] = client
			} else {
				slog.Warn("failed to create qbittorrent client", "name", name, "config", config.Config, "error", err)
			}
		case "transmission":
			if client, err := transmissionx.NewTransmission(config.Config); err == nil {
				clientMap[name] = client
			} else {
				slog.Warn("failed to create transmission client", "name", name, "config", config.Config, "error", err)
			}
		case "deluge", "deluge_v2":
			if client, err := delugex.NewDeluge(ctx, config.Config); err == nil {
				clientMap[name] = client
			} else {
				slog.Warn("failed to create deluge client", "name", name, "config", config.Config, "error", err)
			}
		default:
			slog.Warn("unsupported client type", "client_name", name, "client_type", config.Type)
		}
	}

	return clientMap
}

func run(ctx context.Context, c *model.Config, clientMap map[string]client.Client, dryRun bool) error {
	for _, profile := range c.Profiles {
		client, ok := clientMap[profile.Client]
		if !ok {
			slog.Error("client not found", "client_id", profile.Client)
			continue
		}

		torrents, err := client.GetTorrents(ctx)
		if err != nil {
			slog.Error("failed to get torrent list", "error", err)
			continue
		}
		slog.Debug("available torrents", "value", torrents)
		for _, st := range profile.Strategy {
			if st.Prog == nil {
				prog, err := exprx.Compile(st.RemoveExpr, client)
				if err != nil {
					slog.Error("failed to compile expr", "strategy", st.Name, "client_id", profile.Client, "error", err)
					continue
				}
				st.Prog = prog
			}

			freeSpace, err := client.GetFreeSpaceOnDisk(ctx, utils.IfOr(st.Mountpath != "", st.Mountpath, profile.Mountpath))
			if err != nil {
				slog.Warn("failed to get free space on disk", "strategy", st.Name, "client_id", profile.Client, "error", err)
			}
			stats, err := client.SessionStats(ctx)
			if err != nil {
				slog.Warn("failed to get session stats", "strategy", st.Name, "client_id", profile.Client, "error", err)
			}
			filteredTorrents := model.FilterTorrents(st.Filter, freeSpace, torrents)
			if len(filteredTorrents) < 1 {
				slog.Debug("no matching torrents found", "strategy", st.Name)
				continue
			}

			expr := exprx.New(st.Prog, client)
			if err := expr.Run(ctx, filteredTorrents, st.Name, exprx.RunOptions{
				DryRun:       dryRun,
				Reannounce:   profile.Reannounce || st.Reannounce,
				DeleteFiles:  profile.DeleteFiles || st.DeleteFiles,
				Interval:     utils.IfOr(st.DeleteDelay != 0, time.Duration(st.DeleteDelay)*time.Second, time.Duration(profile.DeleteDelay)*time.Second),
				Disk:         int64(freeSpace),
				Limit:        st.Limit,
				Action:       st.Action,
				SessionStats: stats,
			}); err != nil {
				slog.Error("failed to execute expr", "strategy", st.Name, "client_id", profile.Client, "error", err)
			}
		}
	}

	return nil
}
