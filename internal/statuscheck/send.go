package statuscheck

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

type githubStatus struct {
	State       string `json:"state"`
	TargetURL   string `json:"target_url"`
	Description string `json:"description"`
	Context     string `json:"context"`
}

// Send sends the status check to github.
func Send(reportPath string, target string, token string) error {

	status := githubStatus{}
	data, err := os.ReadFile(reportPath)
	if err != nil {
		return fmt.Errorf("unable to read report file: %w", err)
	}

	if err := json.Unmarshal(data, &status); err != nil {
		return fmt.Errorf("unable to unmarshal data file: %w", err)
	}

	return send(status, target, token)
}

// Write writes the reports file.
func Write(path string, coverage int, threshold int) error {

	status := githubStatus{
		Context: "cov",
		State: func() string {
			if coverage >= threshold {
				return "success"
			}
			return "failure"
		}(),
		Description: func() string {
			info := fmt.Sprintf("%d%% / %d%%", coverage, threshold)
			if coverage >= threshold {
				return fmt.Sprintf("success %s", info)
			}
			return fmt.Sprintf("failure %s", info)
		}(),
	}

	return write(status, path)
}

// SendNoop sends the noop check.
func WriteNoop(path string) error {

	status := githubStatus{
		Context:     "cov",
		State:       "success",
		Description: "no change in any go files",
	}

	return write(status, path)
}

func write(status githubStatus, path string) error {

	data, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		return fmt.Errorf("unable encode report status: %w", err)
	}

	return os.WriteFile(path, data, 0600)
}

func send(status githubStatus, target string, token string) error {

	parts := strings.SplitN(target, "@", 2)
	repo := parts[0]
	sha := parts[1]

	if token == "" {
		token = os.Getenv("GITHUB_TOKEN")
	}

	buf := bytes.NewBuffer(nil)
	if err := json.NewEncoder(buf).Encode(status); err != nil {
		return fmt.Errorf("unable to encode github status check: %w", err)
	}

	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("https://api.github.com/repos/%s/statuses/%s", repo, sha),
		buf,
	)
	if err != nil {
		return fmt.Errorf("unable to build request: %w", err)
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("unable to send request: %w", err)
	}
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("github rejected the request: %s", resp.Status)
	}

	return nil
}
