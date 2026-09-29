package main

import (
	"errors"
	"image"
)

func grabScreen() (*image.RGBA, error) {
	return nil, errors.New("Capture d’écran non prise en charge sur macOS.")
}
