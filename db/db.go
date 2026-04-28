package db

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"

	"github.com/FarelGhozali/web-qa-automation/models"
	_ "modernc.org/sqlite"
)

var DB *sql.DB

type TestRun struct {
	ID          int64  `json:"id"`
	ProjectName string `json:"project_name"`
	TestType    string `json:"test_type"`
	TargetURL   string `json:"target_url"`
	Status      string `json:"status"`
	Timestamp   string `json:"timestamp"`
	ReportDir   string `json:"report_dir"`
}

// Init initializes the SQLite database and creates the runs table if it doesn't exist.
func Init() {
	dbPath := "./proofs/qa_automation.db"

	// Ensure proofs dir exists
	dir := filepath.Dir(dbPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, 0755)
	}

	var err error
	DB, err = sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("❌ Failed to open database: %v", err)
	}

	// Enable WAL mode for better concurrency and performance
	_, err = DB.Exec("PRAGMA journal_mode=WAL;")
	if err != nil {
		log.Printf("⚠️ Warning: Failed to set WAL mode: %v", err)
	}

	// Set busy timeout to 5 seconds to avoid "database is locked" errors
	_, err = DB.Exec("PRAGMA busy_timeout=5000;")
	if err != nil {
		log.Printf("⚠️ Warning: Failed to set busy_timeout: %v", err)
	}

	createTableQuery := `
	CREATE TABLE IF NOT EXISTS runs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		project_name TEXT NOT NULL,
		test_type TEXT NOT NULL,
		target_url TEXT NOT NULL,
		status TEXT NOT NULL,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		report_dir TEXT NOT NULL
	);`

	_, err = DB.Exec(createTableQuery)
	if err != nil {
		log.Fatalf("❌ Failed to create table: %v", err)
	}

	// Create visual_masks table for Smart Visual Regression ignore-regions
	createMasksQuery := `
	CREATE TABLE IF NOT EXISTS visual_masks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		target_url TEXT NOT NULL,
		x INTEGER NOT NULL,
		y INTEGER NOT NULL,
		width INTEGER NOT NULL,
		height INTEGER NOT NULL,
		label TEXT NOT NULL DEFAULT ''
	);`

	_, err = DB.Exec(createMasksQuery)
	if err != nil {
		log.Fatalf("❌ Failed to create visual_masks table: %v", err)
	}

	log.Println("🗄️ Database initialized successfully.")
}

// CreateRun inserts a new test run record and returns its ID.
func CreateRun(project, testType, target, reportDir string) (int64, error) {
	query := `INSERT INTO runs (project_name, test_type, target_url, status, report_dir) 
	          VALUES (?, ?, ?, 'running', ?)`
	res, err := DB.Exec(query, project, testType, target, reportDir)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// UpdateRunStatus updates the status of an existing test run.
func UpdateRunStatus(id int64, status string) error {
	_, err := DB.Exec("UPDATE runs SET status = ? WHERE id = ?", status, id)
	return err
}

// GetProjects returns a list of unique project names from the database.
func GetProjects() ([]string, error) {
	rows, err := DB.Query("SELECT DISTINCT project_name FROM runs ORDER BY project_name ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err == nil {
			projects = append(projects, name)
		}
	}
	return projects, nil
}

// GetHistoryByProject retrieves all runs for a specific project.
func GetHistoryByProject(project string) ([]TestRun, error) {
	query := `SELECT id, project_name, test_type, target_url, status, timestamp, report_dir 
	          FROM runs WHERE project_name = ? ORDER BY id DESC`
	rows, err := DB.Query(query, project)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []TestRun
	for rows.Next() {
		var r TestRun
		err := rows.Scan(&r.ID, &r.ProjectName, &r.TestType, &r.TargetURL, &r.Status, &r.Timestamp, &r.ReportDir)
		if err == nil {
			history = append(history, r)
		}
	}
	return history, nil
}

// CreateVisualMask inserts a new ignore-region mask for smart visual regression.
func CreateVisualMask(mask models.VisualMask) (int64, error) {
	query := `INSERT INTO visual_masks (target_url, x, y, width, height, label)
	          VALUES (?, ?, ?, ?, ?, ?)`
	res, err := DB.Exec(query, mask.TargetURL, mask.X, mask.Y, mask.Width, mask.Height, mask.Label)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// GetVisualMasks retrieves all masks for a given target URL.
func GetVisualMasks(targetURL string) ([]models.VisualMask, error) {
	query := `SELECT id, target_url, x, y, width, height, label
	          FROM visual_masks WHERE target_url = ? ORDER BY id ASC`
	rows, err := DB.Query(query, targetURL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var masks []models.VisualMask
	for rows.Next() {
		var m models.VisualMask
		err := rows.Scan(&m.ID, &m.TargetURL, &m.X, &m.Y, &m.Width, &m.Height, &m.Label)
		if err == nil {
			masks = append(masks, m)
		}
	}
	return masks, nil
}

// GetAllVisualMasks retrieves every mask in the database.
func GetAllVisualMasks() ([]models.VisualMask, error) {
	query := `SELECT id, target_url, x, y, width, height, label
	          FROM visual_masks ORDER BY target_url, id ASC`
	rows, err := DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var masks []models.VisualMask
	for rows.Next() {
		var m models.VisualMask
		err := rows.Scan(&m.ID, &m.TargetURL, &m.X, &m.Y, &m.Width, &m.Height, &m.Label)
		if err == nil {
			masks = append(masks, m)
		}
	}
	return masks, nil
}

// DeleteVisualMask removes a single mask by its ID.
func DeleteVisualMask(id int) error {
	_, err := DB.Exec("DELETE FROM visual_masks WHERE id = ?", id)
	return err
}
