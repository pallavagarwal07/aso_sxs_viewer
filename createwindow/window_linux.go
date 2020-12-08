// +build linux

package createwindow

import (
	"fmt"
	"log"

	"github.com/pallavagarwal07/aso_sxs_viewer/command"
)

// ChromeCommand will take structured data after config file is implemented.
func ChromeCommand(layout Layout, userdatadir, display string, debuggingPort int) command.ExternalCommand {
	cmd := command.ExternalCommand{
		Path: "google-chrome",
		Arg: []string{
			"--user-data-dir=" + userdatadir,
			fmt.Sprintf("--window-position=%d,%d", layout.x, layout.y),
			fmt.Sprintf("--window-size=%d,%d", layout.w, layout.h),
			"--disable-session-crashed-bubble",
			"--disble-infobars",
			"--disable-extensions",
			fmt.Sprintf("--remote-debugging-port=%d", debuggingPort),
		},
		Env: []string{"DISPLAY=" + display},
	}
	return cmd
}

func DisplayError(err error, cmdErrorHandler func(err error) error, isFatal bool) error {
	if err == nil {
		return nil
	}

	arg := displayErrorArg(err, isFatal)
	zenity := command.ExternalCommand{
		Path: "zenity",
		Arg: arg,
	}
	if _, err := command.ExecuteProgram(zenity, cmdErrorHandler); err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func displayErrorArg(err error, isFatal bool) []string {
	messageType := "warning"
	if isFatal {
		messageType = "error"
	}
	
	arg := []string{
		fmt.Sprintf("--%s", messageType),
		"--title", fmt.Sprintf("%s message", messageType),
		"--text", fmt.Sprintf(`<span font= "14">%s</span>`, err),
		"--width", "500",
	}
	return arg 
}