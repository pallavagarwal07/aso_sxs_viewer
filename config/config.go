package config

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/pallavagarwal07/aso_sxs_viewer/proto"
	"google.golang.org/protobuf/encoding/prototext"
)

type ViewerConfig struct {
	proto.ViewerConfig
}

type BrowserConfig struct {
	proto.BrowserConfig
}

// GetConfig generates .aso_sxs_viewer directory and config file, if it doesn't already exists.
// It validates the pre-existing config if found.
// You should avoid repeated calls to avoid validation overhead.
func GetConfig() (*ViewerConfig, error) {
	if err := createDir(AsoSxSViewerDir); err != nil {
		return nil, fmt.Errorf("Error %s encountered while creating aso_sxs_viewer directory", err.Error())
	}

	return createOrValidateConfig(AsoSxSViewerConfig)
}

func createDir(dirPath string) error {
	// check if directory exists
	_, err := os.Stat(dirPath)

	// create directory if does not exist
	if os.IsNotExist(err) {
		return os.MkdirAll(dirPath, 0755)
	}
	return err
}

func createOrValidateConfig(filePath string) (*ViewerConfig, error) {
	// check if file exists
	_, err := os.Stat(filePath)

	// create file if not exists
	if os.IsNotExist(err) {
		_, err := os.Create(filePath)
		if err != nil {
			return nil, err
		}

		viewerConfig, err := generateDefaultConfig(filePath)
		if err != nil {
			return nil, err
		}
		marshalOpts := prototext.MarshalOptions{Multiline: true, Indent: "\t"}
		out, err := marshalOpts.Marshal(viewerConfig)
		if err != nil {
			return nil, fmt.Errorf("Failed to encode aso_sxs_viewer config %s", err)
		}
		out = append(out, []byte(BrowserWindowExample)...)
		if err := ioutil.WriteFile(filePath, out, 0644); err != nil {
			return nil, fmt.Errorf("Failed to write aso_sxs_viewer config %s", err)
		}

		return &ViewerConfig{*viewerConfig}, nil
	} else if err != nil {
		return nil, err
	}
	return validateConfig(filePath)
}

func generateDefaultConfig(configPath string) (*proto.ViewerConfig, error) {
	viewerConfig := &proto.ViewerConfig{}

	viewerConfig.Url = &DefaultURL
	viewerConfig.CssSelector = &proto.CSSSelector{
		Selector: &DefaultCSSSelector.Selector,
		Position: &DefaultCSSSelector.Position,
	}
	viewerConfig.BrowserWindowCount = &DefaultBrowserCount
	viewerConfig.UserDataDirPrefix = &DefaultUserDataDirPrefix

	DefaultInputWindowPosition := proto.ViewerConfig_DEFAULT_BOTTOM
	viewerConfig.InputWindowPosition = &DefaultInputWindowPosition

	rootWindowLayout := &proto.Layout{
		Width:  &DefaultRootWindowWidth,
		Height: &DefaultRootWindowHeight,
	}
	viewerConfig.RootWindowConfig = &proto.RootWindowConfig{
		Layout: rootWindowLayout,
	}
	return viewerConfig, nil
}

// validateConfig validates the given configFile.
func validateConfig(configPath string) (*ViewerConfig, error) {
	in, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("Failed to read aso_sxs_viewer config %s", err)
	}

	viewerConfig := &proto.ViewerConfig{}

	if err := prototext.Unmarshal(in, viewerConfig); err != nil {
		return nil, fmt.Errorf("Failed to parse aso_sxs_viewer config %s", err)
	}

	if viewerConfig == nil {
		return nil, fmt.Errorf("Found an empty config file, please delete the existing config and try again")
	}

	if viewerConfig.BrowserWindowCount == nil {
		viewerConfig.BrowserWindowCount = &DefaultBrowserCount
	} else if viewerConfig.GetBrowserWindowCount() <= 0 {
		return nil, fmt.Errorf("A positive browser_window_count field is required in the config.textproto file")
	}

	// If either the CSSSelector and url are nil, we try to iterate through the overriden windows to check
	// if they have a missing field which does not have a default value, and throw an error if such a window is found.
	if CSSSelector, url := viewerConfig.GetCssSelector().GetSelector(), viewerConfig.GetUrl(); CSSSelector == "" || url == "" {
		windowsOverrides := viewerConfig.GetWindowOverrides()
		overrideLength := min(len(windowsOverrides), int(viewerConfig.GetBrowserWindowCount()))
		for i := 0; i < overrideLength; i++ {
			if url == "" && windowsOverrides[i].GetUrl() == "" {
				return nil, fmt.Errorf("A empty css_selector or url found, try populating the default values in the config.textproto file and try again")
			}
			if CSSSelector == "" && windowsOverrides[i].GetCssSelector().GetSelector() == "" {
				return nil, fmt.Errorf("A empty css_selector or url found, try populating the default values in the config.textproto file and try again")
			}
		}
	}

	if rootWindow := viewerConfig.GetRootWindowConfig(); rootWindow != nil {
		if layout := rootWindow.GetLayout(); layout != nil {
			if layout.GetX() < 0 || layout.GetY() < 0 {
				return nil, fmt.Errorf("Invalid x or y in root_window_config layout, a non-negative int is expected in the config.textproto file")
			}
			if layout.GetWidth() < 0 || layout.GetHeight() < 0 {
				return nil, fmt.Errorf("Invalid height or width in root_window_config layout, a non-negative int is expected in the config.textproto file")
			}
		}
	}
	return &ViewerConfig{*viewerConfig}, nil
}
