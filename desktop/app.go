package main

import (
	"context"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// Write intercepts the log output and sends it to the Wails frontend.
// This allows crawler and API logs to be displayed in the Svelte UI.
func (a *App) Write(p []byte) (n int, err error) {
	if a.ctx != nil {
		// Emit the log message to the frontend event listener
		runtime.EventsEmit(a.ctx, "backend-log", string(p))
	}
	return len(p), nil
}
