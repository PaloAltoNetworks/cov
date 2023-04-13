package gitlabcheck

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/PaloAltoNetworks/cov/internal/statuscheck"
	"github.com/xanzy/go-gitlab"
)

type gitlabStatusCheck struct {
	State       string `json:"state"`
	TargetURL   string `json:"target_url"`
	Description string `json:"description"`
	Context     string `json:"context"`
}

// New returns a new statuc checker for GitLab.
func New() statuscheck.StatusChecker {

	return &gitlabStatusCheck{}
}

// Send sends the status check to gitlab.
func (s *gitlabStatusCheck) Send(reportPath string, target string, token string) error {

	data, err := os.ReadFile(reportPath)
	if err != nil {
		return fmt.Errorf("unable to read report file: %w", err)
	}

	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("unable to unmarshal data file: %w", err)
	}

	return s.send(target, token)
}

// Write writes the reports file.
func (s *gitlabStatusCheck) Write(path string, coverage int, threshold int) error {

	s.Context = "cov"
	s.State = func() string {
		if coverage >= threshold {
			return "success"
		}
		return "failure"
	}()
	s.Description = func() string {
		info := fmt.Sprintf("%d%% / %d%%", coverage, threshold)
		if coverage >= threshold {
			return fmt.Sprintf("success %s", info)
		}
		return fmt.Sprintf("failure %s", info)
	}()

	return s.write(path)
}

// SendNoop sends the noop check.
func (s *gitlabStatusCheck) WriteNoop(path string) error {

	s.Context = "cov"
	s.State = "success"
	s.Description = "no change in any go files"

	return s.write(path)
}

func (s *gitlabStatusCheck) write(path string) error {

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("unable encode report status: %w", err)
	}

	return os.WriteFile(path, data, 0600)
}

func (s *gitlabStatusCheck) send(target string, token string) error {

	parts := strings.SplitN(target, "@", 2)
	repo := parts[0]
	sha := parts[1]

	if token == "" {
		token = os.Getenv("GITLAB_TOKEN")
	}

	buf := bytes.NewBuffer(nil)
	if err := json.NewEncoder(buf).Encode(s); err != nil {
		return fmt.Errorf("unable to encode gitlab status check: %w", err)
	}

	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("https://git.scm.prismacloud.io/api/v4/projects/%s/statuses/%s",
			gitlab.PathEscape(repo),
			url.PathEscape(sha),
		),
		buf,
	)
	if err != nil {
		return fmt.Errorf("unable to build request: %w", err)
	}

	req.Header.Add("PRIVATE-TOKEN", token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("unable to send request: %w", err)
	}
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("gitlab rejected the request: %s", resp.Status)
	}

	return nil
}
