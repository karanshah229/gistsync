package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/karanshah229/gistsync/internal/domain"
)

// RepairResult represents the result of a single mapping repair
type RepairResult struct {
	OldPath string
	NewPath string
	Status  string // "REPAIRED", "MISSING", "SKIPPED", "VALID"
}

// RepairConfig scans mappings and attempts to fix paths for the current OS
func RepairConfig(state *domain.State) ([]RepairResult, bool, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, false, fmt.Errorf("failed to get home directory: %w", err)
	}

	// regex for common home directory patterns:
	// Mac: /Users/username/
	// Linux: /home/username/
	// Windows: C:\Users\username\
	homeRegex := regexp.MustCompile(`^(/Users/[^/]+/|/home/[^/]+/|[a-zA-Z]:\\Users\\[^\\]+\\)`)

	var results []RepairResult
	modified := false

	for i := range state.Mappings {
		m := &state.Mappings[i]
		res := RepairResult{OldPath: m.LocalPath}

		// 1. Check if existing path is valid
		if _, err := os.Stat(m.LocalPath); err == nil {
			res.Status = "VALID"
			res.NewPath = m.LocalPath
			results = append(results, res)
			continue
		}

		// 2. Try to remap if it looks like a home-dir path
		match := homeRegex.FindString(m.LocalPath)
		if match != "" {
			// Extract the relative part after the home dir prefix
			rel := m.LocalPath[len(match):]
			// Normalize separators for current OS
			rel = filepath.FromSlash(strings.ReplaceAll(rel, "\\", "/"))
			
			newPath := filepath.Join(home, rel)
			if _, err := os.Stat(newPath); err == nil {
				res.NewPath = newPath
				res.Status = "REPAIRED"
				m.LocalPath = newPath
				modified = true
			} else {
				res.NewPath = newPath
				res.Status = "MISSING"
			}
		} else {
			res.Status = "SKIPPED"
		}
		
		results = append(results, res)
	}

	return results, modified, nil
}
