package main

import (
	"io"
	"log"
	"os"

	"github.com/FarelGhozali/web-qa-automation/api"
	"github.com/FarelGhozali/web-qa-automation/frontend"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

func main() {
	// Create an instance of the app structure
	app := NewApp()

	// Intercept all standard log output so it gets pushed to the Svelte UI
	// Fallback to stdout as well so terminal output is preserved during debug
	multi := io.MultiWriter(app, os.Stdout)
	log.SetOutput(multi)

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "Web QA Automation",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: frontend.DistFS,
			// The AssetsHandler intercepts requests (like our /api/run from Svelte fetch calls)
			// that don't match static files, and routes them through our existing backend router.
			Handler: api.SetupRouter(nil),
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
		// Adjust OS specific options for a premium feel
		Windows: &windows.Options{
			WebviewIsTransparent:              true,
			WindowIsTranslucent:               true,
			BackdropType:                      windows.Mica,
			DisableWindowIcon:                 false,
			DisableFramelessWindowDecorations: false,
			Theme:                             windows.Dark,
		},
		Mac: &mac.Options{
			TitleBar:             mac.TitleBarHiddenInset(),
			Appearance:           mac.NSAppearanceNameDarkAqua,
			WebviewIsTransparent: true,
			WindowIsTranslucent:  true,
		},
		Linux: &linux.Options{
			WindowIsTranslucent: true,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
