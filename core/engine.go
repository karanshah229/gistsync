package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
func (e *Engine) SyncFile(localPath string) error {
	absPath, err := filepath.Abs(localPath)
	if err != nil {
		return err
	}

	mapping := e.State.GetMapping(absPath)
	if mapping == nil {
		// New file -> Create gist
		return e.initialSync(absPath, false)
	}

	// Existing mapping -> 3-way sync
	localHash, err := ComputeFileHash(absPath)
	if err != nil {
		return err
	}

	remoteFiles, err := e.Provider.Fetch(mapping.RemoteID)
	if err != nil {
		return err
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
		return fmt.Errorf("remote gist %s is empty", mapping.RemoteID)
	}

	action := DetermineAction(localHash, remoteFile.Hash, mapping.LastSyncedHash)

	switch action {
	case ActionNoop:
		mapping.LastSyncedHash = localHash
		return e.State.Save()
	case ActionPush:
		content, err := os.ReadFile(absPath)
		if err != nil {
			return err
		}
		err = e.Provider.Update(mapping.RemoteID, []File{{Path: filepath.Base(absPath), Content: content}})
		if err != nil {
			return err
		}
		mapping.LastSyncedHash = localHash
		return e.State.Save()
	case ActionPull:
		err = os.WriteFile(absPath, remoteFile.Content, 0644)
		if err != nil {
			return err
		}
		mapping.LastSyncedHash = remoteFile.Hash
		return e.State.Save()
	case ActionConflict:
		return &ConflictError{
			LocalHash:      localHash,
			RemoteHash:     remoteFile.Hash,
			LastSyncedHash: mapping.LastSyncedHash,
		}
	}

	return nil
}

// SyncDir performs a sync for a directory
func (e *Engine) SyncDir(localPath string) error {
	absPath, err := filepath.Abs(localPath)
	if err != nil {
		return err
	}

	mapping := e.State.GetMapping(absPath)
	if mapping == nil {
		return e.initialSync(absPath, false)
	}

	// Fetch remote
	remoteFiles, err := e.Provider.Fetch(mapping.RemoteID)
	if err != nil {
		return err
	}

	// Read local files
	localFiles, err := e.ReadLocalDir(absPath)
	if err != nil {
		return err
	}

	currentLocalHash := e.ComputeDirHash(localFiles)
	remoteHash := e.ComputeDirHash(remoteFiles)
	
	action := DetermineAction(currentLocalHash, remoteHash, mapping.LastSyncedHash) 

	switch action {
	case ActionNoop:
		mapping.LastSyncedHash = currentLocalHash
		return e.State.Save()
	case ActionPush:
		err = e.Provider.Update(mapping.RemoteID, localFiles)
		if err != nil {
			return err
		}
		mapping.LastSyncedHash = currentLocalHash
		return e.State.Save()
	case ActionPull:
		// MVP: Directory pull-all might be complex if we need to delete local files not in gist.
		// For now, write/overwrite what we have.
		for _, rf := range remoteFiles {
			target := filepath.Join(absPath, rf.Path)
			os.MkdirAll(filepath.Dir(target), 0755)
			if err := os.WriteFile(target, rf.Content, 0644); err != nil {
				return err
			}
		}
		mapping.LastSyncedHash = remoteHash
		return e.State.Save()
	case ActionConflict:
		return &ConflictError{
			LocalHash:      currentLocalHash,
			RemoteHash:     remoteHash,
			LastSyncedHash: mapping.LastSyncedHash,
		}
	}

	return nil
}


func (e *Engine) ReadLocalDir(absPath string) ([]File, error) {
	var files []File
	err := filepath.Walk(absPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if path != absPath && strings.HasPrefix(filepath.Base(path), ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasPrefix(info.Name(), ".") {
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

	if isFolder {
		var err error
		files, err = e.ReadLocalDir(absPath)
		if err != nil {
			return err
		}
	} else {
		content, err := os.ReadFile(absPath)
		if err != nil {
			return err
		}
		hash := ComputeHash(content)
		files = append(files, File{Path: filepath.Base(absPath), Content: content, Hash: hash})
	}

	remoteID, err := e.Provider.Create(files, public)
	if err != nil {
		return err
	}

	var hash string
	if isFolder {
		hash = e.ComputeDirHash(files)
	} else {
		hash = files[0].Hash
	}

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

	err = WithLock(func(state *State) error {
		mapping := state.GetMapping(absPath)
		if mapping == nil {
			return fmt.Errorf("path %s is not tracked", path)
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
			fmt.Printf("Warning: failed to delete old gist %s: %v\n", oldRemoteID, err)
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
		remoteHash = e.ComputeDirHash(remoteFiles)
	} else {
		if len(remoteFiles) == 0 {
			return "REMOTE_DELETED", nil
		}
		remoteHash = remoteFiles[0].Hash
	}

	action := DetermineAction(localHash, remoteHash, mapping.LastSyncedHash)
	if action == ActionConflict {
		return SyncAction(fmt.Sprintf("CONFLICT (Local: %s, Remote: %s, LastSynced: %s)", 
			localHash[:8], remoteHash[:8], mapping.LastSyncedHash[:8])), nil
	}
	return action, nil
}
