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
		return e.initialSync(absPath)
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
		return e.initialSync(absPath)
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

func (e *Engine) initialSync(absPath string) error {
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

	remoteID, err := e.Provider.Create(files)
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
		LastSyncedHash: hash,
	})
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
