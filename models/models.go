package models

// Report is the top-level structure containing all audit results.
type Report struct {
	TargetURL        string           `json:"target_url"`
	MobileEmulation  bool             `json:"mobile_emulation"`
	ScannedURLs      []string         `json:"scanned_urls"`
	NetworkAnomalies []NetworkAnomaly `json:"network_anomalies"`
	CriticalBugs     []Bug            `json:"critical_bugs"`
	UIUXIssues       []Bug            `json:"ui_ux_issues"`
}

func NewReport(target string, mobile bool) *Report {
	return &Report{
		TargetURL:        target,
		MobileEmulation:  mobile,
		ScannedURLs:      []string{},
		NetworkAnomalies: []NetworkAnomaly{},
		CriticalBugs:     []Bug{},
		UIUXIssues:       []Bug{},
	}
}

// NetworkAnomaly tracks HTTP 4xx, 5xx, or request timeouts
type NetworkAnomaly struct {
	URL      string `json:"url"`
	Method   string `json:"method"`
	Status   int    `json:"status"`
	ErrorMsg string `json:"error_msg"`
	Payload  string `json:"payload,omitempty"` // For POST/PUT
}

// Bug represents an actionable flaw found during fuzzing or interaction
type Bug struct {
	Severity    string `json:"severity"` // "CRITICAL", "HIGH", "MEDIUM"
	URL         string `json:"url"`
	ActionTaken string `json:"action_taken"`
	Expected    string `json:"expected"`
	Actual      string `json:"actual"`
	ProofPath   string `json:"proof_path,omitempty"` // Local path to screenshot/video
}

// VisualMask defines an ignore-region for Smart Visual Regression.
// When comparing screenshots, masked areas are filled with a solid color
// before diffing so that dynamic content (ads, timestamps, chat widgets)
// does not trigger false-positive regressions.
type VisualMask struct {
	ID        int    `json:"id"`
	TargetURL string `json:"target_url"` // URL of the page this mask applies to
	X         int    `json:"x"`
	Y         int    `json:"y"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	Label     string `json:"label"` // Human-readable label, e.g. "Ad Banner"
}
