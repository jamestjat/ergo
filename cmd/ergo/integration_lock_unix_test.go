//go:build !windows

package main

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
)

func TestPrune_LockBusy(t *testing.T) {
	dir := setupErgo(t)

	stdout, _, code := runErgo(t, dir, `{"title":"Task A"}`, "new", "task")
	if code != 0 {
		t.Fatalf("new task failed: exit %d", code)
	}
	taskID := strings.TrimSpace(stdout)
	_, _, code = runErgo(t, dir, `{"state":"done"}`, "set", taskID)
	if code != 0 {
		t.Fatalf("set state=done failed: exit %d", code)
	}

	lockPath := filepath.Join(dir, ".ergo", "lock")
	lockFile, err := os.OpenFile(lockPath, os.O_RDONLY, 0)
	if err != nil {
		t.Fatalf("open lock file: %v", err)
	}
	defer lockFile.Close()
	if err := syscall.Flock(int(lockFile.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		t.Fatalf("failed to acquire lock: %v", err)
	}
	defer func() {
		if err := syscall.Flock(int(lockFile.Fd()), syscall.LOCK_UN); err != nil {
			t.Errorf("failed to release lock: %v", err)
		}
	}()

	before := countEventLines(t, dir)
	_, stderr, code := runErgo(t, dir, "", "prune", "--yes")
	if code == 0 || !strings.Contains(stderr, "lock busy") {
		t.Fatalf("expected lock busy error, got code=%d stderr=%q", code, stderr)
	}
	after := countEventLines(t, dir)
	if before != after {
		t.Fatalf("expected no writes on lock busy (lines %d -> %d)", before, after)
	}
}
