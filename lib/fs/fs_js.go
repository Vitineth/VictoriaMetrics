//go:build js

package fs

import (
	"fmt"
	"github.com/VictoriaMetrics/VictoriaMetrics/lib/fs/fsproxy"
	"github.com/VictoriaMetrics/VictoriaMetrics/lib/logger"
	"os"
)

func mustSyncPath(path string) {}

func createFlockFile(flockFile string) (*fsproxy.ProxyFile, error) {
	flockF, err := fsproxy.Create(flockFile)
	if err != nil {
		return nil, fmt.Errorf("cannot create lock file %q: %w", flockFile, err)
	}

	return flockF, nil
}

func mustGetFreeSpace(path string) uint64 {
	return 1000000000
}

func mustRemoveDirAtomic(dir string) {
	n := atomicDirRemoveCounter.Add(1)
	tmpDir := fmt.Sprintf("%s.must-remove.%d", dir, n)
	if err := os.Rename(dir, tmpDir); err != nil {
		logger.Panicf("FATAL: cannot move %s to %s: %s", dir, tmpDir, err)
	}
	MustRemoveAll(tmpDir)
}
