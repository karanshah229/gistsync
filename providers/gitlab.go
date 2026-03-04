package providers

import (
	"errors"
	"time"

	"github.com/karan/gistsync/core"
)

type GitLabProvider struct{}

func NewGitLabProvider() *GitLabProvider {
	return &GitLabProvider{}
}

var _ core.Provider = (*GitLabProvider)(nil)

func (p *GitLabProvider) Create(files []core.File) (string, error) {
	return "", errors.New("GitLab provider not implemented yet")
}

func (p *GitLabProvider) Update(remoteID string, files []core.File) error {
	return errors.New("GitLab provider not implemented yet")
}

func (p *GitLabProvider) Fetch(remoteID string) ([]core.File, error) {
	return nil, errors.New("GitLab provider not implemented yet")
}

func (p *GitLabProvider) Delete(remoteID string) error {
	return errors.New("GitLab provider not implemented yet")
}

func (p *GitLabProvider) CheckRateLimit() (int, time.Time, error) {
	return 1000, time.Now().Add(time.Hour), nil // Placeholder
}

