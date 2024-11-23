package fsproxy

import (
	"errors"
	"github.com/google/uuid"
	"github.com/hack-pad/hackpadfs"
	"io/fs"
	"log/slog"
	"os"
	"path"
	"path/filepath"
)

func Glob(pattern string) ([]string, error) {
	initFs()
	pattern = filepath.Clean(pattern)
	pattern = filepath.ToSlash(pattern)

	return fs.Glob(fileSystem, pattern)
}

func ReadFile(name string) ([]byte, error) {
	initFs()
	slog.Info("reading file", "name", fixPath(name))
	return fs.ReadFile(fileSystem, fixPath(name))
}

func Abs(name string) (string, error) {
	initFs()
	slog.Info("getting abs for", "name", name, "fixed", fixPath(name))
	return fixPath(name), nil
}

func Stat(name string) (hackpadfs.FileInfo, error) {
	initFs()
	slog.Info("stat", "name", fixPath(name))
	return hackpadfs.Stat(fileSystem, fixPath(name))
}

func MkdirAll(path string, perm hackpadfs.FileMode) error {
	initFs()
	slog.Info("mkdir all", "path", fixPath(path))
	return hackpadfs.MkdirAll(fileSystem, fixPath(path), perm)
}

func WriteFile(name string, data []byte, perm os.FileMode) error {
	initFs()
	slog.Info("write file", "name", fixPath(name))
	file, err := hackpadfs.OpenFile(fileSystem, fixPath(name), os.O_CREATE|os.O_WRONLY, perm)
	if err != nil {
		return err
	}
	_, err = hackpadfs.WriteFile(file, data)
	if err != nil {
		return err
	}
	err = file.Close()
	if err != nil {
		return err
	}

	return nil
}

type ProxyFile struct {
	descriptor *fs.File
	path       string
}

func Create(name string) (*ProxyFile, error) {
	initFs()
	fixedPath := fixPath(name)
	slog.Info("create file", "name", fixedPath)
	create, err := hackpadfs.Create(fileSystem, fixedPath)
	if err != nil {
		return nil, err
	}

	return &ProxyFile{descriptor: &create, path: fixedPath}, nil
}

func Rename(from string, to string) error {
	initFs()
	slog.Info("name file", "from", fixPath(from), "to", fixPath(to))
	return hackpadfs.Rename(fileSystem, fixPath(from), fixPath(to))
}

func Open(name string) (*ProxyFile, error) {
	initFs()
	fixedPath := fixPath(name)
	slog.Info("open file", "name", fixedPath)
	open, err := fileSystem.Open(fixedPath)
	if err != nil {
		return nil, err
	}

	return &ProxyFile{descriptor: &open, path: fixedPath}, nil
}

func OpenFile(name string, flag int, perm os.FileMode) (*ProxyFile, error) {
	file, err := hackpadfs.OpenFile(fileSystem, fixPath(name), flag, perm)
	if err != nil {
		return nil, err
	}

	return &ProxyFile{descriptor: &file, path: fixPath(name)}, nil
}

func ReadDir(name string) ([]os.DirEntry, error) {
	initFs()
	slog.Info("reading dir", "dir", fixPath(name))
	return hackpadfs.ReadDir(fileSystem, fixPath(name))
}

func Link(from string, to string) error {
	initFs()
	slog.Info("link file", "from", fixPath(from), "to", fixPath(to))
	return hackpadfs.Symlink(fileSystem, fixPath(from), fixPath(to))
}

func CreateTemp(dir string, pattern string) (*ProxyFile, error) {
	initFs()
	slog.Info("create temp in dir", "dir", fixPath(dir))
	createTemp, err := Create(fixPath(path.Join(dir, uuid.NewString())))
	if err != nil {
		return nil, err
	}

	return createTemp, nil
}

func RemoveAll(path string) error {
	return hackpadfs.RemoveAll(fileSystem, fixPath(path))
}

func (p *ProxyFile) Close() error {
	return (*p.descriptor).Close()
}

func (p *ProxyFile) Stat() (os.FileInfo, error) {
	return (*p.descriptor).Stat()
}

func (p *ProxyFile) Readdirnames(n int) ([]string, error) {
	file, err := hackpadfs.ReadDirFile(*p.descriptor, n)
	if err != nil {
		return nil, err
	}

	names := make([]string, len(file))
	for i, entry := range file {
		names[i] = entry.Name()
	}

	return names, nil
}

func (p *ProxyFile) ReadAt(b []byte, off int64) (int, error) {
	return hackpadfs.ReadAtFile(*p.descriptor, b, off)
}

func (p *ProxyFile) Read(b []byte) (int, error) {
	return (*p.descriptor).Read(b)
}

func (p *ProxyFile) Write(b []byte) (int, error) {
	slog.Info("writing", "file", p.path, "bytes", len(b))
	return hackpadfs.WriteFile(*p.descriptor, b)
}

func (p *ProxyFile) Name() string {
	stat, err := (*p.descriptor).Stat()
	if err != nil {
		panic(err)
	}
	return stat.Name()
}

func (p *ProxyFile) Sync() error {
	err := hackpadfs.SyncFile(*p.descriptor)
	if errors.Is(err, hackpadfs.ErrNotImplemented) {
		return nil
	}
	return err
}

func (p *ProxyFile) Fd() uintptr {
	return 0
}

func (p *ProxyFile) Seek(offset int64, whence int) (int64, error) {
	return hackpadfs.SeekFile(*p.descriptor, offset, whence)
}
