package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/FarelGhozali/web-qa-automation/a11y"
	"github.com/FarelGhozali/web-qa-automation/config"
	"github.com/FarelGhozali/web-qa-automation/crawler"
	"github.com/FarelGhozali/web-qa-automation/db"
	"github.com/FarelGhozali/web-qa-automation/loadtester"
	"github.com/FarelGhozali/web-qa-automation/visual"
)

// RunRequest is the JSON payload the frontend sends
type RunRequest struct {
	Command         string `json:"command"`
	ProjectName     string `json:"project_name"`
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

// SetupRouter creates and configures the HTTP router with API endpoints.
// If staticFS is provided, it also serves the embedded frontend.
func SetupRouter(staticFS http.Handler) *http.ServeMux {
	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("/api/run", handleRun)
	mux.HandleFunc("/api/projects", handleProjects)
	mux.HandleFunc("/api/history", handleHistory)
	mux.HandleFunc("/api/gallery", handleGallery)
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	// Serve proofs folder for history gallery
	mux.Handle("/proofs/", http.StripPrefix("/proofs/", http.FileServer(http.Dir("./proofs"))))

	// Serve the embedded Svelte frontend for everything else
	if staticFS != nil {
		mux.Handle("/", staticFS)
	}

	return mux
}

// StartServer boots the HTTP server with API routes and static file serving.
// The staticFS parameter serves the embedded Svelte frontend.
func StartServer(port int, staticFS http.Handler) {
	mux := SetupRouter(staticFS)

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

	if req.ProjectName == "" {
		req.ProjectName = "default"
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
		go executeScan(cfg, req.ProjectName)
		writeJSON(w, http.StatusOK, RunResponse{Message: "Scan dimulai! Bot sedang merayap ke " + req.Target})

	case "baseline":
		go executeBaseline(cfg, req.ProjectName)
		writeJSON(w, http.StatusOK, RunResponse{Message: "Baseline dimulai! Mengambil snapshot dari " + req.Target})

	case "compare":
		go executeCompare(cfg, req.ProjectName, req.BaselineDate)
		writeJSON(w, http.StatusOK, RunResponse{Message: "Compare dimulai! Membandingkan visual " + req.Target})

	case "a11y":
		go executeA11y(cfg, req.ProjectName)
		writeJSON(w, http.StatusOK, RunResponse{Message: "Audit Aksesibilitas dimulai terhadap " + req.Target})

	case "load":
		go executeLoad(cfg, req.ProjectName, req.Users, req.Duration, req.Method, req.BodyJSON)
		writeJSON(w, http.StatusOK, RunResponse{Message: fmt.Sprintf("Load Test dimulai! %d users → %s", req.Users, req.Target)})

	default:
		writeJSON(w, http.StatusBadRequest, RunResponse{Error: "Unknown command: " + req.Command})
	}
}

func handleProjects(w http.ResponseWriter, r *http.Request) {
	projects, err := db.GetProjects()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, projects)
}

func handleHistory(w http.ResponseWriter, r *http.Request) {
	projectName := r.URL.Query().Get("project")
	if projectName == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "project parameter is required"})
		return
	}

	history, err := db.GetHistoryByProject(projectName)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	if history == nil {
		history = []db.TestRun{}
	}

	writeJSON(w, http.StatusOK, history)
}

func handleGallery(w http.ResponseWriter, r *http.Request) {
	dir := r.URL.Query().Get("dir")
	if dir == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "dir parameter is required"})
		return
	}

	files, err := os.ReadDir(dir)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	gallery := struct {
		Images  []string `json:"images"`
		Reports []string `json:"reports"`
	}{
		Images:  []string{},
		Reports: []string{},
	}

	for _, f := range files {
		if f.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(f.Name()))
		if ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".gif" {
			gallery.Images = append(gallery.Images, f.Name())
		} else if ext == ".md" || ext == ".html" || ext == ".json" {
			gallery.Reports = append(gallery.Reports, f.Name())
		}
	}
	writeJSON(w, http.StatusOK, gallery)
}

// --- Execution Helpers (run in goroutines) ---

func executeScan(cfg *config.AppConfig, projectName string) {
	timestamp := time.Now().Format("20060102_150405")
	proofDir := fmt.Sprintf("./proofs/%s/%s_scan", projectName, timestamp)
	
	runID, _ := db.CreateRun(projectName, "scan", cfg.Target, proofDir)
	log.Printf("[UI] Starting Scan on %s (Project: %s) ...", cfg.Target, projectName)
	
	spider := crawler.NewSpider(cfg)
	spider.ProofDir = proofDir

	if err := spider.Run(); err != nil {
		log.Printf("[UI] Scan error: %v", err)
		db.UpdateRunStatus(runID, "failed")
	} else {
		log.Println("✅ Scan completed successfully.")
		db.UpdateRunStatus(runID, "completed")
	}
}

func executeBaseline(cfg *config.AppConfig, projectName string) {
	timestamp := time.Now().Format("20060102_150405")
	proofDir := fmt.Sprintf("./proofs/%s/%s_baseline", projectName, timestamp)
	
	runID, _ := db.CreateRun(projectName, "baseline", cfg.Target, proofDir)
	log.Printf("[UI] Starting Baseline on %s (Project: %s) ...", cfg.Target, projectName)
	
	cfg.FastMode = true
	spider := crawler.NewSpider(cfg)
	spider.ProofDir = proofDir
	spider.SkipFuzzing = true
	if err := spider.Run(); err != nil {
		log.Printf("[UI] Baseline error: %v", err)
		db.UpdateRunStatus(runID, "failed")
	} else {
		log.Println("✅ Baseline completed successfully.")
		db.UpdateRunStatus(runID, "completed")
	}
}

func executeCompare(cfg *config.AppConfig, projectName string, baselinePath string) {
	timestamp := time.Now().Format("20060102_150405")
	proofDir := fmt.Sprintf("./proofs/%s/%s_compare", projectName, timestamp)
	
	runID, _ := db.CreateRun(projectName, "compare", cfg.Target, proofDir)
	log.Printf("[UI] Starting Compare on %s (Project: %s) ...", cfg.Target, projectName)
	
	cfg.FastMode = true
	spider := crawler.NewSpider(cfg)
	spider.ProofDir = proofDir
	spider.SkipFuzzing = true

	if err := spider.Run(); err != nil {
		log.Printf("[UI] Compare crawl error: %v", err)
		db.UpdateRunStatus(runID, "failed")
		return
	}

	baselineDir := baselinePath
	if !strings.Contains(baselineDir, "/") && !strings.Contains(baselineDir, "\\") {
		baselineDir = filepath.Join("proofs", projectName, baselinePath)
	}

	diffDir := filepath.Join(spider.ProofDir, "diff")
	if err := visual.GenerateRegressionReport(baselineDir, spider.ProofDir, diffDir, "visual_regression_report.md"); err != nil {
		log.Printf("[UI] Visual diff error: %v", err)
		db.UpdateRunStatus(runID, "failed")
	} else {
		log.Println("✅ Visual Regression compare completed.")
		db.UpdateRunStatus(runID, "completed")
	}
}

func executeA11y(cfg *config.AppConfig, projectName string) {
	timestamp := time.Now().Format("20060102_150405")
	proofDir := fmt.Sprintf("./proofs/%s/%s_a11y", projectName, timestamp)
	
	runID, _ := db.CreateRun(projectName, "a11y", cfg.Target, proofDir)
	log.Printf("[UI] Starting A11y Audit on %s (Project: %s) ...", cfg.Target, projectName)
	
	os.MkdirAll(proofDir, 0755)
	reportFile := filepath.Join(proofDir, "accessibility_audit_report.md")
	if err := a11y.InitializeReport(reportFile); err != nil {
		log.Printf("[UI] A11y init error: %v", err)
		db.UpdateRunStatus(runID, "failed")
		return
	}
	
	cfg.FastMode = true
	spider := crawler.NewSpider(cfg)
	spider.ProofDir = proofDir
	spider.SkipFuzzing = true
	spider.RunA11y = true
	if err := spider.Run(); err != nil {
		log.Printf("[UI] A11y error: %v", err)
		db.UpdateRunStatus(runID, "failed")
	} else {
		log.Println("✅ Accessibility Audit completed.")
		db.UpdateRunStatus(runID, "completed")
	}
}

func executeLoad(cfg *config.AppConfig, projectName string, users int, duration string, method string, bodyJSON string) {
	timestamp := time.Now().Format("20060102_150405")
	proofDir := fmt.Sprintf("./proofs/%s/%s_load", projectName, timestamp)
	
	runID, _ := db.CreateRun(projectName, "load", cfg.Target, proofDir)
	log.Printf("[UI] Starting Load Test on %s (Project: %s) with %d users...", cfg.Target, projectName, users)
	
	if users <= 0 {
		users = 50
	}
	parsedDuration, err := time.ParseDuration(duration)
	if err != nil {
		log.Printf("[UI] Invalid duration: %v", err)
		db.UpdateRunStatus(runID, "failed")
		return
	}

	os.MkdirAll(proofDir, 0755)
	res := loadtester.RunLoadTest(cfg.Target, users, parsedDuration, method, bodyJSON)

	reportPath := filepath.Join(proofDir, "load_test_report.md")
	if err := loadtester.GenerateLoadReport(reportPath, cfg.Target, users, res); err != nil {
		log.Printf("[UI] Load report error: %v", err)
		db.UpdateRunStatus(runID, "failed")
	} else {
		log.Println("✅ Load Test completed.")
		db.UpdateRunStatus(runID, "completed")
	}
}


func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}
