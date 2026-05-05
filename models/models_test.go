package models

import (
	"testing"
)

func TestNewReport(t *testing.T) {
	target := "https://example.com"
	mobile := true

	report := NewReport(target, mobile)

	if report.TargetURL != target {
		t.Errorf("Expected TargetURL to be %s, got %s", target, report.TargetURL)
	}

	if report.MobileEmulation != mobile {
		t.Errorf("Expected MobileEmulation to be %v, got %v", mobile, report.MobileEmulation)
	}

	if report.ScannedURLs == nil {
		t.Error("Expected ScannedURLs to be initialized, got nil")
	}

	if report.NetworkAnomalies == nil {
		t.Error("Expected NetworkAnomalies to be initialized, got nil")
	}

	if report.CriticalBugs == nil {
		t.Error("Expected CriticalBugs to be initialized, got nil")
	}

	if report.UIUXIssues == nil {
		t.Error("Expected UIUXIssues to be initialized, got nil")
	}
}
