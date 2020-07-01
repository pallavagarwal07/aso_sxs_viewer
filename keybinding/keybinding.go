package keybinding

import (
	"fmt"
	"unicode"

	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
)

type KeyPressEvent struct {
	*xproto.KeyPressEvent
}

var (
	KeyMap *xproto.GetKeyboardMappingReply
	ModMap *xproto.GetModifierMappingReply
)

func InterpretKeyPressEvent(X *xgb.Conn, e KeyPressEvent) (str string, modifers uint16) {
	return InterpretKeycode(X, e.State, e.Detail), e.State
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

//UpdateMaps updates our view of Keyboard and Modifier Mapping
func UpdateMaps(X *xgb.Conn) error {
	min, max := getMinMaxKeycode(X)
	newKeymap, keyErr := xproto.GetKeyboardMapping(X, min,
		byte(max-min+1)).Reply()
	newModmap, modErr := xproto.GetModifierMapping(X).Reply()

	// We can't do any key binding without a mapping from the server.
	if keyErr != nil {
		return fmt.Errorf("COULD NOT GET KEYBOARD MAPPING: %v\n"+
			"UNRECOVERABLE ERROR.\n",
			keyErr)
	}
	if modErr != nil {
		return fmt.Errorf("COULD NOT GET MODIFIER MAPPING: %v\n"+
			"UNRECOVERABLE ERROR.\n",
			modErr)
	}

	KeyMap = newKeymap
	ModMap = newModmap
	return nil
}

// getMinMaxKeycode returns the minimum and maximum keycodes. They are typically 8 and 255, respectively.
func getMinMaxKeycode(X *xgb.Conn) (xproto.Keycode, xproto.Keycode) {
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
	min, _ := getMinMaxKeycode(X)
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
