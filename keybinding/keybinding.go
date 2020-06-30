package keybinding

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"strings"
	"unicode"

	"../chrometool"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/input"
	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/kb"
	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
)

type KeyPressEvent struct {
	*xproto.KeyPressEvent
}

var (
	KeyMap      *xproto.GetKeyboardMappingReply
	ModMap      *xproto.GetModifierMappingReply
	BrowserList []context.Context
	Focus       bool
)

// to be removed after made configurable
var (
	sel      = `input`
	nthchild = 7
)

func KeyPressHandler(X *xgb.Conn, e KeyPressEvent) error {
	if !Focus {
		for _, ctx := range BrowserList {
			if err := chromedp.Run(ctx,
				chrometool.ClickNthElement(sel, nthchild, chromedp.ByQueryAll),
			); err != nil {
				return err
			}
		}
		Focus = true
	}

	str := InterpretKeycode(X, e.State, e.Detail)

	for _, ctx := range BrowserList {
		ctx = cdp.WithExecutor(ctx, chromedp.FromContext(ctx).Target)
		keyEvents, err := dispatchKeyEvent(ctx, str, e.State)
		if err != nil {
			return err
		}
		for _, k := range keyEvents {
			if err := k.Do(ctx); err != nil {
				return err
			}
		}
	}

	return nil
}

// TODO: implement numlock and other modifiers
// InterpretKeycode as given in https://tronche.com/gui/x/xlib/input/keyboard-encoding.html
func InterpretKeycode(X *xgb.Conn, modifiers uint16, keycode xproto.Keycode) string {

	keySym1, keySym2, _, _ := interpretSymList(X, keycode)

	shift := modifiers&xproto.ModMaskShift > 0
	lock := modifiers&xproto.ModMaskLock > 0

	switch {
	case !shift && !lock:
		return keySym1
	case !shift && lock:
		if len(keySym1) == 1 && unicode.IsLower(rune(keySym1[0])) {
			return keySym2
		} else {
			return keySym1
		}
	case shift && lock:
		if len(keySym2) == 1 && unicode.IsLower(rune(keySym2[0])) {
			return string(unicode.ToUpper(rune(keySym2[0])))
		} else {
			return keySym2
		}
	case shift:
		return keySym2
	}

	return ""
}

func dispatchKeyEvent(ctx context.Context, str string, modifiers uint16) ([]*input.DispatchKeyEventParams, error) {
	r := keyStrToRune[str]

	// force \n -> \r
	if r == '\n' {
		r = '\r'
	}

	// if not known key, encode as unidentified
	v, ok := kb.Keys[r]
	if !ok {
		return kb.EncodeUnidentified(r), nil
	}

	keyDown := input.DispatchKeyEventParams{
		Key:                   v.Key,
		Code:                  v.Code,
		NativeVirtualKeyCode:  v.Native,
		WindowsVirtualKeyCode: v.Windows,
		Text:                  v.Text,
		UnmodifiedText:        v.Unmodified,
	}
	if runtime.GOOS == "darwin" {
		keyDown.NativeVirtualKeyCode = 0
	}
	if v.Shift {
		keyDown.Modifiers |= input.ModifierShift
	}
	keyDown.Modifiers |= keyEventModifier(modifiers)

	if keyDown.Modifiers&input.ModifierCommand > 0 && runtime.GOOS == "darwin" {
		if err := clipboardAction(ctx, str, keyDown.Modifiers); err != nil {
			return nil, err
		}
		return nil, nil
	}

	keyUp := keyDown
	keyDown.Type, keyUp.Type = input.KeyDown, input.KeyUp

	return []*input.DispatchKeyEventParams{&keyDown, &keyUp}, nil
}

func keyEventModifier(modifiers uint16) input.Modifier {
	//Bit field representing pressed modifier keys. Alt=1, Ctrl=2, Meta/Command=4, Shift=8 (default: 0).
	var mod input.Modifier

	if modifiers&xproto.ModMaskShift > 0 {
		mod |= input.ModifierShift
	}
	if modifiers&xproto.ModMaskControl > 0 {
		mod |= input.ModifierCtrl
	}
	if modifiers&xproto.ModMask1 > 0 {
		mod |= input.ModifierAlt
	}
	if modifiers&xproto.ModMask2 > 0 && runtime.GOOS == "darwin" {
		mod |= input.ModifierCommand
	}

	return mod
}

//UpdateMaps updates our view of Keyboard and Modifier Mapping
func UpdateMaps(X *xgb.Conn) error {
	min, max := GetMinMaxKeycode(X)
	newKeymap, keyErr := xproto.GetKeyboardMapping(X, min,
		byte(max-min+1)).Reply()
	newModmap, modErr := xproto.GetModifierMapping(X).Reply()

	// We can't do any key binding without a mapping from the server.
	if keyErr != nil {
		return errors.New(fmt.Sprintf("COULD NOT GET KEYBOARD MAPPING: %v\n"+
			"UNRECOVERABLE ERROR.\n",
			keyErr))
	}
	if modErr != nil {
		return errors.New(fmt.Sprintf("COULD NOT GET MODIFIER MAPPING: %v\n"+
			"UNRECOVERABLE ERROR.\n",
			modErr))
	}

	KeyMap = newKeymap
	ModMap = newModmap
	return nil
}

// GetMinMaxKeycode returns the minimum and maximum keycodes. They are typically 8 and 255, respectively.
func GetMinMaxKeycode(X *xgb.Conn) (xproto.Keycode, xproto.Keycode) {
	return xproto.Setup(X).MinKeycode, xproto.Setup(X).MaxKeycode
}

func interpretSymList(X *xgb.Conn, keycode xproto.Keycode) (k1, k2, k3, k4 string) {

	ks1 := GetKeysymFromMap(X, keycode, 0)
	ks2 := GetKeysymFromMap(X, keycode, 1)
	ks3 := GetKeysymFromMap(X, keycode, 2)
	ks4 := GetKeysymFromMap(X, keycode, 3)

	// follow the rules, third paragraph
	switch {
	case ks2 == 0 && ks3 == 0 && ks4 == 0:
		ks3 = ks1
	case ks3 == 0 && ks4 == 0:
		ks3 = ks1
		ks4 = ks2
	case ks4 == 0:
		ks4 = 0
	}

	// Now convert keysyms to strings, so we can do alphabetic shit.
	k1 = GetStrFromKeysym(ks1)
	k2 = GetStrFromKeysym(ks2)
	k3 = GetStrFromKeysym(ks3)
	k4 = GetStrFromKeysym(ks4)

	// follow the rules, fourth paragraph
	if k2 == "" {
		if len(k1) == 1 && unicode.IsLetter(rune(k1[0])) {
			k1 = string(unicode.ToLower(rune(k1[0])))
			k2 = string(unicode.ToUpper(rune(k1[0])))
		} else {
			k2 = k1
		}
	}
	if k4 == "" {
		if len(k3) == 1 && unicode.IsLetter(rune(k3[0])) {
			k3 = string(unicode.ToLower(rune(k3[0])))
			k4 = string(unicode.ToUpper(rune(k4[0])))
		} else {
			k4 = k3
		}
	}
	return
}

// GetKeysymFromMap uses the KeyMap and finds a keysym associated
// with the given keycode in the current X environment.
func GetKeysymFromMap(X *xgb.Conn, keycode xproto.Keycode, column byte) xproto.Keysym {
	min, _ := GetMinMaxKeycode(X)
	i := (int(keycode)-int(min))*int(KeyMap.KeysymsPerKeycode) + int(column)

	return KeyMap.Keysyms[i]
}

func GetStrFromKeysym(keysym xproto.Keysym) string {
	str, ok := keysymsToStr[keysym]
	if !ok {
		return ""
	}

	symbol, ok := strToSymbol[str]
	if ok {
		str = string(symbol)
	}

	return str
}

func clipboardAction(ctx context.Context, str string, modifiers input.Modifier) error {
	str = strings.ToLower(str)
	isShift := modifiers & input.ModifierShift

	switch str {
	case "c":
		if err := chrometool.ClipboardCommand(ctx, "copy"); err != nil {
			return err
		}

	// TODO
	// case "v":
	// 	if err := chrometool.ClipboardCommand(ctx, "paste"); err != nil {
	// 		return err
	// 	}

	case "x":
		if err := chrometool.ClipboardCommand(ctx, "cut"); err != nil {
			return err
		}

	case "a":
		if err := chrometool.SelectNthElement(ctx, sel, nthchild); err != nil {
			return err
		}

	case "z":
		if isShift > 0 {
			if err := chrometool.ClipboardCommand(ctx, "redo"); err != nil {
				return err
			}
		} else {
			if err := chrometool.ClipboardCommand(ctx, "undo"); err != nil {
				return err
			}
		}
	}

	return nil
}
