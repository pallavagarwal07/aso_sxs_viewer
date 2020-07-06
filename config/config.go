package config

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/googleinterns/aso_sxs_viewer/proto"
	"google.golang.org/protobuf/encoding/prototext"
)

type ViewerConfig struct {
	*proto.ViewerConfig
}

type BrowserConfig struct {
	Selector string
	Position int
	URL      string
}

// GetConfig generates .aso_sxs_viewer directory and config file, if it doesn't already exists.
// It validates the pre-existing config if found.
// You should avoid repeated calls to avoid validation overhead.
func GetConfig() (ViewerConfig, error) {
	if err := createDir(AsoSxSViewerDir); err != nil {
		return ViewerConfig{nil}, fmt.Errorf("Error %s encountered while creating aso_sxs_viewer directory", err.Error())
	}

	return createOrValidateConfig(AsoSxSViewerConfig)
}

func createDir(dirPath string) error {
	// check if directory exists
	_, err := os.Stat(dirPath)

	// create directory if does not exist
	if os.IsNotExist(err) {
		err = os.MkdirAll(dirPath, 0755)
		if err != nil {
			return err
		}
		return nil
	} else if err != nil {
		return err
	}

	return nil
}

func createOrValidateConfig(filePath string) (ViewerConfig, error) {
	// check if file exists
	_, err := os.Stat(filePath)

	// create file if not exists
	if os.IsNotExist(err) {
		file, err := os.Create(filePath)
		defer file.Close()
		if err != nil {
			return ViewerConfig{nil}, err
		}
		return generateConfig(filePath)
	} else if err != nil {
		return ViewerConfig{nil}, err
	}
	return validateConfig(filePath)
}

func generateConfig(configPath string) (ViewerConfig, error) {
	viewerConfig := &proto.ViewerConfig{}

	viewerConfig.CssSelector = &proto.CSSSelector{
		Selector: &DefaultCSSSelector.Selector,
		Position: &DefaultCSSSelector.Position,
	}
	viewerConfig.Url = &DefaultURL
	viewerConfig.BrowserWindowCount = &DefaultBrowserCount
	viewerConfig.UseCookies = &DefaultUseCookies
	viewerConfig.UserDataDirPrefix = &DefaultUserDataDirPrefix

	DefaultInputWindowPosition := proto.ViewerConfig_BOTTOM
	viewerConfig.InputWindowPosition = &DefaultInputWindowPosition

	rootWindowLayout := &proto.Layout{
		Width:  &DefaultRootWindowWidth,
		Height: &DefaultRootWindowHeight,
	}
	viewerConfig.RootWindowConfig = &proto.RootWindowConfig{
		Layout: rootWindowLayout,
	}

	marshalOpts := prototext.MarshalOptions{Multiline: true, Indent: "\t"}
	out, err := marshalOpts.Marshal(viewerConfig)
	if err != nil {
		return ViewerConfig{nil}, fmt.Errorf("Failed to encode aso_sxs_viewer config %s", err)
	}
	out = append(out, []byte(BrowserWindowExample)...)
	if err := ioutil.WriteFile(configPath, out, 0644); err != nil {
		return ViewerConfig{nil}, fmt.Errorf("Failed to write aso_sxs_viewer config %s", err)
	}
	return ViewerConfig{viewerConfig}, nil
}

// validateConfig validate the given configFile. Any assumptions made to handle essential empty fields are written back to the config file.
func validateConfig(configPath string) (ViewerConfig, error) {
	var writeChanges bool
	in, err := ioutil.ReadFile(configPath)
	if err != nil {
		return ViewerConfig{nil}, fmt.Errorf("Failed to read aso_sxs_viewer config %s", err)
	}

	viewerConfig := &proto.ViewerConfig{}

	if err := prototext.Unmarshal(in, viewerConfig); err != nil {
		return ViewerConfig{nil}, fmt.Errorf("Failed to parse aso_sxs_viewer config %s", err)
	}

	if viewerConfig == nil {
		return ViewerConfig{nil}, fmt.Errorf("Found an empty config file, please delete the existing config and try again")
	}

	if viewerConfig.BrowserWindowCount == nil {
		viewerConfig.BrowserWindowCount = &DefaultBrowserCount
		writeChanges = true
	} else if viewerConfig.GetBrowserWindowCount() <= 0 {
		return ViewerConfig{nil}, fmt.Errorf("A positive browser_window_count field is required in the config.textproto file")
	}

	// If both the CSSSelector and url are nil, try to populate them using the first WindowOverrides
	// Otherwise use the populate them using the default values.
	// If only one of the fields is non nil throw an error.
	if CSSSelector, url := viewerConfig.GetCssSelector().GetSelector(), viewerConfig.GetUrl(); CSSSelector == "" && url == "" {
		if windowsOverrides := viewerConfig.GetWindowOverrides(); windowsOverrides != nil && windowsOverrides[0].GetCssSelector().GetSelector() != "" && windowsOverrides[0].GetUrl() != "" {
			viewerConfig.CssSelector = windowsOverrides[0].GetCssSelector()
			url = windowsOverrides[0].GetUrl()
			viewerConfig.Url = &url
		} else {
			viewerConfig.CssSelector = &proto.CSSSelector{
				Selector: &DefaultCSSSelector.Selector,
				Position: &DefaultCSSSelector.Position,
			}
			viewerConfig.Url = &DefaultURL
		}
		writeChanges = true
	} else if CSSSelector == "" || url == "" {
		return ViewerConfig{nil}, fmt.Errorf("A empty css_selector or url found, populate them in the config.textproto file and try again")
	}

	if rootWindow := viewerConfig.GetRootWindowConfig(); rootWindow != nil {
		if layout := rootWindow.GetLayout(); layout != nil {
			if layout.GetX() < 0 || layout.GetY() < 0 {
				return ViewerConfig{nil}, fmt.Errorf("Invalid x or y in root_window_config layout, a non-negative int is expected in the config.textproto file")
			}
			if layout.GetWidth() < 0 || layout.GetHeight() < 0 {
				return ViewerConfig{nil}, fmt.Errorf("Invalid height or width in root_window_config layout, a non-negative int is expected in the config.textproto file")
			}
		}
	}

	if writeChanges {
		marshalOpts := prototext.MarshalOptions{Multiline: true, Indent: "\t"}
		out, err := marshalOpts.Marshal(viewerConfig)
		if err != nil {
			return ViewerConfig{nil}, fmt.Errorf("Failed to encode aso_sxs_viewer config %s", err)
		}
		if err := ioutil.WriteFile(configPath, out, 0644); err != nil {
			return ViewerConfig{nil}, fmt.Errorf("Failed to write aso_sxs_viewer config %s", err)
		}
	}
	return ViewerConfig{viewerConfig}, nil
}
