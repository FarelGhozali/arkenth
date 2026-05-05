package reporter

import (
	"os"
	"testing"

	"github.com/FarelGhozali/arkenth/models"
)

func TestGenerateOutputs(t *testing.T) {
	// Setup mock report
	rep := models.NewReport("https://example.com", false)
	rep.ScannedURLs = []string{"https://example.com", "https://example.com/about"}
	rep.NetworkAnomalies = []models.NetworkAnomaly{
		{URL: "https://example.com/api/fail", Method: "GET", Status: 500, ErrorMsg: "Internal Server Error"},
	}
	rep.CriticalBugs = []models.Bug{
		{Severity: "CRITICAL", URL: "https://example.com/login", ActionTaken: "Fuzz Form", Expected: "Blocked", Actual: "Crash"},
	}

	r := NewReporter(rep, "proofs")

	err := r.GenerateOutputs()
	if err != nil {
		t.Fatalf("GenerateOutputs failed: %v", err)
	}

	// Verify files were created
	jsonFile := "network_anomalies.json"
	mdFile := "qa_audit_report.md"

	if _, err := os.Stat(jsonFile); os.IsNotExist(err) {
		t.Errorf("Expected %s to be created", jsonFile)
	}
	if _, err := os.Stat(mdFile); os.IsNotExist(err) {
		t.Errorf("Expected %s to be created", mdFile)
	}

	// Cleanup
	os.Remove(jsonFile)
	os.Remove(mdFile)
}
