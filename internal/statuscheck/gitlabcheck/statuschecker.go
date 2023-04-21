package gitlabcheck

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/PaloAltoNetworks/cov/internal/statuscheck"
	"github.com/google/go-querystring/query"
	"github.com/xanzy/go-gitlab"
)

type gitlabStatusCheck struct {
	Context     string `url:"context"`
	Description string `url:"description"`
	PipelineID  string `url:"pipeline_id,omitempty"`
	State       string `url:"state"`
	TargetURL   string `url:"target_url,omitempty"`

	hostURL string `url:"-"`
}

// New returns a new statuc checker for GitLab.
func New(hostURL string, targetURL string, pipelineID string) statuscheck.StatusChecker {

	return &gitlabStatusCheck{
		PipelineID: pipelineID,
		TargetURL:  targetURL,
		hostURL:    hostURL,
	}
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

	params, err := query.Values(s)
	if err != nil {
		return fmt.Errorf("unable to create params: %w", err)
	}

	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("%s/api/v4/projects/%s/statuses/%s?%s",
			s.hostURL,
			gitlab.PathEscape(repo),
			url.PathEscape(sha),
			params.Encode(),
		),
		nil,
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
