package providers

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/karanshah229/gistsync/core"
)

type GitHubProvider struct{}

func NewGitHubProvider() *GitHubProvider {
	return &GitHubProvider{}
}

// Ensure GitHubProvider implements core.Provider
var _ core.Provider = (*GitHubProvider)(nil)

func (p *GitHubProvider) Create(files []core.File) (string, error) {
	if len(files) == 0 {
		return "", fmt.Errorf("no files to create gist with")
	}

	cmd := exec.Command("gh", "gist", "create", "-", "-f", files[0].Path)
	cmd.Stdin = strings.NewReader(string(files[0].Content))
	
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to create gist: %w (output: %s)", err, string(out))
	}

	gistURL := strings.TrimSpace(string(out))
	parts := strings.Split(gistURL, "/")
	id := parts[len(parts)-1]

	if len(files) > 1 {
		if err := p.Update(id, files[1:]); err != nil {
			return id, err
		}
	}

	return id, nil
}

func (p *GitHubProvider) Update(remoteID string, files []core.File) error {
	if len(files) == 0 {
		return nil
	}

	args := []string{"api", "-X", "PATCH", fmt.Sprintf("gists/%s", remoteID)}
	for _, f := range files {
		args = append(args, "-F", fmt.Sprintf("files[%s][content]=%s", f.Path, string(f.Content)))
	}

	cmd := exec.Command("gh", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to update gist %s using api: %w (output: %s)", remoteID, err, string(out))
	}
	return nil
}

func (p *GitHubProvider) Fetch(remoteID string) ([]core.File, error) {
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

	var files []core.File
	for _, f := range res.Files {
		files = append(files, core.File{
			Path:    f.Filename,
			Content: []byte(f.Content),
			Hash:    core.ComputeHash([]byte(f.Content)),
		})
	}

	return files, nil
}


func (p *GitHubProvider) Delete(remoteID string) error {
	cmd := exec.Command("gh", "gist", "delete", remoteID, "-y")
	return cmd.Run()
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

