//go:build js

package fs

import (
	"errors"
	"github.com/VictoriaMetrics/VictoriaMetrics/lib/fs/fsproxy"
	"io"
	"sync"
	"sync/atomic"

	"github.com/VictoriaMetrics/VictoriaMetrics/lib/logger"
	"github.com/VictoriaMetrics/metrics"
)

// MustReadAtCloser is rand-access read interface.
type MustReadAtCloser interface {
	// Path must return path for the reader (e.g. file path, url or in-memory reference)
	Path() string

	// MustReadAt must read len(p) bytes from offset off to p.
	MustReadAt(p []byte, off int64)

	// MustClose must close the reader.
	MustClose()
}

// ReaderAt implements rand-access reader.
type ReaderAt struct {
	readCalls atomic.Int64
	readBytes atomic.Int64

	// path contains the path to the file for reading
	path string

	// mr is used for lazy opening of the file at path on the first access.
	mr     atomic.Pointer[mmapReader]
	mrLock sync.Mutex

	useLocalStats bool
}

// Path returns path to r.
func (r *ReaderAt) Path() string {
	return r.path
}

// MustReadAt reads len(p) bytes at off from r.
func (r *ReaderAt) MustReadAt(p []byte, off int64) {
	if len(p) == 0 {
		return
	}
	if off < 0 {
		logger.Panicf("BUG: off=%d cannot be negative", off)
	}

	// Lazily open the file at r.path on the first access
	mr := r.getMmapReader()

	// Read len(p) bytes at offset off to p.
	n, err := mr.f.ReadAt(p, off)
	if err != nil {
		if !errors.Is(err, io.EOF) {
			logger.Panicf("FATAL: cannot read %d bytes at offset %d of file %q: %s", len(p), off, r.path, err)
		}
	}
	if n != len(p) {
		logger.Panicf("FATAL: unexpected number of bytes read from file %q; got %d; want %d", r.path, n, len(p))
	}

	if r.useLocalStats {
		r.readCalls.Add(1)
		r.readBytes.Add(int64(len(p)))
	} else {
		readCalls.Inc()
		readBytes.Add(len(p))
	}
}

func (r *ReaderAt) getMmapReader() *mmapReader {
	mr := r.mr.Load()
	if mr != nil {
		return mr
	}
	r.mrLock.Lock()
	mr = r.mr.Load()
	if mr == nil {
		mr = newMmapReaderFromPath(r.path)
		r.mr.Store(mr)
	}
	r.mrLock.Unlock()
	return mr
}

var (
	readCalls    = metrics.NewCounter(`vm_fs_read_calls_total`)
	readBytes    = metrics.NewCounter(`vm_fs_read_bytes_total`)
	readersCount = metrics.NewCounter(`vm_fs_readers`)
)

// MustClose closes r.
func (r *ReaderAt) MustClose() {
	mr := r.mr.Load()
	if mr != nil {
		mr.mustClose()
		r.mr.Store(nil)
	}

	if r.useLocalStats {
		readCalls.AddInt64(r.readCalls.Load())
		readBytes.AddInt64(r.readBytes.Load())
		r.readCalls.Store(0)
		r.readBytes.Store(0)
		r.useLocalStats = false
	}
}

// SetUseLocalStats switches to local stats collection instead of global stats collection.
//
// This function must be called before the first call to MustReadAt().
//
// Collecting local stats may improve performance on systems with big number of CPU cores,
// since the locally collected stats is pushed to global stats only at MustClose() call
// instead of pushing it at every MustReadAt call.
func (r *ReaderAt) SetUseLocalStats() {
	r.useLocalStats = true
}

// MustFadviseSequentialRead hints the OS that f is read mostly sequentially.
//
// if prefetch is set, then the OS is hinted to prefetch f data.
func (r *ReaderAt) MustFadviseSequentialRead(prefetch bool) {
	mr := r.getMmapReader()
	if err := fadviseSequentialRead(mr.f, prefetch); err != nil {
		logger.Panicf("FATAL: error in fadviseSequentialRead(%q, %v): %s", r.path, prefetch, err)
	}
}

// MustOpenReaderAt opens ReaderAt for reading from the file located at path.
//
// MustClose must be called on the returned ReaderAt when it is no longer needed.
func MustOpenReaderAt(path string) *ReaderAt {
	var r ReaderAt
	r.path = path
	return &r
}

// NewReaderAt returns ReaderAt for reading from f.
//
// NewReaderAt takes ownership for f, so it shouldn't be closed by the caller.
//
// MustClose must be called on the returned ReaderAt when it is no longer needed.
func NewReaderAt(f *fsproxy.ProxyFile) *ReaderAt {
	mr := newMmapReaderFromFile(f)
	var r ReaderAt
	r.path = f.Name()
	r.mr.Store(mr)
	return &r
}

type mmapReader struct {
	f *fsproxy.ProxyFile
}

func newMmapReaderFromPath(path string) *mmapReader {
	f, err := fsproxy.Open(path)
	if err != nil {
		logger.Panicf("FATAL: cannot open file for reading: %s", err)
	}
	return newMmapReaderFromFile(f)
}

func newMmapReaderFromFile(f *fsproxy.ProxyFile) *mmapReader {
	readersCount.Inc()
	return &mmapReader{
		f: f,
	}
}

func (mr *mmapReader) mustClose() {
	MustClose(mr.f)
	mr.f = nil

	readersCount.Dec()
}
