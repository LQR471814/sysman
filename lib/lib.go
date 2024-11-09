package lib

import (
	"io/fs"
	"log/slog"
	"os"
	"os/exec"
)

var HOME = os.Getenv("HOME")

func init() {
	if HOME == "" {
		slog.Error("$HOME environment variable is not set.")
		os.Exit(1)
	}
}

func Mkdir(path string) {

}

func WriteFile(path string, data []byte, perm fs.FileMode) error {
	slog.Info(
		"write",
		"path", path,
		"bytes_written", len(data),
		"perm", perm,
	)
	return os.WriteFile(path, data, perm)
}

func RmFile(path string) error {
	slog.Info("rm", "path", path)
	return os.Remove(path)
}

func Cmd(name string, args ...string) error {
	slog.Info("running command", "name", name, "args", args)

	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
