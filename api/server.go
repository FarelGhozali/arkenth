package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"web-qa-automation/a11y"
	"web-qa-automation/config"
	"web-qa-automation/crawler"
	"web-qa-automation/loadtester"
	"web-qa-automation/visual"
)

// RunRequest is the JSON payload the frontend sends
type RunRequest struct {
	Command         string `json:"command"`
	Target          string `json:"target"`
	Depth           int    `json:"depth"`
	FastMode        bool   `json:"fast_mode"`
	RecordVideo     bool   `json:"record_video"`
	MobileEmulation string `json:"mobile_emulation"`
	AuthJSON        string `json:"auth_json"`

	// Compare-specific
	BaselineDate string `json:"baseline_date"`

	// Load-specific
	Users    int    `json:"users"`
	Duration string `json:"duration"`
	Method   string `json:"method"`
	BodyJSON string `json:"body_json"`
}

// RunResponse is the JSON reply the frontend receives
type RunResponse struct {
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// StartServer boots the HTTP server with API routes and static file serving.
// The staticFS parameter serves the embedded Svelte frontend.
func StartServer(port int, staticFS http.Handler) {
	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("/api/run", handleRun)
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	// Serve the embedded Svelte frontend for everything else
	mux.Handle("/", staticFS)

	addr := fmt.Sprintf(":%d", port)
	log.Printf("🌐 Web QA Dashboard is live at http://localhost:%d", port)
	log.Println("   Press Ctrl+C to stop the server.")

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func handleRun(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, RunResponse{Error: "Only POST is allowed"})
		return
	}

	var req RunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, RunResponse{Error: "Invalid JSON: " + err.Error()})
		return
	}

	if req.Target == "" {
		writeJSON(w, http.StatusBadRequest, RunResponse{Error: "Target URL is required"})
		return
	}
	if req.Depth <= 0 {
		req.Depth = 1
	}

	cfg := &config.AppConfig{
		Target:          req.Target,
		Depth:           req.Depth,
		FastMode:        req.FastMode,
		RecordVideo:     req.RecordVideo,
		MobileEmulation: req.MobileEmulation,
		AuthJSON:        req.AuthJSON,
	}

	switch req.Command {
	case "scan":
		go executeScan(cfg)
		writeJSON(w, http.StatusOK, RunResponse{Message: "Scan dimulai! Bot sedang merayap ke " + req.Target})

	case "baseline":
		go executeBaseline(cfg)
		writeJSON(w, http.StatusOK, RunResponse{Message: "Baseline dimulai! Mengambil snapshot dari " + req.Target})

	case "compare":
		go executeCompare(cfg, req.BaselineDate)
		writeJSON(w, http.StatusOK, RunResponse{Message: "Compare dimulai! Membandingkan visual " + req.Target})

	case "a11y":
		go executeA11y(cfg)
		writeJSON(w, http.StatusOK, RunResponse{Message: "Audit Aksesibilitas dimulai terhadap " + req.Target})

	case "load":
		go executeLoad(cfg, req.Users, req.Duration, req.Method, req.BodyJSON)
		writeJSON(w, http.StatusOK, RunResponse{Message: fmt.Sprintf("Load Test dimulai! %d users → %s", req.Users, req.Target)})

	default:
		writeJSON(w, http.StatusBadRequest, RunResponse{Error: "Unknown command: " + req.Command})
	}
}

// --- Execution Helpers (run in goroutines) ---

func executeScan(cfg *config.AppConfig) {
	log.Printf("[UI] Starting Scan on %s ...", cfg.Target)
	spider := crawler.NewSpider(cfg)
	if err := spider.Run(); err != nil {
		log.Printf("[UI] Scan error: %v", err)
	} else {
		log.Println("[UI] ✅ Scan completed successfully.")
	}
}

func executeBaseline(cfg *config.AppConfig) {
	log.Printf("[UI] Starting Baseline on %s ...", cfg.Target)
	cfg.FastMode = true
	spider := crawler.NewSpider(cfg)
	dateStr := time.Now().Format("02-01-2006")
	spider.ProofDir = fmt.Sprintf("./proofs/%s/baseline", dateStr)
	spider.SkipFuzzing = true
	if err := spider.Run(); err != nil {
		log.Printf("[UI] Baseline error: %v", err)
	} else {
		log.Println("[UI] ✅ Baseline completed successfully.")
	}
}

func executeCompare(cfg *config.AppConfig, baselineDate string) {
	log.Printf("[UI] Starting Compare on %s ...", cfg.Target)
	cfg.FastMode = true
	spider := crawler.NewSpider(cfg)
	dateStr := time.Now().Format("02-01-2006")
	spider.ProofDir = fmt.Sprintf("./proofs/%s/current", dateStr)
	spider.SkipFuzzing = true

	if err := spider.Run(); err != nil {
		log.Printf("[UI] Compare crawl error: %v", err)
		return
	}

	bDate := baselineDate
	if bDate == "" {
		bDate = dateStr
	}
	baselineDir := fmt.Sprintf("./proofs/%s/baseline", bDate)
	diffDir := fmt.Sprintf("./proofs/%s/diff", dateStr)
	if err := visual.GenerateRegressionReport(baselineDir, spider.ProofDir, diffDir, "visual_regression_report.md"); err != nil {
		log.Printf("[UI] Visual diff error: %v", err)
	} else {
		log.Println("[UI] ✅ Visual Regression compare completed.")
	}
}

func executeA11y(cfg *config.AppConfig) {
	log.Printf("[UI] Starting A11y Audit on %s ...", cfg.Target)
	reportFile := "accessibility_audit_report.md"
	if err := a11y.InitializeReport(reportFile); err != nil {
		log.Printf("[UI] A11y init error: %v", err)
		return
	}
	cfg.FastMode = true
	spider := crawler.NewSpider(cfg)
	spider.SkipFuzzing = true
	spider.RunA11y = true
	if err := spider.Run(); err != nil {
		log.Printf("[UI] A11y error: %v", err)
	} else {
		log.Println("[UI] ✅ Accessibility Audit completed.")
	}
}

func executeLoad(cfg *config.AppConfig, users int, duration string, method string, bodyJSON string) {
	log.Printf("[UI] Starting Load Test on %s with %d users...", cfg.Target, users)
	if users <= 0 {
		users = 50
	}
	parsedDuration, err := time.ParseDuration(duration)
	if err != nil {
		log.Printf("[UI] Invalid duration: %v", err)
		return
	}
	res := loadtester.RunLoadTest(cfg.Target, users, parsedDuration, method, bodyJSON)

	if err := loadtester.GenerateLoadReport("load_test_report.md", cfg.Target, users, res); err != nil {
		log.Printf("[UI] Load report error: %v", err)
	} else {
		log.Println("[UI] ✅ Load Test completed.")
	}
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}
