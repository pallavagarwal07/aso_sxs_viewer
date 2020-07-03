package config

import (
	"fmt"
	"io/ioutil"
	"os"

	"../proto"
	"google.golang.org/protobuf/encoding/prototext"
)

// GetConfig generates .aso_sxs_viewer directory and config file, if it doesn't already exists.
// It validates the pre-existing config if found.
// You should avoid repeated calls to avoid validation overhead.
func GetConfig() (*proto.ViewerConfig, error) {
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

func createOrValidateConfig(filePath string) (*proto.ViewerConfig, error) {
	// check if file exists
	_, err := os.Stat(filePath)

	// create file if not exists
	if os.IsNotExist(err) {
		file, err := os.Create(filePath)
		defer file.Close()
		if err != nil {
			return nil, err
		}
		return generateConfig(filePath)
	} else if err != nil {
		return nil, err
	}
	return validateConfig(filePath)
}

func generateConfig(configPath string) (*proto.ViewerConfig, error) {
	viewerConfig := &proto.ViewerConfig{}

	viewerConfig.DefaultSelector = &proto.CSSSelector{
		Selector: &DefaultCSSSelector.Selector,
		Position: &DefaultCSSSelector.Position,
	}

	viewerConfig.Default_URL = &DefaultURL
	viewerConfig.BrowserWindowCount = &DefaultBrowserCount
	viewerConfig.UseCookies = &DefaultUseCookies
	DefaultInputWindowPosition := proto.ViewerConfig_BOTTOM
	viewerConfig.InputWindowPosition = &DefaultInputWindowPosition

	xephyrLayout := &proto.Layout{
		Width:  &DefaultXephyrWidth,
		Height: &DefaultXephyrHeight,
	}

	viewerConfig.XephyrConfig = &proto.XephyrConfig{
		Layout: xephyrLayout,
	}

	marshalOpts := prototext.MarshalOptions{Multiline: true}
	out, err := marshalOpts.Marshal(viewerConfig)
	if err != nil {
		return nil, fmt.Errorf("Failed to encode aso_sxs_viewer config %s", err)
	}
	if err := ioutil.WriteFile(configPath, out, 0644); err != nil {
		return nil, fmt.Errorf("Failed to write aso_sxs_viewer config %s", err)
	}
	return viewerConfig, nil
}

func validateConfig(configPath string) (*proto.ViewerConfig, error) {
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

	if browserCount := viewerConfig.GetBrowserWindowCount(); browserCount <= 0 {
		return nil, fmt.Errorf("A positive browser_window_count field is required in the config.textproto file")
	}

	if CSSSelector := viewerConfig.GetDefaultSelector(); CSSSelector.GetSelector() == "" {
		if browserWindows := viewerConfig.GetBrowserWindow(); browserWindows == nil {
			return nil, fmt.Errorf("No valid CSS Selector found")
		} else if browserWindows[0].GetOverrideCssSelector().GetSelector() == "" {
			return nil, fmt.Errorf("No valid CSS Selector found")
		} else {
			//Set this CSS Selector as default to use in case a browserWindow does not have a selector.
			viewerConfig.DefaultSelector = browserWindows[0].GetOverrideCssSelector()
		}
	}

	if url := viewerConfig.GetDefault_URL(); url == "" {
		if browserWindows := viewerConfig.GetBrowserWindow(); browserWindows == nil {
			return nil, fmt.Errorf("No valid URL found")
		} else if browserWindows[0].GetOverride_URL() == "" {
			return nil, fmt.Errorf("No valid URL found")
		} else {
			//Set this URL as default to use in case a browserWindow does not have a URL.
			url = browserWindows[0].GetOverride_URL()
			viewerConfig.Default_URL = &url
		}
	}

	return viewerConfig, nil
}
