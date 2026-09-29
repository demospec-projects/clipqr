package main

import (
	"embed"
	"flag"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed icons/tray.png
var trayIcon []byte

//go:embed icons/tray-template.png
var trayTemplateIcon []byte

func main() {
	configDir, err := os.UserConfigDir()
	if err != nil {
		log.Fatal(err)
	}
	dataDir := flag.String("data-dir", filepath.Join(configDir, "ClipQR"), "Dossier de l’historique")
	paused := flag.Bool("paused", false, "Démarrer avec la collecte suspendue")
	flag.Parse()

	clip := newClipService(filepath.Join(*dataDir, "history.json"), *paused)
	var show func()

	app := application.New(application.Options{
		Name:        "ClipQR",
		Description: "Historique du presse-papiers et QR codes",
		Services:    []application.Service{application.NewService(clip)},
		Assets:      application.AssetOptions{Handler: application.AssetFileServerFS(assets)},
		// A second launch opens the panel of the running ClipQR instead.
		SingleInstance: &application.SingleInstanceOptions{
			UniqueID: "ca.demospec.clipqr",
			OnSecondInstanceLaunch: func(application.SecondInstanceData) {
				if show != nil {
					show()
				}
			},
		},
		Mac:     application.MacOptions{ActivationPolicy: application.ActivationPolicyAccessory},
		Windows: application.WindowsOptions{DisableQuitOnLastWindowClosed: true},
		Linux:   application.LinuxOptions{DisableQuitOnLastWindowClosed: true, ProgramName: "clipqr"},
	})
	clip.emit = func(st State) { app.Event.Emit(stateEvent, st) }

	// The panel floats above the other windows and stays open while the user
	// pastes elsewhere: only the reduce button or the tray icon puts it away.
	panel := app.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:             "panneau",
		Title:            "ClipQR",
		Width:            panelWidth,
		Height:           panelHeight,
		Frameless:        true,
		AlwaysOnTop:      true,
		Hidden:           true,
		DisableResize:    true,
		BackgroundColour: application.NewRGB(15, 23, 42),
		URL:              "/",
		Windows:          application.WindowsWindow{HiddenOnTaskbar: true},
	})
	panel.RegisterHook(events.Common.WindowClosing, func(e *application.WindowEvent) {
		panel.Hide()
		e.Cancel()
	})

	// Every appearance puts the panel back above the clock.
	show = func() {
		placeNearClock(app, panel)
		panel.Show()
		panel.Focus()
		if runtime.GOOS == "linux" {
			go placeAgainAfterWindowManager(app, panel)
		}
	}
	app.Event.OnApplicationEvent(events.Common.ApplicationStarted, func(*application.ApplicationEvent) { show() })

	tray := app.SystemTray.New()
	tray.SetIcon(trayIcon)
	if runtime.GOOS == "darwin" {
		tray.SetTemplateIcon(trayTemplateIcon)
	}
	tray.SetTooltip("ClipQR — historique du presse-papiers")
	tray.SetMenu(trayMenu(app, show, clip, *paused))
	tray.OnClick(func() {
		if panel.IsVisible() {
			panel.Hide()
		} else {
			show()
		}
	})

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}

func trayMenu(app *application.App, show func(), clip *ClipService, paused bool) *application.Menu {
	menu := app.NewMenu()
	menu.Add("Ouvrir ClipQR").OnClick(func(*application.Context) { show() })
	pause := menu.AddCheckbox("Suspendre la collecte", paused)
	pause.OnClick(func(ctx *application.Context) { clip.SetPaused(ctx.ClickedMenuItem().Checked()) })
	clip.pausedChanged = func(p bool) { pause.SetChecked(p) }
	menu.AddSeparator()
	menu.Add("Quitter").OnClick(func(*application.Context) { app.Quit() })
	return menu
}
