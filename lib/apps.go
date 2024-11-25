package lib

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
)

type Flatpak struct {
	// a friendly path-safe alias for the full flatpak id
	Alias     string
	FlatpakId string
}

func installFlatpak(app Flatpak) error {
	slog.Info("installing...", "alias", app.Alias, "flatpak", app.FlatpakId)
	err := Cmd("flatpak", "install", app.FlatpakId)
	if err != nil {
		return err
	}
	err = Symln(
		filepath.Join("/var/lib/flatpak/exports/bin", app.FlatpakId),
		filepath.Join("/usr/bin", app.Alias),
	)
	if err != nil {
		return err
	}
	return nil
}

func SyncFlatpak(apps []Flatpak) error {
	for _, a := range apps {
		err := installFlatpak(a)
		if err != nil {
			return fmt.Errorf("install flatpak (%s): %w", a.FlatpakId, err)
		}
	}
	return nil
}

type AppImage struct {
	// a friendly path-safe alias for the app
	Alias string
	Url   string
}

func SyncAppImages(apps []AppImage) error {
	os.Mkdir(filepath.Join(HOME, "AppImages"))

	for _, a := range apps {
		slog.Info("installing...", "alias", a.Alias, "url", a.Url)
		res, err := http.Get(a.Url)
		if err != nil {
			return err
		}
		io.Copy()
	}
}
