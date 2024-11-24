//go:build !js

package storage

func (s *Storage) mustSaveCache(c *workingsetcache.Cache, name string) {
	saveCacheLock.Lock()
	defer saveCacheLock.Unlock()

	path2 := path.Join(s.cachePath, name)
	if err := c.Save(path2); err != nil {
		logger.Panicf("FATAL: cannot save cache to %q: %s", path2, err)
	}
}
