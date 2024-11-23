//go:build js

package fsproxy

import (
	"github.com/hack-pad/hackpadfs"
	"github.com/hack-pad/hackpadfs/mem"
	"log/slog"
	"path"
	"strings"
)

var virtualWorkingDir = "victoriametrics/"

var fileSystem *mem.FS

func initFs() hackpadfs.FS {
	if fileSystem != nil {
		return fileSystem
	}

	slog.Info("initializing fsproxy using mem backend")
	newFS, err := mem.NewFS()
	if err != nil {
		panic(err)
	}

	fileSystem = newFS
	return fileSystem
}

func fixPath(name string) string {
	if strings.HasPrefix(name, "/") {
		return name[1:]
	}
	return path.Clean(path.Join("/"+virtualWorkingDir, name))[1:]
}
