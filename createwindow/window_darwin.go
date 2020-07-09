// +build darwin

package createwindow

import (
	"fmt"
	"log"
	"strconv"

	"github.com/googleinterns/aso_sxs_viewer/command"
)

// ChromeCommand will take structured data after config file is implemented.
func ChromeCommand(layout Layout, userdatadir, display string, debuggingPort int) command.ExternalCommand {
	cmd := command.ExternalCommand{
		Path: "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
		Arg: []string{"--user-data-dir=" + userdatadir,
			fmt.Sprintf("--window-position=%d,%d", layout.x, layout.y),
			fmt.Sprintf("--window-size=%d,%d", layout.w, layout.h),
			"--disable-session-crashed-bubble", "--disble-infobars", "--disable-extensions",
			fmt.Sprintf("--remote-debugging-port=%d", debuggingPort),
		},
	}
	return cmd
}

func DisplayError(err error, cmdErrorHandler func(err error) error, isFatal bool) error {
	if err == nil {
		return nil
	}

	arg := displayErrorArg(err, isFatal)
	applescript := command.ExternalCommand{
		Path: "osascript",
		Arg:  arg,
	}
	if _, err := command.ExecuteProgram(applescript, cmdErrorHandler); err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func displayErrorArg(err error, isFatal bool) []string {
	icon := "caution"
	if isFatal {
		icon = "stop"
	}

	arg := []string{"-e", fmt.Sprintf("display dialog %s with icon %s", strconv.Quote(err.Error()), icon)}
	return arg
}
