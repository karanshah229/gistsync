package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/karanshah229/gistsync/core"
	"github.com/karanshah229/gistsync/internal/logger"
	"github.com/karanshah229/gistsync/internal/storage"
	"github.com/karanshah229/gistsync/pkg/ui"
	"github.com/spf13/cobra"
)

type walEntry struct {
	Time  time.Time              `json:"time"`
	Type  string                 `json:"type"`
	Data  map[string]interface{} `json:"data"`
	Level string                 `json:"level"`
	Msg   string                 `json:"msg"`
}

var recoverCmd = &cobra.Command{
	Use:   "recover",
	Short: "Recover state.json from log files (WAL replay)",
	Long:  `Scans the logs directory for JSON-formatted logs and reconstructs the state.json file by replaying successful sync events.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		logDir, err := storage.GetLogDir()
		if err != nil {
			return fmt.Errorf("failed to get log directory for recovery: %w", err)
		}

		files, err := os.ReadDir(logDir)
		if err != nil {
			return fmt.Errorf("failed to scan log directory: %w", err)
		}

		var entries []walEntry
		ui.Info("RecoverScanning", nil)

		for _, f := range files {
			if f.IsDir() || !strings.HasSuffix(f.Name(), ".log") {
				continue
			}

			path := filepath.Join(logDir, f.Name())
			file, err := os.Open(path)
			if err != nil {
				ui.Warning("RecoverLogOpenFailed", map[string]interface{}{"File": f.Name()})
				continue
			}

			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				line := scanner.Bytes()
				var raw map[string]interface{}
				if err := json.Unmarshal(line, &raw); err != nil {
					continue
				}

				entryType, _ := raw["type"].(string)
				if entryType == "" {
					continue
				}

				// Parse time
				var entryTime time.Time
				if tStr, ok := raw["time"].(string); ok {
					entryTime, _ = time.Parse(time.RFC3339Nano, tStr)
				}

				entries = append(entries, walEntry{
					Time: entryTime,
					Type: entryType,
					Data: raw,
				})
			}
			file.Close()
		}

		if len(entries) == 0 {
			ui.Warning("RecoverNoHistory", nil)
			return nil
		}

		// Sort by time ascending
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].Time.Before(entries[j].Time)
		})

		// 1. Baseline Initialization: Load existing state
		mappings := make(map[string]core.Mapping)
		state, err := core.LoadState()
		if err == nil && state != nil {
			for _, m := range state.Mappings {
				mappings[m.LocalPath] = m
			}
		} else {
			state = &core.State{} // Empty state if missing
		}
		initialCount := len(mappings)

		// 2. HWM (High Water Mark) Pruning: Find the latest committed or recovered point
		var hwm time.Time
		if initialCount > 0 {
			for i := len(entries) - 1; i >= 0; i-- {
				if entries[i].Type == logger.TypeRecoveryComplete || entries[i].Type == logger.TypeCheckpoint {
					// We trust any Checkpoint (local save) or RecoveryComplete as an HWM
					hwm = entries[i].Time
					break
				}
			}
		}

		// Group entries by PID (only those after HWM)
		pidGroups := make(map[int][]walEntry)
		for _, entry := range entries {
			if !hwm.IsZero() && !entry.Time.After(hwm) {
				continue
			}
			pid, _ := entry.Data["pid"].(float64)
			pidGroups[int(pid)] = append(pidGroups[int(pid)], entry)
		}

		if len(pidGroups) == 0 {
			ui.Success("RecoverUpToDate", nil)
			return nil
		}

		// Sort PIDs chronologically
		var pids []int
		for p := range pidGroups {
			pids = append(pids, p)
		}
		sort.Ints(pids)

		changesFound := false
		for _, pid := range pids {
			group := pidGroups[pid]
			recentRemoteID := make(map[string]string)

			for i, entry := range group {
				localPath, _ := entry.Data["local_path"].(string)
				if localPath == "" {
					continue
				}

				if entry.Type == logger.TypeSyncStart {
					if rid, ok := entry.Data["remote_id"].(string); ok && rid != "" {
						recentRemoteID[localPath] = rid
					} else if gid, ok := entry.Data["gist_id"].(string); ok && gid != "" {
						recentRemoteID[localPath] = gid
					}
				}

				if entry.Type == logger.TypeSyncSuccess {
					hasCheckpoint := false
					for j := i + 1; j < len(group); j++ {
						if group[j].Type == logger.TypeCheckpoint {
							hasCheckpoint = true
							break
						}
						if group[j].Type == logger.TypeSyncStart || group[j].Type == logger.TypeSyncSuccess {
							if p, _ := group[j].Data["local_path"].(string); p == localPath {
								break
							}
						}
					}

					if !hasCheckpoint {
						ui.Warning("RecoverInterrupted", map[string]interface{}{"Path": localPath})
					}

					remoteID, _ := entry.Data["remote_id"].(string)
					if remoteID == "" {
						remoteID = recentRemoteID[localPath]
					}

					hash, _ := entry.Data["hash"].(string)
					provider, _ := entry.Data["provider"].(string)
					public, _ := entry.Data["public"].(bool)
					isFolder, ok := entry.Data["is_folder"].(bool)

					if !ok {
						info, err := os.Stat(localPath)
						if err == nil {
							isFolder = info.IsDir()
						}
					}
					if provider == "" {
						provider = "github"
					}

					newMapping := core.Mapping{
						LocalPath:      localPath,
						RemoteID:       remoteID,
						LastSyncedHash: hash,
						IsFolder:       isFolder,
						Provider:       provider,
						Public:         public,
					}

					if old, exists := mappings[localPath]; !exists || old != newMapping {
						mappings[localPath] = newMapping
						changesFound = true
					}
				}
			}
		}

		if !changesFound {
			ui.Success("RecoverAllExist", nil)
			// Log a completion to move the HWM past these redundant entries
			logger.Log.Info("Recovery skipped: no changes", slog.String("type", logger.TypeRecoveryComplete))
			return nil
		}

		// 3. Save Baseline + Applied Changes
		state.Mappings = make([]core.Mapping, 0, len(mappings))
		for _, m := range mappings {
			state.Mappings = append(state.Mappings, m)
		}

		if err := state.Save(); err != nil {
			return fmt.Errorf("failed to save recovered state: %w", err)
		}

		addedCount := len(mappings) - initialCount
		if addedCount > 0 {
			ui.Success("RecoverNewMappings", map[string]interface{}{"Count": addedCount})
		} else {
			ui.Success("RecoverUpdated", nil)
		}
		
		logger.Log.Info("Recovery complete", slog.String("type", logger.TypeRecoveryComplete), slog.Int("count", len(state.Mappings)))
		logger.Checkpoint("State recovered from WAL")
		
		return nil
	},
}

func init() {
	rootCmd.AddCommand(recoverCmd)
}
