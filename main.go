package main

import (
	"embed"
	"io/fs"
	"log"

	"web-qa-automation/cmd"
	"web-qa-automation/db"
)

//go:embed frontend/dist/*
var frontendFS embed.FS

func main() {
	// Initialize SQLite Database
	db.Init()

	// Extract the sub-filesystem and inject it into the cmd package
	stripped, err := fs.Sub(frontendFS, "frontend/dist")
	if err != nil {
		log.Fatalf("Failed to load embedded frontend: %v", err)
	}
	cmd.FrontendFS = stripped

	cmd.Execute()
}
