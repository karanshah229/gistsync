package watcher

import (
	"log/slog"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/karanshah229/gistsync/internal/domain"
	"github.com/karanshah229/gistsync/internal/logger"
	"github.com/karanshah229/gistsync/internal/storage"
	"github.com/karanshah229/gistsync/internal/sync"
	"github.com/karanshah229/gistsync/pkg/ui"
)

type Watcher struct {
	Manager *sync.SyncManager
}

func NewWatcher(manager *sync.SyncManager) *Watcher {
	return &Watcher{Manager: manager}
}

func (w *Watcher) Start() error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	// Add all mapped paths to watcher
	state, err := w.Manager.Repo.Load()
	if err != nil {
		return err
	}

	for _, m := range state.Mappings {
		err = watcher.Add(m.LocalPath)
		if err != nil {
			logger.Log.Warn("Failed to watch path", slog.String("path", m.LocalPath), slog.String("error", err.Error()))
		}
	}

	ui.Print("WatcherStarted", nil)

	// Polling ticker from config
	pollTicker := time.NewTicker(time.Duration(w.Manager.Config.WatchInterval) * time.Second)
	defer pollTicker.Stop()

	// Debounce timer from config
	debounceTimer := time.NewTimer(time.Duration(w.Manager.Config.WatchDebounce) * time.Millisecond)
	debounceTimer.Stop()
	var lastPath string

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}
			if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) || event.Has(fsnotify.Rename) {
				name := filepath.Base(event.Name)
				configDir, _ := storage.GetConfigDir()

				if filepath.Dir(event.Name) == configDir {
					if name == storage.StateFileName || name == storage.ConfigFileName {
						lastPath = configDir
						debounceTimer.Reset(time.Duration(w.Manager.Config.WatchDebounce) * time.Millisecond)
					}
					continue
				}

				lastPath = event.Name
				debounceTimer.Reset(time.Duration(w.Manager.Config.WatchDebounce) * time.Millisecond)
			}
		case <-debounceTimer.C:
			state, _ := w.Manager.Repo.Load()
			mapping := state.GetMapping(lastPath)
			if mapping == nil {
				for _, m := range state.Mappings {
					if m.IsFolder && (lastPath == m.LocalPath || filepath.HasPrefix(lastPath, m.LocalPath)) {
						mapping = &m
						break
					}
				}
			}

			if mapping != nil {
				var currentHash string
				var err error
				engine, engineErr := w.Manager.GetEngine(mapping.Provider)
				if engineErr == nil {
					if mapping.IsFolder {
						files, err := engine.ReadLocalDir(mapping.LocalPath)
						if err == nil {
							currentHash = domain.ComputeAggregateHash(files)
						}
					} else {
						currentHash, err = domain.ComputeFileHash(lastPath)
					}
				}

				configDir, _ := storage.GetConfigDir()
				if err == nil && currentHash == mapping.LastSyncedHash && mapping.LocalPath != configDir {
					continue
				}
			}

			ui.Print("LocalChangeDetected", map[string]interface{}{"Path": lastPath})
			w.syncPath(lastPath)

		case <-pollTicker.C:
			ui.Print("CheckingRemoteChanges", nil)
			w.Manager.SyncAll("") // Use default provider

		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			logger.Log.Error("Watcher error", slog.String("error", err.Error()))
		}
	}
}

func (w *Watcher) syncPath(path string) {
	state, _ := w.Manager.Repo.Load()
	mapping := state.GetMapping(path)
	if mapping == nil {
		// Search for parent mapping if it's a file in a folder
		for _, m := range state.Mappings {
			if m.IsFolder && filepath.HasPrefix(path, m.LocalPath) {
				mapping = &m
				break
			}
		}
	}

	if mapping != nil {
		err := w.Manager.SyncPath(mapping.LocalPath, mapping.Provider, "", mapping.Public)
		if err != nil {
			logger.Log.Error("Auto-sync error", slog.String("path", mapping.LocalPath), slog.String("error", err.Error()))
		}
	}
}

