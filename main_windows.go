package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/skip2/go-qrcode"
	"golang.org/x/sys/windows"
)

var user32 = syscall.NewLazyDLL("user32.dll")
var sequence = user32.NewProc("GetClipboardSequenceNumber")

type app struct {
	window       *walk.MainWindow
	search       *walk.LineEdit
	pauseControl *walk.CheckBox
	rows         *walk.Composite
	status       *walk.Label
	data         history
	path         string
	lastSequence uintptr
	paused       bool
}

func main() {
	runtime.LockOSThread()
	name, _ := syscall.UTF16PtrFromString("Local\\ClipQR.Demospec")
	mutex, err := windows.CreateMutex(nil, false, name)
	if err == syscall.ERROR_ALREADY_EXISTS {
		walk.MsgBox(nil, "ClipQR", "ClipQR est déjà ouvert. Cliquez sur son icône près de l’horloge.", walk.MsgBoxIconInformation)
		return
	}
	if err != nil {
		walk.MsgBox(nil, "ClipQR", err.Error(), walk.MsgBoxIconError)
		return
	}
	defer windows.CloseHandle(mutex)
	dir, err := os.UserConfigDir()
	if err != nil {
		return
	}
	dataDir := flag.String("data-dir", filepath.Join(dir, "ClipQR"), "Dossier de l’historique")
	paused := flag.Bool("paused", false, "Démarrer avec la collecte suspendue")
	flag.Parse()
	a := &app{path: filepath.Join(*dataDir, "history.json"), paused: *paused}
	var loadErr error
	a.data, loadErr = loadHistory(a.path)
	err = MainWindow{
		AssignTo: &a.window, Title: "ClipQR — Historique du presse-papiers", Font: Font{Family: "Segoe UI", PointSize: 10}, Size: Size{Width: 760, Height: 600}, MinSize: Size{Width: 560, Height: 360}, Layout: VBox{},
		Children: []Widget{
			Label{Text: "CLIPQR  /  Un clic pour copier. Un scan pour emporter.", Font: Font{PointSize: 12, Bold: true}},
			LineEdit{AssignTo: &a.search, CueBanner: "Rechercher dans les textes copiés…", OnTextChanged: func() { a.render() }},
			Composite{Layout: HBox{}, Children: []Widget{
				CheckBox{AssignTo: &a.pauseControl, Text: "Suspendre la collecte", Checked: *paused, OnCheckedChanged: func() { a.paused = a.pauseControl.Checked(); a.lastSequence, _, _ = sequence.Call(); a.render() }},
				HSpacer{},
				PushButton{Text: "Effacer l’historique", OnClicked: func() {
					if walk.MsgBox(a.window, "Effacer l’historique", "Supprimer tous les textes enregistrés ?", walk.MsgBoxYesNo|walk.MsgBoxIconQuestion) == walk.DlgCmdYes {
						a.data.Items = nil
						a.persist()
						a.render()
					}
				}},
			}},
			ScrollView{Layout: VBox{}, HorizontalFixed: true, Children: []Widget{Composite{AssignTo: &a.rows, Layout: VBox{MarginsZero: true, Spacing: 6}}}},
			Label{AssignTo: &a.status},
			Label{Text: "Textes conservés localement • 100 entrées maximum • Fermer masque la fenêtre."},
		},
	}.Create()
	if err != nil {
		walk.MsgBox(nil, "ClipQR", err.Error(), walk.MsgBoxIconError)
		return
	}
	defer a.window.Dispose()
	a.pauseControl.SetChecked(*paused)
	a.paused = *paused
	icon, err := makeIcon()
	if err != nil {
		a.fail(err)
		return
	}
	defer icon.Dispose()
	a.window.SetIcon(icon)
	tray, err := walk.NewNotifyIcon(a.window)
	if err != nil {
		a.fail(err)
		return
	}
	defer tray.Dispose()
	if err = tray.SetIcon(icon); err != nil {
		a.fail(err)
		return
	}
	tray.SetToolTip("ClipQR — cliquez pour ouvrir l’historique")
	show := func() { a.window.Show(); a.window.Activate(); a.search.SetFocus() }
	tray.MouseDown().Attach(func(x, y int, b walk.MouseButton) {
		if b == walk.LeftButton {
			show()
		}
	})
	open := walk.NewAction()
	open.SetText("Ouvrir ClipQR")
	open.Triggered().Attach(show)
	tray.ContextMenu().Actions().Add(open)
	quit := walk.NewAction()
	quit.SetText("Quitter")
	quit.Triggered().Attach(func() { walk.App().Exit(0) })
	tray.ContextMenu().Actions().Add(quit)
	if err = tray.SetVisible(true); err != nil {
		a.fail(err)
		return
	}
	a.window.Closing().Attach(func(canceled *bool, reason walk.CloseReason) { *canceled = true; a.window.Hide() })
	a.render()
	if loadErr != nil {
		a.fail(fmt.Errorf("Impossible de charger l’historique : %w", loadErr))
	}
	stop := make(chan struct{})
	defer close(stop)
	go func() {
		ticker := time.NewTicker(400 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				a.window.Synchronize(a.capture)
			}
		}
	}()
	a.window.Run()
}

func (a *app) fail(err error) { walk.MsgBox(a.window, "ClipQR", err.Error(), walk.MsgBoxIconError) }
func (a *app) persist() {
	if err := a.data.save(a.path); err != nil {
		a.fail(fmt.Errorf("Impossible d’enregistrer l’historique : %w", err))
	}
}
func (a *app) capture() {
	seq, _, _ := sequence.Call()
	if seq == a.lastSequence {
		return
	}
	if a.paused {
		a.lastSequence = seq
		return
	}
	text, err := walk.Clipboard().Text()
	if err != nil {
		return
	}
	a.lastSequence = seq
	if a.data.add(text) {
		a.persist()
		a.render()
	}
}
func (a *app) copy(text string) {
	if err := walk.Clipboard().SetText(text); err != nil {
		a.fail(err)
		return
	}
	a.lastSequence, _, _ = sequence.Call()
	a.window.Hide()
	// Defer rebuilding controls until the originating click handler has returned.
	a.window.Synchronize(func() {
		if a.data.add(text) {
			a.persist()
		}
		a.render()
	})
}
func (a *app) render() {
	if a.rows == nil {
		return
	}
	a.rows.SetSuspended(true)
	defer a.rows.SetSuspended(false)
	for a.rows.Children().Len() > 0 {
		a.rows.Children().At(0).Dispose()
	}
	query := strings.ToLower(a.search.Text())
	count := 0
	for _, text := range a.data.Items {
		if !strings.Contains(strings.ToLower(text), query) {
			continue
		}
		count++
		text := text
		err := (Composite{Layout: Grid{Columns: 2, MarginsZero: true}, Children: []Widget{
			PushButton{Text: strings.ReplaceAll(preview(text), "&", "&&"), ToolTipText: text, Column: 0, StretchFactor: 1, MinSize: Size{Width: 380, Height: 42}, OnClicked: func() { a.copy(text) }},
			PushButton{Text: "QR code", Column: 1, MinSize: Size{Width: 85, Height: 42}, MaxSize: Size{Width: 85, Height: 0}, OnClicked: func() { a.qr(text) }},
		}}).Create(NewBuilder(a.rows))
		if err != nil {
			a.status.SetText(err.Error())
			return
		}
	}
	if count == 0 {
		msg := "Copiez un texte pour commencer (Ctrl+C)."
		if query != "" {
			msg = "Aucun texte ne correspond à cette recherche."
		}
		(Label{Text: msg}).Create(NewBuilder(a.rows))
	}
	(VSpacer{}).Create(NewBuilder(a.rows))
	state := "Collecte active"
	if a.paused {
		state = "Collecte suspendue"
	}
	a.status.SetText(fmt.Sprintf("%s  •  %d / %d textes affichés", state, count, len(a.data.Items)))
}
func (a *app) qr(text string) {
	qr, err := qrcode.New(text, qrcode.Medium)
	if err != nil {
		walk.MsgBox(a.window, "Texte trop long", "Ce texte dépasse la capacité d’un QR code. Copiez un extrait plus court.", walk.MsgBoxIconInformation)
		return
	}
	bitmap, err := walk.NewBitmapFromImage(qr.Image(420))
	if err != nil {
		a.fail(err)
		return
	}
	defer bitmap.Dispose()
	var dialog *walk.Dialog
	_, err = (Dialog{AssignTo: &dialog, Title: "ClipQR — Scanner avec le téléphone", Size: Size{Width: 480, Height: 540}, MinSize: Size{Width: 460, Height: 500}, Layout: VBox{}, Children: []Widget{
		Label{Text: "Scannez ce QR code pour récupérer le texte.", Font: Font{PointSize: 11, Bold: true}},
		ImageView{Image: bitmap, Mode: ImageViewModeShrink, MinSize: Size{Width: 420, Height: 420}},
		Label{Text: "Le texte est dans le QR : aucune connexion requise."},
		PushButton{Text: "Fermer", OnClicked: func() { dialog.Accept() }},
	}}).Run(a.window)
	if err != nil {
		a.fail(err)
	}
}
func makeIcon() (*walk.Icon, error) {
	img := image.NewRGBA(image.Rect(0, 0, 32, 32))
	for y := 0; y < 32; y++ {
		for x := 0; x < 32; x++ {
			c := color.RGBA{26, 71, 115, 255}
			if x >= 8 && x < 24 && y >= 6 && y < 27 {
				c = color.RGBA{245, 250, 255, 255}
			}
			if x >= 11 && x < 21 && (y == 13 || y == 18 || y == 23) {
				c = color.RGBA{26, 71, 115, 255}
			}
			img.SetRGBA(x, y, c)
		}
	}
	return walk.NewIconFromImageForDPI(img, 96)
}
