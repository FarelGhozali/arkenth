package main

import (
	"embed"
	"io/fs"
	"log"

	"web-qa-automation/cmd"
)

//go:embed frontend/dist/*
var frontendFS embed.FS

func main() {
	// Extract the sub-filesystem and inject it into the cmd package
	stripped, err := fs.Sub(frontendFS, "frontend/dist")
	if err != nil {
		log.Fatalf("Failed to load embedded frontend: %v", err)
	}
	cmd.FrontendFS = stripped

	cmd.Execute()
}
