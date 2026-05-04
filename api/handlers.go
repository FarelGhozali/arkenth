package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/FarelGhozali/arkenth/config"
	"github.com/FarelGhozali/arkenth/db"
	"github.com/FarelGhozali/arkenth/models"
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

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
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

	cleanDir := filepath.Clean(dir)
	if filepath.IsAbs(cleanDir) || strings.Contains(cleanDir, "..") {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid directory path. Path traversal is not allowed."})
		return
	}

	files, err := os.ReadDir(cleanDir)
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

// --- Smart Visual Regression Handlers ---

// handleVisualBaselines returns a list of baseline proof directories for a project.
func handleVisualBaselines(w http.ResponseWriter, r *http.Request) {
	project := r.URL.Query().Get("project")
	if project == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "project parameter is required"})
		return
	}

	safeProject := filepath.Base(project)
	if safeProject == "." || safeProject == "/" || safeProject == "\\" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid project name"})
		return
	}

	projectDir := filepath.Join("proofs", safeProject)
	entries, err := os.ReadDir(projectDir)
	if err != nil {
		writeJSON(w, http.StatusOK, []string{}) // No baselines yet
		return
	}

	type BaselineInfo struct {
		Name   string   `json:"name"`
		Path   string   `json:"path"`
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
