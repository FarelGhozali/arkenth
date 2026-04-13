package cmd

import (
	"fmt"
	"io/fs"
	"log"
	"net/http"

	"web-qa-automation/api"

	"github.com/spf13/cobra"
)

var uiPort int

// FrontendFS is set by main.go to inject the embedded filesystem
var FrontendFS fs.FS

var uiCmd = &cobra.Command{
	Use:   "ui",
	Short: "Launch the Web QA Dashboard in your browser",
	Long: `Start a local web server that serves the Web QA Automation Dashboard.
All CLI features (scan, baseline, compare, a11y, load) are accessible
through a premium, modern web interface — no command-line typing needed.`,
	Run: func(cmd *cobra.Command, args []string) {
		if FrontendFS == nil {
			log.Fatal("Frontend assets not embedded. Build with 'go build' first.")
		}

		fmt.Println()
		fmt.Println("╔══════════════════════════════════════════════╗")
		fmt.Println("║     🕵️  Web QA Automation Dashboard          ║")
		fmt.Println("╠══════════════════════════════════════════════╣")
		fmt.Printf("║  🌐 Open: http://localhost:%-5d             ║\n", uiPort)
		fmt.Println("║  ⏹  Stop: Press Ctrl+C                      ║")
		fmt.Println("╚══════════════════════════════════════════════╝")
		fmt.Println()

		handler := createFrontendHandler(FrontendFS)
		api.StartServer(uiPort, handler)
	},
}

func init() {
	uiCmd.Flags().IntVar(&uiPort, "port", 8080, "Port to serve the web dashboard on")
	rootCmd.AddCommand(uiCmd)
}

// createFrontendHandler builds an http.Handler that serves static files from
// the embedded FS and falls back to index.html for SPA client-side routing.
func createFrontendHandler(frontendRoot fs.FS) http.Handler {
	fileServer := http.FileServer(http.FS(frontendRoot))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path == "/" {
			path = "/index.html"
		}

		// Check if the file exists in the embedded FS
		f, err := frontendRoot.Open(path[1:]) // strip leading "/"
		if err != nil {
			// File not found — serve index.html for SPA routing
			r.URL.Path = "/"
			fileServer.ServeHTTP(w, r)
			return
		}
		f.Close()

		fileServer.ServeHTTP(w, r)
	})
}
