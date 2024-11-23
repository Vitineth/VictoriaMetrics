//go:build !js

package fsproxy

import (
	"github.com/hack-pad/hackpadfs"
	"github.com/hack-pad/hackpadfs/os"
	"log/slog"
	goOS "os"
)

var fileSystem hackpadfs.FS

func initFs() hackpadfs.FS {
	if fileSystem != nil {
		return fileSystem
	}

	slog.Info("initializing fsproxy using mem backend")
	newFS := os.NewFS()
	workingDirectory, _ := goOS.Getwd()
	workingDirectory, _ = newFS.FromOSPath(workingDirectory)
	workingDirFS, _ := newFS.Sub(workingDirectory)

	fileSystem = workingDirFS
	return workingDirFS
}

func fixPath(name string) string {
	return name
}
