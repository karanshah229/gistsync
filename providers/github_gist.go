package providers

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/karanshah229/gistsync/internal/domain"
)

const gistPathSeparator = "---"

type GitHubProvider struct{}

func NewGitHubProvider() *GitHubProvider {
	return &GitHubProvider{}
}

// Ensure GitHubProvider implements domain.Provider
var _ domain.Provider = (*GitHubProvider)(nil)

func (p *GitHubProvider) Create(files []domain.File, public bool) (string, error) {
	filtered := p.filterEmptyFiles(files)
	if len(filtered) == 0 {
		return "", fmt.Errorf("cannot sync: all provided files are empty, and GitHub Gists do not support blank files")
	}

	payload := map[string]interface{}{
		"public": public,
		"files":  p.makeFilesMap(filtered),
	}
	
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal gist create payload: %w", err)
	}

	cmd := exec.Command("gh", "api", "gists", "--input", "-")
	cmd.Stdin = strings.NewReader(string(jsonData))
	
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to create gist: %w (output: %s)", err, string(out))
	}

	var res struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(out, &res); err != nil {
		return "", fmt.Errorf("failed to parse gist create response: %w", err)
	}

	return res.ID, nil
}

func (p *GitHubProvider) Update(remoteID string, files []domain.File) error {
	filtered := p.filterEmptyFiles(files)
	if len(filtered) == 0 {
		return nil
	}

	payload := map[string]interface{}{
		"files": p.makeFilesMap(filtered),
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal gist update payload: %w", err)
	}

	cmd := exec.Command("gh", "api", "-X", "PATCH", fmt.Sprintf("gists/%s", remoteID), "--input", "-")
	cmd.Stdin = strings.NewReader(string(jsonData))
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to update gist %s: %w (output: %s)", remoteID, err, string(out))
	}
	return nil
}

func (p *GitHubProvider) filterEmptyFiles(files []domain.File) []domain.File {
	var filtered []domain.File
	for _, f := range files {
		if len(strings.TrimSpace(string(f.Content))) > 0 {
			filtered = append(filtered, f)
		}
	}
	return filtered
}

func (p *GitHubProvider) makeFilesMap(files []domain.File) map[string]interface{} {
	filesMap := make(map[string]interface{})
	for _, f := range files {
		// Flatten path for Gist API (no slashes allowed in creation)
		flatPath := strings.ReplaceAll(f.Path, "/", gistPathSeparator)
		filesMap[flatPath] = map[string]string{
			"content": string(f.Content),
		}
	}
	return filesMap
}

func (p *GitHubProvider) Fetch(remoteID string) ([]domain.File, error) {
	cmd := exec.Command("gh", "api", fmt.Sprintf("gists/%s", remoteID))
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch gist %s using api: %w (output: %s)", remoteID, err, string(out))
	}

	var res struct {
		Files map[string]struct {
			Filename string `json:"filename"`
			Content  string `json:"content"`
		} `json:"files"`
	}

	if err := json.Unmarshal(out, &res); err != nil {
		return nil, fmt.Errorf("failed to unmarshal gist api response: %w", err)
	}

	var files []domain.File
	for _, f := range res.Files {
		// Unflatten path from Gist API
		origPath := strings.ReplaceAll(f.Filename, gistPathSeparator, "/")
		files = append(files, domain.File{
			Path:    origPath,
			Content: []byte(f.Content),
			Hash:    domain.ComputeHash([]byte(f.Content)),
		})
	}

	return files, nil
}


func (p *GitHubProvider) Delete(remoteID string) error {
	cmd := exec.Command("gh", "gist", "delete", remoteID, "--yes")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to delete gist: %w (output: %s)", err, string(out))
	}
	return nil
}

func (p *GitHubProvider) CheckRateLimit() (int, time.Time, error) {
	cmd := exec.Command("gh", "api", "rate_limit")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return 0, time.Time{}, fmt.Errorf("failed to check rate limit: %w (output: %s)", err, string(out))
	}

	var res struct {
		Resources struct {
			Core struct {
				Remaining int   `json:"remaining"`
				Reset     int64 `json:"reset"`
			} `json:"core"`
		} `json:"resources"`
	}

	if err := json.Unmarshal(out, &res); err != nil {
		return 0, time.Time{}, err
	}

	return res.Resources.Core.Remaining, time.Unix(res.Resources.Core.Reset, 0), nil
}

func (p *GitHubProvider) List() ([]domain.GistInfo, error) {
	cmd := exec.Command("gh", "api", "gists")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to list gists using api: %w (output: %s)", err, string(out))
	}

	var res []struct {
		ID          string `json:"id"`
		Description string `json:"description"`
		UpdatedAt   string `json:"updated_at"`
		Files       map[string]interface{} `json:"files"`
	}

	if err := json.Unmarshal(out, &res); err != nil {
		return nil, fmt.Errorf("failed to unmarshal gists api response: %w", err)
	}

	var infos []domain.GistInfo
	for _, g := range res {
		updatedAt, _ := time.Parse(time.RFC3339, g.UpdatedAt)
		
		var files []string
		for filename := range g.Files {
			// Unflatten path for display/logic
			origPath := strings.ReplaceAll(filename, gistPathSeparator, "/")
			files = append(files, origPath)
		}

		infos = append(infos, domain.GistInfo{
			ID:          g.ID,
			Description: g.Description,
			UpdatedAt:   updatedAt,
			Files:       files,
		})
	}

	return infos, nil
}

func (p *GitHubProvider) Verify() (bool, string, error) {
	// 1. Check if 'gh' is in PATH
	_, err := exec.LookPath("gh")
	if err != nil {
		return false, "GitHub CLI (gh) not found in PATH", nil
	}

	// 2. Check auth status
	cmd := exec.Command("gh", "auth", "status")
	out, err := cmd.CombinedOutput()
	output := string(out)
	if err != nil {
		return false, "GitHub CLI is not authenticated. Please run 'gh auth login'.", nil
	}

	return true, output, nil
}
