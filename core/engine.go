package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/karanshah229/gistsync/internal/logger"
	"github.com/karanshah229/gistsync/internal/storage"
	"github.com/karanshah229/gistsync/pkg/ui"
)

type Engine struct {
	State    *State
	Provider Provider
}

func NewEngine(state *State, provider Provider) *Engine {
	return &Engine{
		State:    state,
		Provider: provider,
	}
}

// GetAbsPath returns the absolute path of a local file
func GetAbsPath(path string) (string, error) {
	return filepath.Abs(path)
}


// SyncFile performs a 2-way sync for a single file
func (e *Engine) SyncFile(localPath string) (SyncAction, error) {
	absPath, err := filepath.Abs(localPath)
	if err != nil {
		return "", err
	}

	mapping := e.State.GetMapping(absPath)
	if mapping == nil {
		return ActionPush, e.initialSync(absPath, false)
	}

	// 1. Log Transaction Start
	logger.SyncStart(absPath, mapping.RemoteID, mapping.IsFolder)

	// Existing mapping -> 3-way sync
	localHash, err := ComputeFileHash(absPath)
	if err != nil {
		return "", err
	}

	remoteFiles, err := e.Provider.Fetch(mapping.RemoteID)
	if err != nil {
		return "", err
	}

	// For single file gists, we expect one file. 
	// We'll match by name, but if only one exists we use it.
	var remoteFile *File
	for _, f := range remoteFiles {
		if f.Path == filepath.Base(absPath) {
			remoteFile = &f
			break
		}
	}
	if remoteFile == nil && len(remoteFiles) > 0 {
		remoteFile = &remoteFiles[0]
	}

	if remoteFile == nil {
		return "", fmt.Errorf("remote gist %s is empty — it may have been deleted on the provider", mapping.RemoteID)
	}

	action := DetermineAction(localHash, remoteFile.Hash, mapping.LastSyncedHash)

	switch action {
	case ActionNoop:
		if mapping.LastSyncedHash != localHash {
			return ActionNoop, e.State.WithLock(func(state *State) error {
				if m := state.GetMapping(absPath); m != nil {
					m.LastSyncedHash = localHash
				}
				return nil
			})
		}
		return ActionNoop, nil
	case ActionPush:
		content, err := os.ReadFile(absPath)
		if err != nil {
			return "", err
		}
		err = e.Provider.Update(mapping.RemoteID, []File{{Path: filepath.Base(absPath), Content: content}})
		if err != nil {
			logger.SyncError(absPath, err.Error())
			return "", err
		}
		// 2. Log Transaction Success
		logger.SyncSuccess(absPath, mapping.RemoteID, localHash, mapping.IsFolder, mapping.Provider, mapping.Public)
		// 3. Log Commit (logs CHECKPOINT internally)
		return ActionPush, e.State.WithLock(func(state *State) error {
			if m := state.GetMapping(absPath); m != nil {
				m.LastSyncedHash = localHash
			}
			return nil
		})
	case ActionPull:
		err = os.WriteFile(absPath, remoteFile.Content, 0644)
		if err != nil {
			return "", err
		}
		// 2. Log Transaction Success
		logger.SyncSuccess(absPath, mapping.RemoteID, remoteFile.Hash, mapping.IsFolder, mapping.Provider, mapping.Public)
		// 3. Log Commit (logs CHECKPOINT internally)
		return ActionPull, e.State.WithLock(func(state *State) error {
			if m := state.GetMapping(absPath); m != nil {
				m.LastSyncedHash = remoteFile.Hash
			}
			return nil
		})
	case ActionConflict:
		return "", &ConflictError{
			LocalHash:      localHash,
			RemoteHash:     remoteFile.Hash,
			LastSyncedHash: mapping.LastSyncedHash,
		}
	}

	return "", nil
}

// SyncDir performs a sync for a directory
func (e *Engine) SyncDir(localPath string) (SyncAction, error) {
	absPath, err := filepath.Abs(localPath)
	if err != nil {
		return "", err
	}

	mapping := e.State.GetMapping(absPath)
	if mapping == nil {
		return ActionPush, e.initialSync(absPath, false)
	}

	// Fetch remote
	remoteFiles, err := e.Provider.Fetch(mapping.RemoteID)
	if err != nil {
		return "", err
	}

	// Read local files
	localFiles, err := e.ReadLocalDir(absPath)
	if err != nil {
		return "", err
	}

	currentLocalHash := e.ComputeDirHash(localFiles)
	
	// Filter remote files to exclude state.json for consistent hash comparison
	var filteredRemote []File
	configDir, _ := storage.GetConfigDir()
	for _, f := range remoteFiles {
		if absPath == configDir && storage.IsIgnoredConfigFile(f.Path) {
			continue
		}
		filteredRemote = append(filteredRemote, f)
	}
	remoteHash := e.ComputeDirHash(filteredRemote)
	
	// 1. Log Transaction Start
	logger.SyncStart(absPath, mapping.RemoteID, mapping.IsFolder)

	action := DetermineAction(currentLocalHash, remoteHash, mapping.LastSyncedHash) 

	// Special case for config directory: if content is same, but state.json is missing or different in gist,
	// we still want to push (ActionPush) to ensure our backup is current.
	if action == ActionNoop && absPath == configDir {
		stateInGist := false
		var remoteStateContent []byte
		for _, f := range remoteFiles {
			if f.Path == "state.json" {
				stateInGist = true
				remoteStateContent = f.Content
				break
			}
		}

		// Load freshest state for comparison
		freshState, err := LoadState()
		if err == nil {
			localStateJSON, _ := json.MarshalIndent(freshState, "", "  ")
			if !stateInGist || string(remoteStateContent) != string(localStateJSON) {
				action = ActionPush
			}
		} else {
			// If we can't load state, fallback to push to be safe
			action = ActionPush
		}
	}

	switch action {
	case ActionNoop:
		if mapping.LastSyncedHash != currentLocalHash {
			return ActionNoop, e.State.WithLock(func(state *State) error {
				if m := state.GetMapping(absPath); m != nil {
					m.LastSyncedHash = currentLocalHash
				}
				return nil
			})
		}
		return ActionNoop, nil
	case ActionPush:
		uploadFiles := localFiles
		if absPath == configDir {
			// Virtual State Projection: Inject CURRENT state into backup
			// We load the freshest state from disk to ensure we include all recent changes.
			freshState, err := LoadState()
			if err == nil {
				// Update the config mapping's hash in the fresh state BEFORE projection.
				// This ensures that the projected state matches what will be saved locally.
				for i, m := range freshState.Mappings {
					if m.LocalPath == absPath {
						freshState.Mappings[i].LastSyncedHash = currentLocalHash
						break
					}
				}

				stateJSON, err := json.MarshalIndent(freshState, "", "  ")
				if err == nil {
					// Avoid duplicates if it was already in localFiles (though ReadLocalDir should have excluded it)
					found := false
					for i, f := range uploadFiles {
						if f.Path == "state.json" {
							uploadFiles[i].Content = stateJSON
							uploadFiles[i].Hash = ComputeHash(stateJSON)
							found = true
							break
						}
					}
					if !found {
						uploadFiles = append(uploadFiles, File{
							Path:    "state.json",
							Content: stateJSON,
							Hash:    ComputeHash(stateJSON),
						})
					}
				}
			}
		}

		err = e.Provider.Update(mapping.RemoteID, uploadFiles)
		if err != nil {
			logger.SyncError(absPath, err.Error())
			return "", err
		}
		// 2. Log Transaction Success
		logger.SyncSuccess(absPath, mapping.RemoteID, currentLocalHash, mapping.IsFolder, mapping.Provider, mapping.Public)
		// 3. Log Commit (logs CHECKPOINT internally)
		return ActionPush, e.State.WithLock(func(state *State) error {
			if m := state.GetMapping(absPath); m != nil {
				m.LastSyncedHash = currentLocalHash
			}
			return nil
		})
	case ActionPull:
		// MVP: Directory pull-all might be complex if we need to delete local files not in gist.
		// For now, write/overwrite what we have.
		for _, rf := range remoteFiles {
			target := filepath.Join(absPath, rf.Path)
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return "", fmt.Errorf("failed to create directory for %s: %w", target, err)
			}
			if err := os.WriteFile(target, rf.Content, 0644); err != nil {
				return "", err
			}
		}
		// 2. Log Transaction Success
		logger.SyncSuccess(absPath, mapping.RemoteID, remoteHash, mapping.IsFolder, mapping.Provider, mapping.Public)
		// 3. Log Commit (logs CHECKPOINT internally)
		return ActionPull, e.State.WithLock(func(state *State) error {
			if m := state.GetMapping(absPath); m != nil {
				m.LastSyncedHash = remoteHash
			}
			return nil
		})
	case ActionConflict:
		return "", &ConflictError{
			LocalHash:      currentLocalHash,
			RemoteHash:     remoteHash,
			LastSyncedHash: mapping.LastSyncedHash,
		}
	}

	return "", nil
}


func (e *Engine) ReadLocalDir(absPath string) ([]File, error) {
	configDir, _ := storage.GetConfigDir()
	isConfigDir := absPath == configDir

	var files []File
	err := filepath.Walk(absPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if path != absPath && strings.HasPrefix(filepath.Base(path), ".") {
				return filepath.SkipDir
			}

			if isConfigDir && storage.IsIgnoredConfigFile(info.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasPrefix(info.Name(), ".") {
			return nil
		}

		// Exclude state files only if in config directory
		if isConfigDir && storage.IsIgnoredConfigFile(info.Name()) {
			return nil
		}

		rel, _ := filepath.Rel(absPath, path)
		// Ensure gists always use forward slashes for cross-platform compatibility
		rel = filepath.ToSlash(rel)
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		files = append(files, File{
			Path:    rel,
			Content: content,
			Hash:    ComputeHash(content),
		})
		return nil
	})
	return files, err
}

func (e *Engine) ComputeDirHash(files []File) string {
	// Combine hashes of all files
	var combined string
	for _, f := range files {
		combined += f.Path + ":" + f.Hash + ";"
	}
	return ComputeHash([]byte(combined))
}

func (e *Engine) initialSync(absPath string, public bool) error {
	info, err := os.Stat(absPath)
	if err != nil {
		return err
	}

	var files []File
	isFolder := info.IsDir()

	var hash string
	if isFolder {
		var err error
		files, err = e.ReadLocalDir(absPath)
		if err != nil {
			return err
		}
		hash = e.ComputeDirHash(files)
	} else {
		content, err := os.ReadFile(absPath)
		if err != nil {
			return err
		}
		hash = ComputeHash(content)
		files = append(files, File{Path: filepath.Base(absPath), Content: content, Hash: hash})
	}

	uploadFiles := files
	configDir, _ := storage.GetConfigDir()
	if absPath == configDir {
		// Virtual State Projection for initial sync:
		// We need to include the mapping we are ABOUT to create in the freshest state.
		freshState, err := LoadState()
		if err != nil {
			// If state.json doesn't exist yet (e.g. first sync), use empty state
			freshState = &State{Version: e.State.Version}
		}

		freshState.Mappings = append(freshState.Mappings, Mapping{
			LocalPath:      absPath,
			RemoteID:       "PENDING", 
			Provider:       "github",
			IsFolder:       isFolder,
			Public:         public,
			LastSyncedHash: hash,
		})

		stateJSON, err := json.MarshalIndent(freshState, "", "  ")
		if err == nil {
			uploadFiles = append(uploadFiles, File{
				Path:    "state.json",
				Content: stateJSON,
				Hash:    ComputeHash(stateJSON),
			})
		}
	}

	// 1. Log Transaction Start
	logger.SyncStart(absPath, "", isFolder)

	remoteID, err := e.Provider.Create(uploadFiles, public)
	if err != nil {
		logger.SyncError(absPath, err.Error())
		return err
	}

	// 2. Log Transaction Success
	logger.SyncSuccess(absPath, remoteID, hash, isFolder, "github", public)
	// AddMapping handles persistence and updating e.State internally using WithLock
	return e.State.AddMapping(Mapping{
		LocalPath:      absPath,
		RemoteID:       remoteID,
		Provider:       "github",
		IsFolder:       isFolder,
		Public:         public,
		LastSyncedHash: hash,
	})
}

// InitialSyncWithVisibility provides an exported way to set visibility initially
func (e *Engine) InitialSyncWithVisibility(absPath string, public bool) error {
	return e.initialSync(absPath, public)
}

// SetVisibility changes the visibility of an existing mapping using a transactional flow
// It performs a conflict check and ensure the "latest" content is used for the new gist.
func (e *Engine) SetVisibility(path string, public bool) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	var oldRemoteID string
	var newRemoteID string

	err = e.State.WithLock(func(state *State) error {
		mapping := state.GetMapping(absPath)
		if mapping == nil {
			return fmt.Errorf("path %s is not tracked — use 'gistsync sync %s' to start tracking it", path, path)
		}

		if mapping.Public == public {
			return nil // Already at target visibility
		}

		oldRemoteID = mapping.RemoteID

		// 1. Fetch remote content
		remoteFiles, fetchErr := e.Provider.Fetch(oldRemoteID)
		if fetchErr != nil {
			return fmt.Errorf("failed to fetch remote content: %w", fetchErr)
		}

		// 2. Read local content
		var localFiles []File
		var readErr error
		if mapping.IsFolder {
			localFiles, readErr = e.ReadLocalDir(absPath)
		} else {
			content, readErr2 := os.ReadFile(absPath)
			if readErr2 == nil {
				localFiles = []File{{
					Path:    filepath.Base(absPath),
					Content: content,
					Hash:    ComputeHash(content),
				}}
			}
			readErr = readErr2
		}
		if readErr != nil {
			return fmt.Errorf("failed to read local content: %w", readErr)
		}

		// 3. Determine action and handle conflicts
		localHash := e.ComputeDirHash(localFiles)
		if !mapping.IsFolder {
			localHash = localFiles[0].Hash
		}
		remoteHash := e.ComputeDirHash(remoteFiles)
		if !mapping.IsFolder {
			if len(remoteFiles) == 0 {
				return fmt.Errorf("remote gist is empty")
			}
			remoteHash = remoteFiles[0].Hash
		}

		action := DetermineAction(localHash, remoteHash, mapping.LastSyncedHash)
		if action == ActionConflict {
			return &ConflictError{
				LocalHash:      localHash,
				RemoteHash:     remoteHash,
				LastSyncedHash: mapping.LastSyncedHash,
			}
		}

		// 4. Content Selection: Choose latest
		targetFiles := remoteFiles
		newSyncedHash := remoteHash
		if action == ActionPush {
			targetFiles = localFiles
			newSyncedHash = localHash
		}

		// 5. Create new gist
		var createErr error
		newRemoteID, createErr = e.Provider.Create(targetFiles, public)
		if createErr != nil {
			return fmt.Errorf("failed to create new gist: %w", createErr)
		}

		// 6. Update mapping and save
		mapping.RemoteID = newRemoteID
		mapping.Public = public
		mapping.LastSyncedHash = newSyncedHash
		return nil // WithLock calls state.Save() here
	})

	if err != nil {
		// Rollback logic: if we created a new gist (or partially created it) 
		// but the transaction failed to commit to our local state, delete it.
		if newRemoteID != "" {
			_ = e.Provider.Delete(newRemoteID)
		}
		return err
	}

	// 7. Clean up
	if oldRemoteID != "" {
		if err := e.Provider.Delete(oldRemoteID); err != nil {
			ui.Warning("GistDeleteWarning", map[string]interface{}{"ID": oldRemoteID, "Err": err})
		}
	}

	return nil
}

// Status returns the sync status of a path
func (e *Engine) Status(localPath string) (SyncAction, error) {
	absPath, err := filepath.Abs(localPath)
	if err != nil {
		return "", err
	}

	mapping := e.State.GetMapping(absPath)
	if mapping == nil {
		return "UNTRACKED", nil
	}

	configDir, _ := storage.GetConfigDir()

	var localHash string
	if mapping.IsFolder {
		files, err := e.ReadLocalDir(absPath)
		if err != nil {
			return "", err
		}
		localHash = e.ComputeDirHash(files)
	} else {
		localHash, err = ComputeFileHash(absPath)
		if err != nil {
			return "", err
		}
	}

	remoteFiles, err := e.Provider.Fetch(mapping.RemoteID)
	if err != nil {
		return "", err
	}

	var remoteHash string
	if mapping.IsFolder {
		// Filter remote files to match ReadLocalDir logic (exclude state.json etc.)
		var filteredRemote []File
		for _, f := range remoteFiles {
			if absPath == configDir && storage.IsIgnoredConfigFile(f.Path) {
				continue
			}
			filteredRemote = append(filteredRemote, f)
		}
		remoteHash = e.ComputeDirHash(filteredRemote)
	} else {
		if len(remoteFiles) == 0 {
			return "REMOTE_DELETED", nil
		}
		remoteHash = remoteFiles[0].Hash
	}

	action := DetermineAction(localHash, remoteHash, mapping.LastSyncedHash)

	// Special case for config directory: if content is same, but state.json is missing or different in gist,
	// we still want to report ActionPush to ensure our backup is current.
	if action == ActionNoop && absPath == configDir {
		stateInGist := false
		var remoteStateContent []byte
		for _, f := range remoteFiles {
			if f.Path == "state.json" {
				stateInGist = true
				remoteStateContent = f.Content
				break
			}
		}

		// Load freshest state for comparison
		freshState, err := LoadState()
		if err == nil {
			localStateJSON, _ := json.MarshalIndent(freshState, "", "  ")
			if !stateInGist || string(remoteStateContent) != string(localStateJSON) {
				action = ActionPush
			}
		} else {
			action = ActionPush
		}
	}

	if action == ActionConflict {
		return SyncAction(fmt.Sprintf("CONFLICT (Local: %s, Remote: %s, LastSynced: %s)", 
			truncHash(localHash), truncHash(remoteHash), truncHash(mapping.LastSyncedHash))), nil
	}
	return action, nil
}
