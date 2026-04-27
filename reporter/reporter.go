package reporter

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/FarelGhozali/arkenth/models"
)

type Reporter struct {
	Report   *models.Report
	ProofDir string
}

func NewReporter(rep *models.Report, proofDir string) *Reporter {
	return &Reporter{Report: rep, ProofDir: proofDir}
}

// GenerateOutputs drives the creation of JSON anomalies and MD Audit reports
func (r *Reporter) GenerateOutputs() error {
	// 1. Export Network Anomalies to JSON standalone
	if len(r.Report.NetworkAnomalies) > 0 {
		data, _ := json.MarshalIndent(r.Report.NetworkAnomalies, "", "  ")
		err := os.WriteFile("network_anomalies.json", data, 0644)
		if err != nil {
			return err
		}
	}

	// 2. Export Master QA Audit MD Report
	return r.generateAuditMarkdown("qa_audit_report.md")
}

func (r *Reporter) generateAuditMarkdown(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	if r.Report.MobileEmulation {
		f.WriteString(fmt.Sprintf("# 📱 Enterprise QA Audit: %s (Mobile Emulation)\n\n", r.Report.TargetURL))
	} else {
		f.WriteString(fmt.Sprintf("# 💻 Enterprise QA Audit: %s (Desktop)\n\n", r.Report.TargetURL))
	}

	f.WriteString("## 🚀 1. Execution Summary\n")
	f.WriteString(fmt.Sprintf("- **Total Pages Scanned**: `%d`\n", len(r.Report.ScannedURLs)))
	f.WriteString(fmt.Sprintf("- **Network Anomalies (4xx/5xx/Timeouts)**: `%d`\n", len(r.Report.NetworkAnomalies)))
	f.WriteString(fmt.Sprintf("- **Critical Bugs & Crashes**: `%d`\n", len(r.Report.CriticalBugs)))

	f.WriteString("\n## 🕷️ 2. Crawl Map (Pages Tested)\n")
	for _, url := range r.Report.ScannedURLs {
		f.WriteString(fmt.Sprintf("- %s\n", url))
	}

	f.WriteString("\n## 🚨 3. Critical Bugs (Crashes & Exceptions)\n")
	if len(r.Report.CriticalBugs) == 0 {
		f.WriteString("> ✅ No critical bugs or untrapped DOM crashes found during aggressive fuzzing.\n")
	} else {
		for i, bug := range r.Report.CriticalBugs {
			f.WriteString(fmt.Sprintf("\n### Bug #%d: [%s] at %s\n", i+1, bug.Severity, bug.URL))
			f.WriteString(fmt.Sprintf("- **Action**: %s\n", bug.ActionTaken))
			f.WriteString(fmt.Sprintf("- **Expected**: %s\n", bug.Expected))
			f.WriteString(fmt.Sprintf("- **Actual**: %s\n", bug.Actual))

			if bug.ProofPath != "" {
				relPath, _ := filepath.Rel(".", bug.ProofPath)
				f.WriteString(fmt.Sprintf("- **Proof**: ![%s](%s)\n", "Visual Evidence", relPath))
			}
		}
	}

	f.WriteString("\n## 📡 4. Network Anomalies (API & Asset Failures)\n")
	if len(r.Report.NetworkAnomalies) == 0 {
		f.WriteString("> ✅ All HTTP requests returned 200-3xx. No endpoints crashed during tests.\n")
	} else {
		f.WriteString("> ⚠️ Detailed payload and headers logged in `network_anomalies.json`.\n\n")

		limit := len(r.Report.NetworkAnomalies)
		if limit > 10 {
			limit = 10
		}
		for _, anom := range r.Report.NetworkAnomalies[:limit] {
			f.WriteString(fmt.Sprintf("- **%s** `%s` | Status: `%d` | Msg: *%s*\n", anom.Method, anom.URL, anom.Status, anom.ErrorMsg))
		}
		if len(r.Report.NetworkAnomalies) > 10 {
			f.WriteString(fmt.Sprintf("\n*... and %d more anomalies output to JSON.*\n", len(r.Report.NetworkAnomalies)-10))
		}
	}

	return nil
}
