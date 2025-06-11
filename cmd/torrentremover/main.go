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
	_ "time/tzdata"

	"github.com/robfig/cron/v3"
	"github.com/urfave/cli/v3"

	"github.com/swkisdust/torrentremover/internal/client"
	"github.com/swkisdust/torrentremover/internal/client/qbitorrentx"
	"github.com/swkisdust/torrentremover/internal/exprx"
	logx "github.com/swkisdust/torrentremover/internal/log"
	"github.com/swkisdust/torrentremover/model"
)

var (
	Version string
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
		Version: Version,
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
			return setupDaemon(config, dryRun)
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

func setupDaemon(c *model.Config, dryRun bool) error {
	if !c.Daemon.Enabled {
		slog.Info("running in oneshot mode")
		return run(c, nil, dryRun)
	}

	slog.Info("running in daemon mode", "cronexp", c.Daemon.CronExp)
	clientMap := parseClients(c)
	cronLogger := logx.NewCronLogger(slog.Default())
	cronScheduler := cron.New(cron.WithLogger(cronLogger), cron.WithChain(
		cron.Recover(cronLogger),
		cron.SkipIfStillRunning(cronLogger),
	))

	_, err := cronScheduler.AddFunc(c.Daemon.CronExp, func() {
		if err := run(c, clientMap, dryRun); err != nil {
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

func parseClients(c *model.Config) map[string]client.Client {
	clientMap := make(map[string]client.Client)

	for name, config := range c.Clients {
		switch config.Type {
		case "qbittorrent":
			clientMap[name] = qbitorrentx.NewQbittorrent(config.Config)
		default:
			slog.Warn("unsupported client type", "client_name", name, "client_type", config.Type)
		}
	}

	return clientMap
}

func run(c *model.Config, clientMap map[string]client.Client, dryRun bool) error {
	if clientMap == nil {
		clientMap = parseClients(c)
	}

	if len(clientMap) == 0 {
		return errors.New("you didn't configure any client")
	}

	if len(c.Profiles) == 0 {
		return errors.New("you didn't configure any profile")
	}

	for _, profile := range c.Profiles {
		client, ok := clientMap[profile.Client]
		if !ok {
			slog.Warn("client not found", "client_id", profile.Client)
			continue
		}

		torrents, err := client.GetTorrents()
		if err != nil {
			slog.Warn("failed to get torrent list", "error", err)
		}
		slog.Debug("available torrents", "value", torrents)
		for _, st := range profile.Strategy {
			expr, err := exprx.Compile(st.RemoveExpr, client)
			if err != nil {
				slog.Warn("failed to compile expr", "strategy", st.Name, "client_id", profile.Client, "error", err)
				continue
			}

			filteredTorrents := model.FilterTorrents(st.Filter, torrents)
			slog.Debug("filtered torrents", "strategy", st.Name, "value", filteredTorrents)
			if err := expr.Run(filteredTorrents, st.Name, dryRun, st.Reannounce, st.DeleteFiles); err != nil {
				slog.Warn("failed to execute expr", "strategy", st.Name, "client_id", profile.Client, "error", err)
			}
		}
	}

	return nil
}
