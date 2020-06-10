package main

import (
	"fmt"
	"strconv"

	"aso_sxs_viewer/command"

	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/xproto"
)

/*this needs to be fixed*/
func cmdErrorHandler(p *command.ProgramState, err error) error {
	return err
}

func main() {
	X, err := xgb.NewConn()
	if err != nil {
		fmt.Println(err)
		return
	}
	setup := xproto.Setup(X)
	screenInfo := setup.DefaultScreen(X)

	heightScreen := screenInfo.HeightInPixels
	widthScreen := screenInfo.WidthInPixels

	h := int(heightScreen - 150)
	w := int(widthScreen / 2)

	fmt.Println(heightScreen, widthScreen)

	cmd := command.ExternalCommand{
		Path: "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
		Arg:  []string{"--window-position=0,0", "--window-size=" + strconv.Itoa(w) + "," + strconv.Itoa(h)},
	}

	programstate, err := command.ExecuteProgram(cmd, cmdErrorHandler)

	if err != nil {
		fmt.Println(err)
	}

	cmd2 := command.ExternalCommand{
		Path: "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
		Arg:  []string{"--user-data-dir=/Users/aditibhattacharya/chrome-dev-profile", "--window-position=" + strconv.Itoa(w) + ",0", "--window-size=" + strconv.Itoa(w) + "," + strconv.Itoa(h)},
	}
	programstate2, err2 := command.ExecuteProgram(cmd2, cmdErrorHandler)

	if err2 != nil {
		fmt.Println(err2)
	}
	
	command.CloseProgram(programstate2, cmdErrorHandler)
	command.CloseProgram(programstate, cmdErrorHandler)

}
