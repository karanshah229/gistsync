package sync

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/karanshah229/gistsync/core"
	"github.com/karanshah229/gistsync/internal"
	"github.com/karanshah229/gistsync/internal/domain"
	"github.com/karanshah229/gistsync/internal/state"
	"github.com/karanshah229/gistsync/internal/storage"
	"github.com/karanshah229/gistsync/pkg/i18n"
	"github.com/karanshah229/gistsync/pkg/ui"
	"github.com/karanshah229/gistsync/providers"
)

// SyncManager orchestrates sync operations
type SyncManager struct {
	Repo    state.Repository
	Config  *internal.Config
	Version string
}

// NewSyncManager creates a new SyncManager
func NewSyncManager(version string) (*SyncManager, error) {
	repo, err := state.NewFileRepository()
	if err != nil {
		return nil, err
	}

	cfg, err := internal.LoadConfig()
	if err != nil {
		cfg = &internal.Config{DefaultProvider: "github"}
	}

	return &SyncManager{
		Repo:    repo,
		Config:  cfg,
		Version: version,
	}, nil
}

// GetProvider returns a provider by name or the default one
func (m *SyncManager) GetProvider(name string) (domain.Provider, error) {
	if name == "" {
		name = m.Config.DefaultProvider
	}
	if name == "" {
		name = "github"
	}

	switch name {
	case "gitlab":
		// return providers.NewGitLabProvider(), nil
		return nil, fmt.Errorf("gitlab provider not yet implemented in refactored version")
	case "github":
		return providers.NewGitHubProvider(), nil
	default:
		return nil, fmt.Errorf("unknown provider: %s", name)
	}
}

// GetEngine returns a new Engine for the given provider
func (m *SyncManager) GetEngine(providerName string) (*core.Engine, error) {
	p, err := m.GetProvider(providerName)
	if err != nil {
		return nil, err
	}
	return core.NewEngine(m.Repo, p), nil
}

// ConfigSync syncs the configuration directory
func (m *SyncManager) ConfigSync(providerName string) error {
	configDir, err := storage.GetConfigDir()
	if err != nil {
		return err
	}

	// Always private for config
	return m.SyncPath(configDir, providerName, "", false)
}

// SyncAll syncs all tracked mappings
func (m *SyncManager) SyncAll(providerName string) error {
	engine, err := m.GetEngine(providerName)
	if err != nil {
		return err
	}

	state, err := m.Repo.Load()
	if err != nil {
		return err
	}

	var success, failed int
	total := len(state.Mappings)

	for _, mapping := range state.Mappings {
		var action domain.SyncAction
		var syncErr error

		if mapping.IsFolder {
			action, syncErr = engine.SyncDir(mapping.LocalPath)
		} else {
			action, syncErr = engine.SyncFile(mapping.LocalPath)
		}

		if syncErr != nil {
			ui.Error("SyncFailed", map[string]interface{}{"Path": mapping.LocalPath, "Err": syncErr})
			failed++
			continue
		}

		success++
		switch action {
		case domain.ActionNoop:
			ui.Print("SyncNoop", map[string]interface{}{"Path": mapping.LocalPath})
		case domain.ActionPush:
			ui.Success("SyncPushed", map[string]interface{}{"Path": mapping.LocalPath})
		case domain.ActionPull:
			ui.Success("SyncPulled", map[string]interface{}{"Path": mapping.LocalPath})
		}
	}

	if total > 0 {
		ui.Success("SyncAllSummary", map[string]interface{}{
			"Success": success,
			"Failed":  failed,
			"Total":   total,
		})
	}

	return nil
}

// SyncPath syncs a specific path
func (m *SyncManager) SyncPath(path string, providerName string, gistID string, isPublic bool) error {
	engine, err := m.GetEngine(providerName)
	if err != nil {
		return err
	}

	absPath, err := core.GetAbsPath(path)
	if err != nil {
		return err
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("%s", i18n.T("PathNotFoundError", map[string]interface{}{"Path": absPath}))
	}

	mapping, err := m.Repo.GetMapping(absPath)
	if err != nil {
		return err
	}

	if gistID != "" {
		// Manual mapping
		err := m.handleManualMapping(engine, absPath, gistID, providerName)
		if err != nil {
			return err
		}
		// Re-fetch mapping after manual mapping
		mapping, err = m.Repo.GetMapping(absPath)
		if err != nil {
			return err
		}
	}

	if mapping == nil {
		// New sync
		ui.Print("InitialSyncStart", map[string]interface{}{"Path": absPath})
		mapping, err = engine.InitialSyncWithVisibility(absPath, isPublic)
		if err != nil {
			return err
		}
		visibility := "private"
		if mapping.Public {
			visibility = "public"
		}
		ui.Success("InitialSyncSuccess", map[string]interface{}{
			"Path":       mapping.LocalPath,
			"Visibility": visibility,
			"ID":         mapping.RemoteID,
		})
		return nil
	}

	// Regular sync
	info, err := os.Stat(absPath)
	if err != nil {
		return err
	}

	var action domain.SyncAction
	if info.IsDir() {
		action, err = engine.SyncDir(absPath)
	} else {
		action, err = engine.SyncFile(absPath)
	}

	if err != nil {
		if _, ok := err.(*core.ConflictError); ok {
			ui.Error("ConflictDetected", map[string]interface{}{"Err": err})
			return err
		}
		return err
	}

	switch action {
	case domain.ActionNoop:
		ui.Print("SyncNoop", map[string]interface{}{"Path": absPath})
	case domain.ActionPush:
		ui.Success("SyncPushed", map[string]interface{}{"Path": absPath})
	case domain.ActionPull:
		ui.Success("SyncPulled", map[string]interface{}{"Path": absPath})
	}

	return nil
}

// Status shows the sync status for a path or all tracked paths
func (m *SyncManager) Status(path string, providerName string) error {
	engine, err := m.GetEngine(providerName)
	if err != nil {
		return err
	}

	var paths []string
	if path != "" {
		abs, err := core.GetAbsPath(path)
		if err != nil {
			return err
		}
		paths = append(paths, abs)
	} else {
		state, err := m.Repo.Load()
		if err != nil {
			return err
		}
		for _, mapping := range state.Mappings {
			paths = append(paths, mapping.LocalPath)
		}
	}

	if len(paths) == 0 {
		ui.Print("NoFilesTracked", nil)
		return nil
	}

	for _, p := range paths {
		status, err := engine.Status(p)
		if err != nil {
			ui.Print("StatusError", map[string]interface{}{"Path": p, "Err": err})
			continue
		}
		ui.Print("StatusLine", map[string]interface{}{"Path": p, "Status": status})
	}
	return nil
}

// SetVisibility changes the visibility of a gist
func (m *SyncManager) SetVisibility(path string, public bool, providerName string) error {
	absPath, err := core.GetAbsPath(path)
	if err != nil {
		return err
	}

	mapping, err := m.Repo.GetMapping(absPath)
	if err != nil {
		return err
	}
	if mapping == nil {
		return fmt.Errorf("%s", i18n.T("PathNotTrackedHint", map[string]interface{}{"Path": absPath}))
	}

	visibility := "private"
	if public {
		visibility = "public"
	}

	if mapping.Public == public {
		ui.Print("VisibilityAlreadySet", map[string]interface{}{"Path": absPath, "Visibility": visibility})
		return nil
	}

	engine, err := m.GetEngine(providerName)
	if err != nil {
		return err
	}

	ui.Print("ChangingVisibility", map[string]interface{}{"Path": absPath, "Visibility": visibility})

	if err := engine.SetVisibility(absPath, public); err != nil {
		return err
	}

	ui.Success("VisibilityChangeSuccess", map[string]interface{}{"Visibility": visibility})
	return nil
}

func (m *SyncManager) handleManualMapping(engine *core.Engine, absPath string, gistID string, providerName string) error {
	info, err := os.Stat(absPath)
	if err != nil {
		return err
	}

	mapping, err := m.Repo.GetMapping(absPath)
	if err != nil {
		return err
	}

	if mapping != nil {
		confirm := ui.Confirm("ConfirmQuestion", map[string]interface{}{
			"Message": fmt.Sprintf("Path %s is already tracked (Gist %s). Overwrite with Gist %s?", absPath, mapping.RemoteID, gistID),
		})
		if !confirm {
			return nil
		}
	}

	ui.Print("ManualMappingStart", map[string]interface{}{
		"Path":     absPath,
		"ID":       gistID,
		"Provider": providerName,
	})
	return m.Repo.AddMapping(domain.Mapping{
		LocalPath: absPath,
		RemoteID:  gistID,
		Provider:  providerName,
		IsFolder:  info.IsDir(),
		Public:    false,
	})
}

// InitializeState initializes the state.json file
func (m *SyncManager) InitializeState() error {
	statePath, err := storage.GetStateFilePath()
	if err != nil {
		return err
	}

	initialState := domain.State{
		Version:  m.Version,
		Mappings: []domain.Mapping{},
	}

	data, err := json.MarshalIndent(initialState, "", "  ")
	if err != nil {
		return err
	}

	return storage.WriteAtomic(statePath, data)
}

// RestoreConfig restores configuration from a provider gist
func (m *SyncManager) RestoreConfig(providerName string, gistID string) error {
	p, err := m.GetProvider(providerName)
	if err != nil {
		return err
	}

	ui.Print("DownloadingBackup", map[string]interface{}{"ID": gistID})
	files, err := p.Fetch(gistID)
	if err != nil {
		return fmt.Errorf("failed to fetch backup files: %w", err)
	}

	configDir, err := storage.GetConfigDir()
	if err != nil {
		return err
	}

	for _, f := range files {
		if f.Path == "config.json" || f.Path == "state.json" {
			content := f.Content
			if f.Path == "config.json" {
				content = internal.ValidateAndCleanConfig(content)
			}

			target := filepath.Join(configDir, f.Path)

			// Fix "PENDING" remote_id in state.json during restoration
			if f.Path == "state.json" {
				var state domain.State
				if err := json.Unmarshal(content, &state); err == nil {
					modified := false
					for i := range state.Mappings {
						if state.Mappings[i].RemoteID == "PENDING" {
							state.Mappings[i].RemoteID = gistID
							modified = true
						}
					}
					if modified {
						if newContent, err := json.MarshalIndent(state, "", "  "); err == nil {
							content = newContent
						}
					}
				}
			}

			if err := storage.WriteAtomic(target, content); err != nil {
				return fmt.Errorf("failed to write %s: %w", f.Path, err)
			}
			ui.Success("RestoredFile", map[string]interface{}{"Path": f.Path})
		}
	}

	return nil
}

// BackupConfig backs up the current configuration to the default provider
func (m *SyncManager) BackupConfig(providerName string) error {
	engine, err := m.GetEngine(providerName)
	if err != nil {
		return err
	}

	configDir, err := storage.GetConfigDir()
	if err != nil {
		return err
	}

	ui.Print("BackingUpConfig", nil)
	_, err = engine.SyncDir(configDir)
	return err
}

// ListBackups returns a list of gists that look like gistsync backups
func (m *SyncManager) ListBackups(providerName string) ([]domain.GistInfo, error) {
	p, err := m.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	gists, err := p.List()
	if err != nil {
		return nil, err
	}

	var candidates []domain.GistInfo
	for _, g := range gists {
		hasConfig := false
		hasState := false
		for _, f := range g.Files {
			if f == "config.json" {
				hasConfig = true
			}
			if f == "state.json" {
				hasState = true
			}
		}
		if hasConfig && hasState {
			candidates = append(candidates, g)
		}
	}
	return candidates, nil
}

// RemovePath stops tracking a local path
func (m *SyncManager) RemovePath(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	return m.Repo.WithLock(func(st *domain.State) error {
		newMappings := []domain.Mapping{}
		found := false
		for _, mapping := range st.Mappings {
			if mapping.LocalPath == absPath {
				found = true
				continue
			}
			newMappings = append(newMappings, mapping)
		}

		if !found {
			return fmt.Errorf("path not tracked: %s", absPath)
		}

		st.Mappings = newMappings
		return nil
	})
}
