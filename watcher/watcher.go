package watcher

import (
	"fmt"
	"log"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/karanshah229/gistsync/core"
	"github.com/karanshah229/gistsync/internal"
)

type Watcher struct {
	Engine *core.Engine
	Config *internal.Config
}

func NewWatcher(engine *core.Engine, config *internal.Config) *Watcher {
	return &Watcher{Engine: engine, Config: config}
}

func (w *Watcher) Start() error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	// Add all mapped paths to watcher
	for _, m := range w.Engine.State.Mappings {
		err = watcher.Add(m.LocalPath)
		if err != nil {
			log.Printf("Error watching path %s: %v", m.LocalPath, err)
		}
	}

	fmt.Println("GistSync Watcher started. Watching for changes... (Polling remote every 60s)")

	// Polling ticker from config
	pollTicker := time.NewTicker(time.Duration(w.Config.WatchInterval) * time.Second)
	defer pollTicker.Stop()

	// Debounce timer from config
	debounceTimer := time.NewTimer(time.Duration(w.Config.WatchDebounce) * time.Millisecond)
	debounceTimer.Stop()
	var lastPath string

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}
			if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) || event.Has(fsnotify.Rename) {
				lastPath = event.Name
				debounceTimer.Reset(time.Duration(w.Config.WatchDebounce) * time.Millisecond)
			}
		case <-debounceTimer.C:
			fmt.Printf("Debounce triggered for %s\n", lastPath)
			// Differentiate between real local changes and remote poll updates via hash
			mapping := w.Engine.State.GetMapping(lastPath)
			if mapping == nil {
				// Search for parent mapping if it's a file in a folder
				for _, m := range w.Engine.State.Mappings {
					if m.IsFolder && (lastPath == m.LocalPath || filepath.Dir(lastPath) == m.LocalPath) {
						mapping = &m
						fmt.Printf("Found parent mapping %s for %s\n", m.LocalPath, lastPath)
						break
					}
				}
			}
			if mapping != nil {
				var currentHash string
				var err error
				if mapping.IsFolder {
					files, err := w.Engine.ReadLocalDir(mapping.LocalPath)
					if err == nil {
						currentHash = w.Engine.ComputeDirHash(files)
					}
				} else {
					currentHash, err = core.ComputeFileHash(lastPath)
				}

				configDir, _ := internal.GetConfigDir()
				if err == nil && currentHash == mapping.LastSyncedHash && mapping.LocalPath != configDir {
					// Hash matches LastSyncedHash -> this was likely a remote pull, skip message
					// EXCEPT for the config directory, where we want to sync even if content hasn't changed
					// (because state.json itself might have changed)
					continue
				}
			}

			fmt.Printf("Local change detected in %s, syncing...\n", lastPath)
			w.syncPath(lastPath)

		case <-pollTicker.C:
			// Rate limit check
			remaining, _, err := w.Engine.Provider.CheckRateLimit()
			if err != nil {
				log.Printf("Rate limit check failed: %v", err)
			} else if remaining < 10 {
				log.Printf("Rate limit low (%d remaining), skipping poll cycle", remaining)
				continue
			}

			fmt.Println("Checking for remote changes...")
			for _, m := range w.Engine.State.Mappings {
				var err error
				if m.IsFolder {
					err = w.Engine.SyncDir(m.LocalPath)
				} else {
					err = w.Engine.SyncFile(m.LocalPath)
				}
				if err != nil {
					if _, ok := err.(*core.ConflictError); ok {
						log.Printf("Remote change detected for %s: CONFLICT", m.LocalPath)
					} else {
						log.Printf("Remote poll sync error for %s: %v", m.LocalPath, err)
					}
				}
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			log.Println("Watcher error:", err)
		}
	}
}

func (w *Watcher) syncPath(path string) {
	// Determine if it's a file or dir
	mapping := w.Engine.State.GetMapping(path)
	if mapping == nil {
		// Search for parent mapping if it's a file in a folder
		for _, m := range w.Engine.State.Mappings {
			if m.IsFolder && filepath.HasPrefix(path, m.LocalPath) {
				mapping = &m
				break
			}
		}
	}

	if mapping != nil {
		var err error
		if mapping.IsFolder {
			err = w.Engine.SyncDir(mapping.LocalPath)
		} else {
			err = w.Engine.SyncFile(mapping.LocalPath)
		}
		if err != nil {
			log.Printf("Auto-sync error: %v", err)
		} else {
			fmt.Printf("Auto-sync successful for %s\n", mapping.LocalPath)
		}
	}
}

