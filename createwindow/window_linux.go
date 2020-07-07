// +build linux

package createwindow

import (
	"fmt"
	"log"

	"github.com/googleinterns/aso_sxs_viewer/command"
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
	var messageType string
	if isFatal {
		messageType = "error"
	} else {
		messageType = "warning"
	}

	zenity := command.ExternalCommand{
		Path: "zenity",
		Arg: []string{
			fmt.Sprintf("--%s", messageType),
			"--title",
			fmt.Sprintf("\"%s message\"", messageType),
			"--text", fmt.Sprintf("<span font= \"14\">%s</span>", err),
			"--width=500",
		},
	}
	if _, err := command.ExecuteProgram(zenity, cmdErrorHandler); err != nil {
		log.Println(err)
		return err
	}
	return nil
}
