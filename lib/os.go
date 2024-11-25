package lib

import (
	"context"
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

func WriteFile(ctx context.Context, path string, data []byte, perm fs.FileMode) error {
	slog.InfoContext(
		ctx,
		"write",
		"path", path,
		"bytes_written", len(data),
		"perm", perm,
	)
	return os.WriteFile(path, data, perm)
}

func RmFile(ctx context.Context, path string) error {
	slog.DebugContext(ctx, "rm", "path", path)
	return os.Remove(path)
}

func RmDir(ctx context.Context, path string) error {
	slog.DebugContext(ctx, "rm -rf", "path", path)
	return os.RemoveAll(path)
}

func Symln(ctx context.Context, src, dst string) error {
	slog.DebugContext(ctx, "ln -s", "src", src, "dst", dst)
	return os.Symlink(src, dst)
}

func Cmd(ctx context.Context, name string, args ...string) error {
	slog.DebugContext(ctx, "running command", "name", name, "args", args)

	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
