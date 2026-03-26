//go:build !windows

package ergo

import (
	"errors"
	"os"
	"syscall"
)

func syncDir(path string) error {
	dir, err := os.Open(path)
	if err != nil {
		return err
	}
	defer dir.Close()
	return dir.Sync()
}

func withLock(path string, fn func() error) error {
	fd, err := syscall.Open(path, syscall.O_RDONLY, 0)
	if err != nil && errors.Is(err, syscall.ENOENT) {
		if err := ensureFileExists(path, 0644); err != nil {
			return err
		}
		fd, err = syscall.Open(path, syscall.O_RDONLY, 0)
	}
	if err != nil {
		return err
	}
	defer syscall.Close(fd)

	// Fail-fast: non-blocking exclusive lock
	if err := syscall.Flock(fd, syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		if errors.Is(err, syscall.EWOULDBLOCK) || errors.Is(err, syscall.EAGAIN) {
			return ErrLockBusy
		}
		return err
	}
	defer func() {
		_ = syscall.Flock(fd, syscall.LOCK_UN)
	}()
	return fn()
}
