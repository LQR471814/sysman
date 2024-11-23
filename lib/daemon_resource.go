package lib

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

func init() {
	err := os.MkdirAll(servicesDir, 0777)
	if err != nil {
		panic(fmt.Errorf("mkdir -p ~/services: %w", err))
	}
	err = os.MkdirAll(systemdDir, 0777)
	if err != nil {
		panic(fmt.Errorf("mkdir -p ~/.config/systemd/user: %w", err))
	}
}

type Daemon struct {
	// a path-safe name with no spaces
	Id          string
	Description string
	ExecStart   string
}

func (d Daemon) Create(ctx context.Context) error {
	wd := filepath.Join(HOME, "services", d.Id)

	err := os.Mkdir(wd, 0777)
	if err != nil {
		return err
	}

	daemonFile := fmt.Sprintf(
		systemdTemplate,
		d.Description,
		filepath.Join(wd, d.ExecStart),
		wd,
	)

	serviceName := fmt.Sprintf("%s.service", d.Id)

	err = WriteFile(
		ctx,
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

	err = Cmd(ctx, "systemctl", "--user", "enable", d.Id)
	if err != nil {
		return err
	}
	err = Cmd(ctx, "systemctl", "--user", "start", d.Id)
	if err != nil {
		return err
	}

	return nil
}

func (d Daemon) Delete(ctx context.Context) error {
	err := Cmd(ctx, "systemctl", "--user", "stop", d.Id)
	if err != nil {
		return err
	}
	err = Cmd(ctx, "systemctl", "--user", "disable", d.Id)
	if err != nil {
		return err
	}

	err = RmFile(ctx, filepath.Join(systemdDir, d.Id))
	if err != nil {
		return err
	}
	rmdir := filepath.Join(servicesDir, d.Id)
	err = RmDir(ctx, rmdir)
	if err != nil {
		slog.WarnContext(ctx, "remove corresponding service dir", "dir", rmdir, "err", err)
	}

	return nil
}

func (d Daemon) Type() string {
	return "daemon"
}

func (d Daemon) String() string {
	return "daemon:" + d.Id
}

func (d Daemon) Eq(other Resource) bool {
	return d.Id == other.(Daemon).Id
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
