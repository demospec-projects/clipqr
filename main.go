package main

import (
	"embed"
	"flag"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"github.com/adrg/xdg"
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
	capturesDir := flag.String("captures-dir", filepath.Join(xdg.UserDirs.Pictures, "ecrans"), "Dossier des captures Impr. écran")
	flag.Parse()

	clip := newClipService(*dataDir, *capturesDir, *paused)
	var show func()

	app := application.New(application.Options{
		Name:        "ClipQR",
		Description: "Historique du presse-papiers et QR codes",
		Services:    []application.Service{application.NewService(clip)},
		Assets: application.AssetOptions{
			Handler:    application.AssetFileServerFS(assets),
			Middleware: clip.serveThumbnails,
		},
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
			go afterWindowManager(func() { placeNearClock(app, panel) })
		}
	}

	// A screenshot leaves the panel out of the picture, then puts it back
	// where it was, even if the user had moved it.
	clip.withPanelHidden = func(grab func()) {
		if !panel.IsVisible() {
			grab()
			return
		}
		x, y := panel.Position()
		panel.Hide()
		time.Sleep(250 * time.Millisecond) // the screen repaints what the panel covered
		grab()
		panel.SetPosition(x, y)
		panel.Show()
		if runtime.GOOS == "linux" {
			go afterWindowManager(func() { panel.SetPosition(x, y) })
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

	// Closing the session or a kill quits cleanly: the desktop gets Print back.
	// A signal the launch chose to ignore (a background start) stays ignored.
	signals := make(chan os.Signal, 1)
	for _, sig := range []os.Signal{syscall.SIGINT, syscall.SIGTERM} {
		if !signal.Ignored(sig) {
			signal.Notify(signals, sig)
		}
	}
	go func() {
		<-signals
		app.Quit()
	}()

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
	if printKeyAvailable() {
		capture := menu.AddCheckbox("Impr. écran par ClipQR", false)
		capture.OnClick(func(ctx *application.Context) {
			if err := clip.SetCaptureMode(ctx.ClickedMenuItem().Checked()); err != nil {
				clip.fail(err.Error())
			}
		})
		clip.captureModeChanged = func(on bool) { capture.SetChecked(on) }
	}
	menu.AddSeparator()
	menu.Add("Quitter").OnClick(func(*application.Context) { app.Quit() })
	return menu
}
