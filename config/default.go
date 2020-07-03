package config

import (
	"os"
	"path/filepath"
)

var (
	AsoSxSViewerDir    = filepath.Join(os.Getenv("HOME"), ".aso_sxs_viewer")
	AsoSxSViewerConfig = filepath.Join(AsoSxSViewerDir, "config.textproto")
	DefaultCSSSelector = struct {
		Selector string
		Position int32
	}{
		Selector: "input",
		Position: 7,
	}
	DefaultURL          = "https://mail.google.com/"
	DefaultBrowserCount = int32(2)
	DefaultUseCookies   = false
	DefaultXephyrWidth  = int32(1600)
	DefaultXephyrHeight = int32(900)
)
