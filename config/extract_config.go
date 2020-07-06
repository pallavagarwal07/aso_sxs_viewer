package config

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

func (vc *ViewerConfig) GetBrowserConfigList() []BrowserConfig {
	windowsOverrides := vc.GetWindowOverrides()
	BrowserCount := vc.GetBrowserCount()
	var BrowserList []BrowserConfig
	var tempBrowser BrowserConfig
	var i int

	overrideLength := min(len(windowsOverrides), BrowserCount)
	for i = 0; i < overrideLength; i++ {
		if sel := windowsOverrides[i].GetCssSelector().GetSelector(); sel != "" {
			tempBrowser.Selector = sel
			tempBrowser.Position = int(windowsOverrides[i].GetCssSelector().GetPosition())
		} else {
			tempBrowser.Selector = vc.GetCssSelector().GetSelector()
			tempBrowser.Position = int(vc.GetCssSelector().GetPosition())
		}

		if url := windowsOverrides[i].GetUrl(); url != "" {
			tempBrowser.URL = url
		} else {
			tempBrowser.URL = vc.GetUrl()
		}
		BrowserList = append(BrowserList, tempBrowser)
	}

	tempBrowser.Selector = vc.GetCssSelector().GetSelector()
	tempBrowser.Position = int(vc.GetCssSelector().GetPosition())
	tempBrowser.URL = vc.GetUrl()

	for ; i < BrowserCount; i++ {
		BrowserList = append(BrowserList, tempBrowser)
	}
	return BrowserList
}
