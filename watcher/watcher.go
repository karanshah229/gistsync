package watcher

import (
	"log/slog"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/karanshah229/gistsync/core"
	"github.com/karanshah229/gistsync/internal"
	"github.com/karanshah229/gistsync/internal/logger"
	"github.com/karanshah229/gistsync/internal/storage"
	"github.com/karanshah229/gistsync/pkg/ui"
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
			logger.Log.Warn("Failed to watch path", slog.String("path", m.LocalPath), slog.String("error", err.Error()))
		}
	}

	ui.Print("WatcherStarted", nil)

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
				name := filepath.Base(event.Name)
				configDir, _ := storage.GetConfigDir()

				// For events inside the config directory: use an allowlist.
				// Only state.json and config.json changes should trigger a config backup.
				// Everything else (tmp, lock, logs) is silently ignored.
				if filepath.Dir(event.Name) == configDir {
					if name == storage.StateFileName || name == storage.ConfigFileName {
						lastPath = configDir
						debounceTimer.Reset(time.Duration(w.Config.WatchDebounce) * time.Millisecond)
					}
					// All other config dir events are ignored
					continue
				}

				lastPath = event.Name
				debounceTimer.Reset(time.Duration(w.Config.WatchDebounce) * time.Millisecond)
			}
		case <-debounceTimer.C:
			// ui.Print(DebounceTriggered", map[string]interface{}{"Path": lastPath}) // Silent for now?
			// Differentiate between real local changes and remote poll updates via hash
			mapping := w.Engine.State.GetMapping(lastPath)
			if mapping == nil {
				// Search for parent mapping if it's a file in a folder
				for _, m := range w.Engine.State.Mappings {
					if m.IsFolder && (lastPath == m.LocalPath || filepath.Dir(lastPath) == m.LocalPath) {
						mapping = &m
						// ui.Print(FoundParentMapping", map[string]interface{}{"Parent": m.LocalPath, "Path": lastPath}) // Silent
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

				configDir, _ := storage.GetConfigDir()
				if err == nil && currentHash == mapping.LastSyncedHash && mapping.LocalPath != configDir {
					// Hash matches LastSyncedHash -> this was likely a remote pull, skip message
					// EXCEPT for the config directory, where we want to sync even if content hasn't changed
					// (because state.json itself might have changed)
					continue
				}
			}

			ui.Print("LocalChangeDetected", map[string]interface{}{"Path": lastPath})
			w.syncPath(lastPath)

		case <-pollTicker.C:
			// Rate limit check
			remaining, _, err := w.Engine.Provider.CheckRateLimit()
			if err != nil {
				logger.Log.Warn("Rate limit check failed", slog.String("error", err.Error()))
			} else if remaining < 10 {
				logger.Log.Warn("Rate limit low, skipping poll cycle", slog.Int("remaining", remaining))
				continue
			}

			ui.Print("CheckingRemoteChanges", nil)
			for _, m := range w.Engine.State.Mappings {
				var err error
				if m.IsFolder {
					_, err = w.Engine.SyncDir(m.LocalPath)
				} else {
					_, err = w.Engine.SyncFile(m.LocalPath)
				}
				if err != nil {
					if _, ok := err.(*core.ConflictError); ok {
						logger.Log.Warn("Remote change detected: CONFLICT", slog.String("path", m.LocalPath))
					} else {
						logger.Log.Error("Remote poll sync error", slog.String("path", m.LocalPath), slog.String("error", err.Error()))
					}
				}
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			logger.Log.Error("Watcher error", slog.String("error", err.Error()))
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
			_, err = w.Engine.SyncDir(mapping.LocalPath)
		} else {
			_, err = w.Engine.SyncFile(mapping.LocalPath)
		}
		if err != nil {
			logger.Log.Error("Auto-sync error", slog.String("path", mapping.LocalPath), slog.String("error", err.Error()))
		} else {
			ui.Success("AutoSyncSuccess", map[string]interface{}{"Path": mapping.LocalPath})
		}
	}
}

