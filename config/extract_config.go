package config

import (
	"github.com/pallavagarwal07/aso_sxs_viewer/proto"
)

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (vc *ViewerConfig) GetBrowserCount() int {
	return int(vc.GetBrowserWindowCount())
}

func (vc *ViewerConfig) GetInputWindowOrientation() string {
	if vc.GetInputWindowPosition() == 0 {
		return "BOTTOM"
	}
	return "TOP"
}

func (vc *ViewerConfig) GetRootWindowLayout() (x, y uint32, w, h uint16) {
	layout := vc.GetRootWindowConfig().GetLayout()
	x, y = uint32(layout.GetX()), uint32(layout.GetY())
	w, h = uint16(layout.GetWidth()), uint16(layout.GetHeight())

	if w == 0 || h == 0 {
		defaultWidth, defaultHeight := DefaultRootWindowSize()
		if w == 0 {
			w = defaultWidth
		}
		if h == 0 {
			h = defaultHeight
		}
	}
	return
}

func (vc *ViewerConfig) GetUserDataDirPath() string {
	if prefix := vc.GetUserDataDirPrefix(); prefix != "" {
		return prefix
	}
	return DefaultUserDataDirPrefix
}

func (vc *ViewerConfig) GetBrowserConfigList() []*BrowserConfig {
	windowsOverrides := vc.GetWindowOverrides()
	BrowserCount := vc.GetBrowserCount()
	var BrowserList []*BrowserConfig
	tempBrowser := &proto.BrowserConfig{}
	var i int

	overrideLength := min(len(windowsOverrides), BrowserCount)
	for i = 0; i < overrideLength; i++ {
		if sel := windowsOverrides[i].GetCssSelector().GetSelector(); sel != "" {
			tempBrowser.CssSelector = windowsOverrides[i].GetCssSelector()
		} else {
			tempBrowser.CssSelector = vc.GetCssSelector()
		}

		if url := windowsOverrides[i].GetUrl(); url != "" {
			tempBrowser.Url = windowsOverrides[i].Url
		} else {
			tempBrowser.Url = vc.Url
		}
		BrowserList = append(BrowserList, &BrowserConfig{*tempBrowser})
	}

	tempBrowser.CssSelector = vc.GetCssSelector()
	tempBrowser.Url = vc.Url

	for ; i < BrowserCount; i++ {
		BrowserList = append(BrowserList, &BrowserConfig{*tempBrowser})
	}
	return BrowserList
}
