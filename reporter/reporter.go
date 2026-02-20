package reporter

import (
	"encoding/json"
	"fmt"
	"os"
)

type NetworkFinding struct {
	URL    string `json:"url"`
	Method string `json:"method"`
	Status int    `json:"status"`
	Error  string `json:"error,omitempty"`
}

type JSFinding struct {
	URL     string `json:"url"`
	Type    string `json:"type"`
	Message string `json:"message"`
}

type VulnerabilityFinding struct {
	URL      string `json:"url"`
	FormInfo string `json:"form_info"`
	Payload  string `json:"payload"`
	Error    string `json:"error"`
}

type Report struct {
	TargetURL       string                 `json:"target_url"`
	ScannedURLs     []string               `json:"scanned_urls"`
	NetworkFindings []NetworkFinding       `json:"network_findings"`
	JSFindings      []JSFinding            `json:"js_findings"`
	Vulnerabilities []VulnerabilityFinding `json:"vulnerabilities"`
}

func NewReport(target string) *Report {
	return &Report{
		TargetURL:       target,
		ScannedURLs:     []string{},
		NetworkFindings: []NetworkFinding{},
		JSFindings:      []JSFinding{},
		Vulnerabilities: []VulnerabilityFinding{},
	}
}

func (r *Report) AddScannedURL(url string) {
	r.ScannedURLs = append(r.ScannedURLs, url)
}

func (r *Report) AddNetworkFinding(f NetworkFinding) {
	r.NetworkFindings = append(r.NetworkFindings, f)
}

func (r *Report) AddJSFinding(f JSFinding) {
	r.JSFindings = append(r.JSFindings, f)
}

func (r *Report) AddVulnerability(f VulnerabilityFinding) {
	r.Vulnerabilities = append(r.Vulnerabilities, f)
}

func (r *Report) ExportJSON(filenamePrefix string) error {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filenamePrefix+".json", data, 0644)
}

func (r *Report) ExportMarkdown(filenamePrefix string) error {
	f, err := os.Create(filenamePrefix + ".md")
	if err != nil {
		return err
	}
	defer f.Close()

	f.WriteString(fmt.Sprintf("# QA Report for %s\n\n", r.TargetURL))
	f.WriteString(fmt.Sprintf("## Scanned URLs (%d total)\n", len(r.ScannedURLs)))
	for _, url := range r.ScannedURLs {
		f.WriteString(fmt.Sprintf("- %s\n", url))
	}

	f.WriteString(fmt.Sprintf("\n## Network Findings (%d total)\n", len(r.NetworkFindings)))
	if len(r.NetworkFindings) > 0 {
		for _, nf := range r.NetworkFindings {
			f.WriteString(fmt.Sprintf("- **%s** %s | Status: %d | Error: %s\n", nf.Method, nf.URL, nf.Status, nf.Error))
		}
	} else {
		f.WriteString("No network errors or slow responses detected.\n")
	}

	f.WriteString(fmt.Sprintf("\n## JS Console Findings (%d total)\n", len(r.JSFindings)))
	if len(r.JSFindings) > 0 {
		for _, js := range r.JSFindings {
			f.WriteString(fmt.Sprintf("- [%s] %s: %s\n", js.Type, js.URL, js.Message))
		}
	} else {
		f.WriteString("No JS errors or warnings detected.\n")
	}

	f.WriteString(fmt.Sprintf("\n## Vulnerabilities & Fuzzing Crashes (%d total)\n", len(r.Vulnerabilities)))
	if len(r.Vulnerabilities) > 0 {
		for _, v := range r.Vulnerabilities {
			f.WriteString(fmt.Sprintf("- **URL**: %s\n  **Form**: %s\n  **Payload**: %s\n  **Error**: %s\n\n", v.URL, v.FormInfo, v.Payload, v.Error))
		}
	} else {
		f.WriteString("No vulnerabilities or crashes detected during fuzzing.\n")
	}

	return nil
}
