//go:build windows

package ergo

import (
	"golang.org/x/sys/windows"
)

// syncDir is a no-op on Windows; directory fsync is a Unix durability pattern.
func syncDir(_ string) error {
	return nil
}

func withLock(path string, fn func() error) error {
	handle, err := windows.CreateFile(
		windows.StringToUTF16Ptr(path),
		windows.GENERIC_READ,
		windows.FILE_SHARE_READ,
		nil,
		windows.OPEN_ALWAYS,
		windows.FILE_ATTRIBUTE_NORMAL,
		0,
	)
	if err != nil {
		return err
	}
	defer windows.CloseHandle(handle)

	// Fail-fast: non-blocking exclusive lock via LockFileEx
	// LOCKFILE_EXCLUSIVE_LOCK | LOCKFILE_FAIL_IMMEDIATELY
	ol := new(windows.Overlapped)
	err = windows.LockFileEx(handle, windows.LOCKFILE_EXCLUSIVE_LOCK|windows.LOCKFILE_FAIL_IMMEDIATELY, 0, 1, 0, ol)
	if err != nil {
		// ERROR_LOCK_VIOLATION (33) means another process holds the lock
		if err == windows.ERROR_LOCK_VIOLATION || err == windows.ERROR_IO_PENDING {
			return ErrLockBusy
		}
		return err
	}
	defer func() {
		_ = windows.UnlockFileEx(handle, 0, 1, 0, ol)
	}()
	return fn()
}
