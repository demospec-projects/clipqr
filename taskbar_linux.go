//go:build linux && !gtk3

package main

/*
#cgo pkg-config: gtk4
#include <gtk/gtk.h>
#ifdef GDK_WINDOWING_X11
#include <gdk/x11/gdkx.h>
#endif

// skip_taskbar asks the window manager to leave the window out of the list of
// open windows and of the workspace switcher. GTK sends the hint again each
// time the window is shown.
static void skip_taskbar(void *window) {
#ifdef GDK_WINDOWING_X11
	gtk_widget_realize(GTK_WIDGET(window));
	GdkSurface *surface = gtk_native_get_surface(GTK_NATIVE(window));
	if (surface != NULL && GDK_IS_X11_SURFACE(surface)) {
		gdk_x11_surface_set_skip_taskbar_hint(surface, TRUE);
		gdk_x11_surface_set_skip_pager_hint(surface, TRUE);
	}
#endif
}
*/
import "C"

import "github.com/wailsapp/wails/v3/pkg/application"

// hideFromTaskbar: the panel belongs by the clock, not among the open
// windows. Under Wayland, an application cannot ask for it.
func hideFromTaskbar(panel *application.WebviewWindow) {
	application.InvokeSync(func() {
		if window := panel.NativeWindow(); window != nil {
			C.skip_taskbar(window)
		}
	})
}
