package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/karanshah229/gistsync/internal/domain"
	"github.com/karanshah229/gistsync/internal/logger"
	"github.com/karanshah229/gistsync/internal/state"
	"github.com/karanshah229/gistsync/internal/storage"
	"github.com/karanshah229/gistsync/pkg/ui"
)

type Engine struct {
	Repo     state.Repository
	Provider domain.Provider
}

func NewEngine(repo state.Repository, provider domain.Provider) *Engine {
	return &Engine{
		Repo:     repo,
		Provider: provider,
	}
}

// GetAbsPath returns the absolute path of a local file
func GetAbsPath(path string) (string, error) {
	return filepath.Abs(path)
}

// SyncFile performs a 2-way sync for a single file
func (e *Engine) SyncFile(localPath string) (domain.SyncAction, error) {
	absPath, err := filepath.Abs(localPath)
	if err != nil {
		return "", err
	}

	mapping, err := e.Repo.GetMapping(absPath)
	if err != nil {
		return "", err
	}
	if mapping == nil {
		return domain.ActionPush, e.initialSync(absPath, false)
	}

	logger.SyncStart(absPath, mapping.RemoteID, mapping.IsFolder)

	localHash, err := ComputeFileHash(absPath)
	if err != nil {
		return "", err
	}

	remoteFiles, err := e.Provider.Fetch(mapping.RemoteID)
	if err != nil {
		return "", err
	}

	remoteFile := e.matchRemoteFile(absPath, remoteFiles)
	if remoteFile == nil {
		return "", fmt.Errorf("remote gist %s is empty — it may have been deleted on the provider", mapping.RemoteID)
	}

	action := DetermineAction(localHash, remoteFile.Hash, mapping.LastSyncedHash)

	switch action {
	case domain.ActionNoop:
		return domain.ActionNoop, e.updateLastSyncedHashIfNeeded(absPath, mapping, localHash)
	case domain.ActionPush:
		return e.pushFile(absPath, mapping, localHash)
	case domain.ActionPull:
		return e.pullFile(absPath, mapping, remoteFile)
	case domain.ActionConflict:
		return "", &ConflictError{
			LocalHash:      localHash,
			RemoteHash:     remoteFile.Hash,
			LastSyncedHash: mapping.LastSyncedHash,
		}
	}

	return "", nil
}

// SyncDir performs a sync for a directory
func (e *Engine) SyncDir(localPath string) (domain.SyncAction, error) {
	absPath, err := filepath.Abs(localPath)
	if err != nil {
		return "", err
	}

	mapping, err := e.Repo.GetMapping(absPath)
	if err != nil {
		return "", err
	}
	if mapping == nil {
		return domain.ActionPush, e.initialSync(absPath, false)
	}

	remoteFiles, err := e.Provider.Fetch(mapping.RemoteID)
	if err != nil {
		return "", err
	}

	localFiles, err := e.ReadLocalDir(absPath)
	if err != nil {
		return "", err
	}

	currentLocalHash := e.ComputeDirHash(localFiles)
	remoteHash := e.computeRemoteDirHash(absPath, remoteFiles)

	logger.SyncStart(absPath, mapping.RemoteID, mapping.IsFolder)

	action := DetermineAction(currentLocalHash, remoteHash, mapping.LastSyncedHash)
	action = e.handleConfigDirSpecialCase(action, absPath, remoteFiles, currentLocalHash)

	switch action {
	case domain.ActionNoop:
		return domain.ActionNoop, e.updateLastSyncedHashIfNeeded(absPath, mapping, currentLocalHash)
	case domain.ActionPush:
		return e.pushDir(absPath, mapping, localFiles, currentLocalHash)
	case domain.ActionPull:
		return e.pullDir(absPath, mapping, remoteFiles, remoteHash)
	case domain.ActionConflict:
		return "", &ConflictError{
			LocalHash:      currentLocalHash,
			RemoteHash:     remoteHash,
			LastSyncedHash: mapping.LastSyncedHash,
		}
	}

	return "", nil
}

func (e *Engine) matchRemoteFile(absPath string, remoteFiles []domain.File) *domain.File {
	for _, f := range remoteFiles {
		if f.Path == filepath.Base(absPath) {
			return &f
		}
	}
	if len(remoteFiles) > 0 {
		return &remoteFiles[0]
	}
	return nil
}

func (e *Engine) updateLastSyncedHashIfNeeded(absPath string, mapping *domain.Mapping, currentHash string) error {
	if mapping.LastSyncedHash != currentHash {
		return e.Repo.WithLock(func(state *domain.State) error {
			if m := e.findMappingInState(state, absPath); m != nil {
				m.LastSyncedHash = currentHash
			}
			return nil
		})
	}
	return nil
}

func (e *Engine) findMappingInState(state *domain.State, absPath string) *domain.Mapping {
	for i := range state.Mappings {
		if state.Mappings[i].LocalPath == absPath {
			return &state.Mappings[i]
		}
	}
	return nil
}

func (e *Engine) pushFile(absPath string, mapping *domain.Mapping, localHash string) (domain.SyncAction, error) {
	content, err := os.ReadFile(absPath)
	if err != nil {
		return "", err
	}
	err = e.Provider.Update(mapping.RemoteID, []domain.File{{Path: filepath.Base(absPath), Content: content}})
	if err != nil {
		logger.SyncError(absPath, err.Error())
		return "", err
	}
	logger.SyncSuccess(absPath, mapping.RemoteID, localHash, mapping.IsFolder, mapping.Provider, mapping.Public)
	return domain.ActionPush, e.Repo.WithLock(func(state *domain.State) error {
		if m := e.findMappingInState(state, absPath); m != nil {
			m.LastSyncedHash = localHash
		}
		return nil
	})
}

func (e *Engine) pullFile(absPath string, mapping *domain.Mapping, remoteFile *domain.File) (domain.SyncAction, error) {
	err := os.WriteFile(absPath, remoteFile.Content, 0644)
	if err != nil {
		return "", err
	}
	logger.SyncSuccess(absPath, mapping.RemoteID, remoteFile.Hash, mapping.IsFolder, mapping.Provider, mapping.Public)
	return domain.ActionPull, e.Repo.WithLock(func(state *domain.State) error {
		if m := e.findMappingInState(state, absPath); m != nil {
			m.LastSyncedHash = remoteFile.Hash
		}
		return nil
	})
}

func (e *Engine) computeRemoteDirHash(absPath string, remoteFiles []domain.File) string {
	var filteredRemote []domain.File
	configDir, _ := storage.GetConfigDir()
	for _, f := range remoteFiles {
		if absPath == configDir && storage.IsIgnoredConfigFile(f.Path) {
			continue
		}
		filteredRemote = append(filteredRemote, f)
	}
	return e.ComputeDirHash(filteredRemote)
}

func (e *Engine) handleConfigDirSpecialCase(action domain.SyncAction, absPath string, remoteFiles []domain.File, currentLocalHash string) domain.SyncAction {
	configDir, _ := storage.GetConfigDir()
	if action != domain.ActionNoop || absPath != configDir {
		return action
	}

	stateInGist := false
	var remoteStateContent []byte
	for _, f := range remoteFiles {
		if f.Path == "state.json" {
			stateInGist = true
			remoteStateContent = f.Content
			break
		}
	}

	freshState, err := e.Repo.Load()
	if err == nil {
		localStateJSON, _ := json.MarshalIndent(freshState, "", "  ")
		if !stateInGist || string(remoteStateContent) != string(localStateJSON) {
			return domain.ActionPush
		}
	} else {
		return domain.ActionPush
	}
	return action
}

func (e *Engine) pushDir(absPath string, mapping *domain.Mapping, localFiles []domain.File, currentLocalHash string) (domain.SyncAction, error) {
	uploadFiles := localFiles
	configDir, _ := storage.GetConfigDir()
	if absPath == configDir {
		uploadFiles = e.injectStateIntoBackup(absPath, uploadFiles, currentLocalHash)
	}

	err := e.Provider.Update(mapping.RemoteID, uploadFiles)
	if err != nil {
		logger.SyncError(absPath, err.Error())
		return "", err
	}
	logger.SyncSuccess(absPath, mapping.RemoteID, currentLocalHash, mapping.IsFolder, mapping.Provider, mapping.Public)
	return domain.ActionPush, e.Repo.WithLock(func(state *domain.State) error {
		if m := e.findMappingInState(state, absPath); m != nil {
			m.LastSyncedHash = currentLocalHash
		}
		return nil
	})
}

func (e *Engine) injectStateIntoBackup(absPath string, uploadFiles []domain.File, currentLocalHash string) []domain.File {
	freshState, err := e.Repo.Load()
	if err != nil {
		return uploadFiles
	}

	for i, m := range freshState.Mappings {
		if m.LocalPath == absPath {
			freshState.Mappings[i].LastSyncedHash = currentLocalHash
			break
		}
	}

	stateJSON, err := json.MarshalIndent(freshState, "", "  ")
	if err != nil {
		return uploadFiles
	}

	for i, f := range uploadFiles {
		if f.Path == "state.json" {
			uploadFiles[i].Content = stateJSON
			uploadFiles[i].Hash = ComputeHash(stateJSON)
			return uploadFiles
		}
	}

	return append(uploadFiles, domain.File{
		Path:    "state.json",
		Content: stateJSON,
		Hash:    ComputeHash(stateJSON),
	})
}

func (e *Engine) pullDir(absPath string, mapping *domain.Mapping, remoteFiles []domain.File, remoteHash string) (domain.SyncAction, error) {
	for _, rf := range remoteFiles {
		target := filepath.Join(absPath, rf.Path)
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return "", fmt.Errorf("failed to create directory for %s: %w", target, err)
		}
		if err := os.WriteFile(target, rf.Content, 0644); err != nil {
			return "", err
		}
	}
	logger.SyncSuccess(absPath, mapping.RemoteID, remoteHash, mapping.IsFolder, mapping.Provider, mapping.Public)
	return domain.ActionPull, e.Repo.WithLock(func(state *domain.State) error {
		if m := e.findMappingInState(state, absPath); m != nil {
			m.LastSyncedHash = remoteHash
		}
		return nil
	})
}

func (e *Engine) ReadLocalDir(absPath string) ([]domain.File, error) {
	configDir, _ := storage.GetConfigDir()
	isConfigDir := absPath == configDir

	var files []domain.File
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
		if isConfigDir && storage.IsIgnoredConfigFile(info.Name()) {
			return nil
		}

		rel, _ := filepath.Rel(absPath, path)
		rel = filepath.ToSlash(rel)
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		files = append(files, domain.File{
			Path:    rel,
			Content: content,
			Hash:    ComputeHash(content),
		})
		return nil
	})
	return files, err
}

func (e *Engine) ComputeDirHash(files []domain.File) string {
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

	var files []domain.File
	isFolder := info.IsDir()
	var hash string

	if isFolder {
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
		files = []domain.File{{Path: filepath.Base(absPath), Content: content, Hash: hash}}
	}

	uploadFiles := files
	configDir, _ := storage.GetConfigDir()
	if absPath == configDir {
		freshState, err := e.Repo.Load()
		if err != nil {
			freshState = &domain.State{}
		}

		freshState.Mappings = append(freshState.Mappings, domain.Mapping{
			LocalPath:      absPath,
			RemoteID:       "PENDING",
			Provider:       "github",
			IsFolder:       isFolder,
			Public:         public,
			LastSyncedHash: hash,
		})

		stateJSON, err := json.MarshalIndent(freshState, "", "  ")
		if err == nil {
			uploadFiles = append(uploadFiles, domain.File{
				Path:    "state.json",
				Content: stateJSON,
				Hash:    ComputeHash(stateJSON),
			})
		}
	}

	logger.SyncStart(absPath, "", isFolder)

	remoteID, err := e.Provider.Create(uploadFiles, public)
	if err != nil {
		logger.SyncError(absPath, err.Error())
		return err
	}

	logger.SyncSuccess(absPath, remoteID, hash, isFolder, "github", public)
	return e.Repo.AddMapping(domain.Mapping{
		LocalPath:      absPath,
		RemoteID:       remoteID,
		Provider:       "github",
		IsFolder:       isFolder,
		Public:         public,
		LastSyncedHash: hash,
	})
}

func (e *Engine) InitialSyncWithVisibility(absPath string, public bool) error {
	return e.initialSync(absPath, public)
}

func (e *Engine) SetVisibility(path string, public bool) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	var oldRemoteID string
	var newRemoteID string

	err = e.Repo.WithLock(func(state *domain.State) error {
		mapping := e.findMappingInState(state, absPath)
		if mapping == nil {
			return fmt.Errorf("path %s is not tracked — use 'gistsync sync %s' to start tracking it", path, path)
		}

		if mapping.Public == public {
			return nil
		}

		oldRemoteID = mapping.RemoteID

		remoteFiles, fetchErr := e.Provider.Fetch(oldRemoteID)
		if fetchErr != nil {
			return fmt.Errorf("failed to fetch remote content: %w", fetchErr)
		}

		var localFiles []domain.File
		var readErr error
		if mapping.IsFolder {
			localFiles, readErr = e.ReadLocalDir(absPath)
		} else {
			content, err := os.ReadFile(absPath)
			if err == nil {
				localFiles = []domain.File{{
					Path:    filepath.Base(absPath),
					Content: content,
					Hash:    ComputeHash(content),
				}}
			}
			readErr = err
		}
		if readErr != nil {
			return fmt.Errorf("failed to read local content: %w", readErr)
		}

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
		if action == domain.ActionConflict {
			return &ConflictError{
				LocalHash:      localHash,
				RemoteHash:     remoteHash,
				LastSyncedHash: mapping.LastSyncedHash,
			}
		}

		targetFiles := remoteFiles
		newSyncedHash := remoteHash
		if action == domain.ActionPush {
			targetFiles = localFiles
			newSyncedHash = localHash
		}

		var createErr error
		newRemoteID, createErr = e.Provider.Create(targetFiles, public)
		if createErr != nil {
			return fmt.Errorf("failed to create new gist: %w", createErr)
		}

		mapping.RemoteID = newRemoteID
		mapping.Public = public
		mapping.LastSyncedHash = newSyncedHash
		return nil
	})

	if err != nil {
		if newRemoteID != "" {
			_ = e.Provider.Delete(newRemoteID)
		}
		return err
	}

	if oldRemoteID != "" {
		if err := e.Provider.Delete(oldRemoteID); err != nil {
			ui.Warning("GistDeleteWarning", map[string]interface{}{"ID": oldRemoteID, "Err": err})
		}
	}

	return nil
}

func (e *Engine) Status(localPath string) (domain.SyncAction, error) {
	absPath, err := filepath.Abs(localPath)
	if err != nil {
		return "", err
	}

	mapping, err := e.Repo.GetMapping(absPath)
	if err != nil {
		return "", err
	}
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
		remoteHash = e.computeRemoteDirHash(absPath, remoteFiles)
	} else {
		if len(remoteFiles) == 0 {
			return "REMOTE_DELETED", nil
		}
		remoteHash = remoteFiles[0].Hash
	}

	action := DetermineAction(localHash, remoteHash, mapping.LastSyncedHash)
	action = e.handleConfigDirSpecialCase(action, absPath, remoteFiles, localHash)

	if action == domain.ActionConflict {
		return domain.SyncAction(fmt.Sprintf("CONFLICT (Local: %s, Remote: %s, LastSynced: %s)",
			truncHash(localHash), truncHash(remoteHash), truncHash(mapping.LastSyncedHash))), nil
	}
	return action, nil
}
