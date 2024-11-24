//go:build js

package storage

import (
	"github.com/VictoriaMetrics/VictoriaMetrics/lib/fs/fsproxy"
	"github.com/VictoriaMetrics/VictoriaMetrics/lib/logger"
	"github.com/VictoriaMetrics/VictoriaMetrics/lib/workingsetcache"
)

func (s *Storage) mustSaveCache(c *workingsetcache.Cache, name string) {
	// this is a nop as we don't care about saving anything
	err := fsproxy.MkdirAll(s.cachePath, 0666)
	if err != nil {
		logger.Panicf("FATAL: cannot save cache to %q: %s", s.cachePath, err)
	}
}
