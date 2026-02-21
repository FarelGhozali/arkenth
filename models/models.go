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
