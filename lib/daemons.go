package lib

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

type Daemon struct {
	// a path-safe name with no spaces
	Id          string
	Description string
	ExecStart   string
}

const systemdTemplate = `[Unit]
Description=%[1]s

[Service]
Type=simple
TimeoutStartSec=0
ExecStart=%[2]s
WorkingDirectory=%[3]s

[Install]
WantedBy=default.target`

var systemdDir = filepath.Join(HOME, ".config/systemd/user")
var servicesDir = filepath.Join(HOME, "services")

func createDaemon(daemon Daemon) error {
	wd := filepath.Join(HOME, "services", daemon.Id)

	err := os.Mkdir(wd, 0777)
	if err != nil {
		return err
	}

	daemonFile := fmt.Sprintf(
		systemdTemplate,
		daemon.Description,
		filepath.Join(wd, daemon.ExecStart),
		wd,
	)

	serviceName := fmt.Sprintf("sysman.%s.service", daemon.Id)

	err = WriteFile(
		filepath.Join(
			systemdDir,
			serviceName,
		),
		[]byte(daemonFile),
		0666,
	)
	if err != nil {
		return err
	}

	err = Cmd("systemctl", "--user", "enable", serviceName)
	if err != nil {
		return err
	}
	err = Cmd("systemctl", "--user", "start", serviceName)
	if err != nil {
		return err
	}

	return nil
}

func removeOldDaemons() error {
	entries, err := os.ReadDir(systemdDir)
	if err != nil {
		return err
	}

	for _, e := range entries {
		if !strings.HasPrefix(e.Name(), "sysman") {
			continue
		}

		id, found := strings.CutPrefix(e.Name(), "sysman.")
		if !found {
			continue
		}
		id, found = strings.CutSuffix(e.Name(), ".service")
		if !found {
			continue
		}

		slog.Info("removing...", "daemon", id)

		err = Cmd("systemctl", "--user", "stop", e.Name())
		if err != nil {
			return err
		}
		err = Cmd("systemctl", "--user", "disable", e.Name())
		if err != nil {
			return err
		}

		err = RmFile(filepath.Join(systemdDir, e.Name()))
		if err != nil {
			return err
		}
		rmdir := filepath.Join(servicesDir, id)
		err = RmDir(rmdir)
		if err != nil {
			slog.Warn("remove corresponding service dir", "dir", rmdir, "err", err)
		}
	}

	return nil
}

func SyncDaemons(ctx context.Context, daemons []Daemon) error {
	err := os.Mkdir(servicesDir, 0777)
	if err != nil {
		return fmt.Errorf("mkdir ~/services: %w", err)
	}
	err = os.MkdirAll(systemdDir, 0777)
	if err != nil {
		return fmt.Errorf("mkdir ~/.config/systemd/user: %w", err)
	}

	err = removeOldDaemons()
	if err != nil {
		return fmt.Errorf("remove old daemons: %w", err)
	}

	for _, d := range daemons {
		err = createDaemon(d)
		if err != nil {
			return fmt.Errorf("create daemon (%s): %w", d.Id, err)
		}
	}

	return nil
}
