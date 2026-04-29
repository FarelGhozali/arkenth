package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/FarelGhozali/arkenth/a11y"
	"github.com/FarelGhozali/arkenth/config"
	"github.com/FarelGhozali/arkenth/crawler"
	"github.com/FarelGhozali/arkenth/db"
	"github.com/FarelGhozali/arkenth/loadtester"
	"github.com/FarelGhozali/arkenth/models"
	"github.com/FarelGhozali/arkenth/swagger"
	"github.com/FarelGhozali/arkenth/visual"
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
	BaselineDate string  `json:"baseline_date"`
	Threshold    float64 `json:"threshold"` // Euclidean color distance tolerance (0 = exact match)

	// Load-specific
	Users    int    `json:"users"`
	Duration string `json:"duration"`
	Method   string `json:"method"`
	BodyJSON string `json:"body_json"`

	// Fuzz-API-specific
	SwaggerURL  string `json:"swagger_url"`
	Concurrency int    `json:"concurrency"`
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

	// Smart Visual Regression endpoints
	mux.HandleFunc("/api/visual/baselines", handleVisualBaselines)
	mux.HandleFunc("/api/visual/masks", handleVisualMasks)
	mux.HandleFunc("/api/visual/masks/delete", handleDeleteVisualMask)

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
		go executeCompare(cfg, req.ProjectName, req.BaselineDate, req.Threshold)
		writeJSON(w, http.StatusOK, RunResponse{Message: "Compare dimulai! Membandingkan visual " + req.Target})

	case "a11y":
		go executeA11y(cfg, req.ProjectName)
		writeJSON(w, http.StatusOK, RunResponse{Message: "Audit Aksesibilitas dimulai terhadap " + req.Target})

	case "load":
		go executeLoad(cfg, req.ProjectName, req.Users, req.Duration, req.Method, req.BodyJSON)
		writeJSON(w, http.StatusOK, RunResponse{Message: fmt.Sprintf("Load Test dimulai! %d users → %s", req.Users, req.Target)})

	case "fuzz-api":
		if req.SwaggerURL == "" {
			writeJSON(w, http.StatusBadRequest, RunResponse{Error: "Swagger/OpenAPI URL is required"})
			return
		}
		go executeFuzzAPI(req.ProjectName, req.Target, req.SwaggerURL, req.Concurrency)
		writeJSON(w, http.StatusOK, RunResponse{Message: fmt.Sprintf("API Fuzzing started! Spec: %s → %s", req.SwaggerURL, req.Target)})

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

func executeCompare(cfg *config.AppConfig, projectName string, baselinePath string, threshold float64) {
	timestamp := time.Now().Format("20060102_150405")
	proofDir := fmt.Sprintf("./proofs/%s/%s_compare", projectName, timestamp)

	runID, _ := db.CreateRun(projectName, "compare", cfg.Target, proofDir)
	log.Printf("[UI] Starting Smart Compare on %s (Project: %s, Threshold: %.1f) ...", cfg.Target, projectName, threshold)

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

	// Set threshold default if not specified
	if threshold <= 0 {
		threshold = visual.DefaultThreshold
	}

	// Fetch masks from database for the target URL
	masks, err := db.GetVisualMasks(cfg.Target)
	if err != nil {
		log.Printf("[UI] Warning: Could not load visual masks: %v", err)
		masks = nil
	}

	diffDir := filepath.Join(spider.ProofDir, "diff")
	reportPath := filepath.Join(spider.ProofDir, "visual_regression_report.md")
	if err := visual.GenerateSmartRegressionReport(baselineDir, spider.ProofDir, diffDir, reportPath, masks, threshold); err != nil {
		log.Printf("[UI] Visual diff error: %v", err)
		db.UpdateRunStatus(runID, "failed")
	} else {
		log.Printf("✅ Smart Visual Regression complete. Threshold=%.1f, Masks=%d", threshold, len(masks))
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

func executeFuzzAPI(projectName string, targetBaseURL string, swaggerURL string, concurrencyLevel int) {
	timestamp := time.Now().Format("20060102_150405")
	proofDir := fmt.Sprintf("./proofs/%s/%s_fuzz_api", projectName, timestamp)

	runID, _ := db.CreateRun(projectName, "fuzz-api", targetBaseURL, proofDir)
	log.Printf("[UI] Starting API Fuzzing on %s using spec %s (Project: %s) ...", targetBaseURL, swaggerURL, projectName)

	parser := swagger.NewParser()
	if err := parser.LoadSpec(swaggerURL); err != nil {
		log.Printf("[UI] Failed to load Swagger spec: %v", err)
		db.UpdateRunStatus(runID, "failed")
		return
	}

	endpoints, err := parser.ExtractEndpoints()
	if err != nil {
		log.Printf("[UI] Failed to extract endpoints: %v", err)
		db.UpdateRunStatus(runID, "failed")
		return
	}

	if concurrencyLevel <= 0 {
		concurrencyLevel = 10
	}

	fuzzer := swagger.NewFuzzer(targetBaseURL, endpoints, concurrencyLevel)
	fuzzer.Report.SpecURL = swaggerURL
	report := fuzzer.Run()

	if err := swagger.GenerateReport(report, proofDir); err != nil {
		log.Printf("[UI] API Fuzz report error: %v", err)
		db.UpdateRunStatus(runID, "failed")
	} else {
		log.Printf("✅ API Fuzzing completed. %d anomalies found.", report.TotalAnomalies)
		db.UpdateRunStatus(runID, "completed")
	}
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

// --- Smart Visual Regression Handlers ---

// handleVisualBaselines returns a list of baseline proof directories for a project.
func handleVisualBaselines(w http.ResponseWriter, r *http.Request) {
	project := r.URL.Query().Get("project")
	if project == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "project parameter is required"})
		return
	}

	projectDir := filepath.Join("proofs", project)
	entries, err := os.ReadDir(projectDir)
	if err != nil {
		writeJSON(w, http.StatusOK, []string{}) // No baselines yet
		return
	}

	type BaselineInfo struct {
		Name string   `json:"name"`
		Path string   `json:"path"`
		Images []string `json:"images"`
	}

	var baselines []BaselineInfo
	for _, e := range entries {
		if !e.IsDir() || !strings.Contains(e.Name(), "baseline") {
			continue
		}
		// Enumerate images inside the baseline dir
		imgDir := filepath.Join(projectDir, e.Name())
		imgEntries, err := os.ReadDir(imgDir)
		var images []string
		if err == nil {
			for _, img := range imgEntries {
				if !img.IsDir() && strings.HasSuffix(strings.ToLower(img.Name()), ".png") {
					images = append(images, img.Name())
				}
			}
		}
		baselines = append(baselines, BaselineInfo{
			Name:   e.Name(),
			Path:   imgDir,
			Images: images,
		})
	}

	if baselines == nil {
		baselines = []BaselineInfo{}
	}
	writeJSON(w, http.StatusOK, baselines)
}

// handleVisualMasks handles GET (list) and POST (create) for visual masks.
func handleVisualMasks(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		targetURL := r.URL.Query().Get("target_url")
		if targetURL == "" {
			// Return all masks if no target_url specified
			masks, err := db.GetAllVisualMasks()
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
				return
			}
			if masks == nil {
				masks = []models.VisualMask{}
			}
			writeJSON(w, http.StatusOK, masks)
			return
		}

		masks, err := db.GetVisualMasks(targetURL)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		if masks == nil {
			masks = []models.VisualMask{}
		}
		writeJSON(w, http.StatusOK, masks)

	case http.MethodPost:
		var mask models.VisualMask
		if err := json.NewDecoder(r.Body).Decode(&mask); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON: " + err.Error()})
			return
		}
		if mask.TargetURL == "" || mask.Width <= 0 || mask.Height <= 0 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "target_url, width, and height are required"})
			return
		}
		id, err := db.CreateVisualMask(mask)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		mask.ID = int(id)
		writeJSON(w, http.StatusCreated, mask)

	default:
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Only GET and POST are allowed"})
	}
}

// handleDeleteVisualMask handles DELETE requests to remove a mask by ID.
func handleDeleteVisualMask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodDelete {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Only POST/DELETE are allowed"})
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id parameter is required"})
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}

	if err := db.DeleteVisualMask(id); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "Mask deleted successfully"})
}
