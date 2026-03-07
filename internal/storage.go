package internal

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gofrs/flock"
)

// WriteAtomic writes data to a file atomically by writing to a temporary file,
// syncing it to disk, renaming it, and syncing the parent directory.
func WriteAtomic(path string, data []byte) error {
	tmpPath := path + ".tmp"

	// 1. Write data to temporary file
	f, err := os.OpenFile(tmpPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer func() {
		if f != nil {
			f.Close()
		}
	}()

	if _, err := f.Write(data); err != nil {
		return err
	}

	// 2. Flush data to physical storage
	if err := f.Sync(); err != nil {
		return err
	}

	if err := f.Close(); err != nil {
		return err
	}
	f = nil // prevent double close in defer

	// 3. Atomic rename
	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath)
		return err
	}

	// 4. Sync parent directory to ensure rename is persisted
	dirPath := filepath.Dir(path)
	df, err := os.Open(dirPath)
	if err != nil {
		// If we can't open the dir, we still succeeded in renaming,
		// but crash safety for the rename itself might be compromised.
		return nil
	}
	defer df.Close()

	if err := df.Sync(); err != nil {
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
