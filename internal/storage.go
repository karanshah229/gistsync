package internal

import (
	"fmt"
	"os"

	"github.com/gofrs/flock"
)

// WriteAtomic writes data to a file atomically by writing to a temporary file and renaming it.
func WriteAtomic(path string, data []byte) error {
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return err
	}

	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath)
		return err
	}

	return nil
}

// WithFileLock executes the given function while holding a file lock on the specified path.
func WithFileLock(path string, fn func() error) error {
	lockPath := path + ".lock"
	fileLock := flock.New(lockPath)

	if err := fileLock.Lock(); err != nil {
		return fmt.Errorf("failed to acquire lock on %s: %w", path, err)
	}
	defer fileLock.Unlock()

	return fn()
}
