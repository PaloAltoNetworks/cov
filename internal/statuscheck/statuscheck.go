package statuscheck

// StatusChecker is the interface for interacting with status checks.
type StatusChecker interface {

	// Send pushes the status check.
	Send(reportPath string, target string, token string) error

	// Write creates the reports file.
	Write(path string, coverage int, threshold int) error

	// WriteNoop sends a noop check.
	WriteNoop(path string) error
}
