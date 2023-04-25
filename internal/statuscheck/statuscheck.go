package statuscheck

// StatusChecker is the interface for interacting with status checks.
type StatusChecker interface {

	// Send pushes the status check.
	Send(reportPath string, target string, token string) error

	// Write creates the reports file.
	Write(path string, coverage float64, threshold float64) error

	// WriteNoop sends a noop check.
	WriteNoop(path string) error
}
