package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

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
