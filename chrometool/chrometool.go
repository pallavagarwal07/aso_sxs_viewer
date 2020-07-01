package chrometool

import (
	"context"
	"fmt"
	"runtime"
	"strings"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/cdproto/input"
	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/kb"
	"github.com/jezek/xgb/xproto"
)

// NthChildSel hold information about the target node
// Sel refers to the CSS selector and
// position refers to their position among element nodes matching the selector.
type NthChildSel struct {
	Selector string
	Position int
}

const (
	windowDimensionsJS  = `[window.innerWidth, window.innerHeight];`
	windowScrollJS      = `window.scrollBy(%d, %d);`
	clipboardInteractJS = `document.execCommand("%s")`
	selectNthElementJS  = `document.querySelectorAll("%s")[%d].select()`
)

// SendKeysToNthElement is an element query action that synthesizes the key up, char, and down
// events as needed for the runes in v, sending them to the nth element node
// matching the selector.
func SendKeysToNthElement(sel NthChildSel, v string, opts ...chromedp.QueryOption) chromedp.QueryAction {
	return chromedp.QueryAfter(sel.Selector, func(ctx context.Context, nodes ...*cdp.Node) error {
		if len(nodes) < 1 {
			return fmt.Errorf("selector %q did not return any nodes", sel.Selector)
		}
		if len(nodes) < sel.Position {
			return fmt.Errorf("invalid position %d for selector %q", sel.Position, sel.Selector)
		}

		node := nodes[sel.Position]

		return keyEventNode(node, v).Do(ctx)
	}, append(opts, chromedp.NodeReady)...)
}

// keyEventNode is a key action that dispatches a key event on a element node.
func keyEventNode(n *cdp.Node, keys string, opts ...chromedp.KeyOption) chromedp.KeyAction {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		return chromedp.KeyEvent(keys, opts...).Do(ctx)
	})
}

// ClickNthElement is an element query action that sends a mouse click event to the nth element
// node matching the selector.
func ClickNthElement(sel NthChildSel, opts ...chromedp.QueryOption) chromedp.QueryAction {
	return chromedp.QueryAfter(sel.Selector, func(ctx context.Context, nodes ...*cdp.Node) error {
		if len(nodes) < 1 {
			return fmt.Errorf("selector %q did not return any nodes", sel.Selector)
		}
		if len(nodes) < sel.Position {
			return fmt.Errorf("invalid position %d for selector %q", sel.Position, sel.Selector)
		}
		return mouseClickNode(nodes[sel.Position]).Do(ctx)
	}, append(opts, chromedp.NodeReady)...)
}

// mouseClickNode is an action that dispatches a mouse left button click event
// at the center of a specified node.
//
// Note that the window will be scrolled if the node is not within the window's
// viewport.
func mouseClickNode(n *cdp.Node, opts ...chromedp.MouseOption) chromedp.MouseAction {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		var windowDim []float64
		if err := chromedp.Evaluate(windowDimensionsJS, &windowDim).Do(ctx); err != nil {
			return err
		}

		box, err := dom.GetBoxModel().WithNodeID(n.NodeID).Do(ctx)
		if err != nil {
			return err
		}

		x, y, err := evalNodeCenter(box)
		if err != nil {
			return err
		}

		if x > windowDim[0] || x < 0 || y > windowDim[1] || y < 0 {
			if err := ScrollNodeIntoView(ctx, n, x, y, windowDim, box.Border); err != nil {
				return err
			}
			box, err := dom.GetBoxModel().WithNodeID(n.NodeID).Do(ctx)
			if err != nil {
				return err
			}

			x, y, err = evalNodeCenter(box)
			if err != nil {
				return err
			}
		}

		return chromedp.MouseClickXY(x, y, opts...).Do(ctx)
	})
}

// ScrollNodeIntoView scrolles the window if the node is not within the window's
// viewport.
func ScrollNodeIntoView(ctx context.Context, n *cdp.Node, x float64, y float64, windowDim []float64, box []float64) error {
	var xshift, yshift int
	if x < 0 {
		xshift = int(box[0]) - 1
	} else if x > windowDim[0] {
		xshift = int(box[2]-windowDim[0]) + 1
	}

	if y < 0 {
		yshift = int(box[1]) - 1
	} else if y > windowDim[1] {
		yshift = int(box[5]-windowDim[1]) + 1
	}

	var res interface{}
	if err := chromedp.Evaluate(fmt.Sprintf(windowScrollJS, xshift, yshift), &res).Do(ctx); err != nil {
		if err.Error() == "encountered an undefined value" {
			err = nil
		}
		return err
	}
	return nil
}

// evalNodeCenter evaluates a node's central coordinates and returns x, y
func evalNodeCenter(box *dom.BoxModel) (float64, float64, error) {
	content := box.Border
	c := len(content)
	if c%2 != 0 || c < 1 {
		return 0, 0, chromedp.ErrInvalidDimensions
	}

	var x, y float64
	for i := 0; i < c; i += 2 {
		x += content[i]
		y += content[i+1]
	}
	x /= float64(c / 2)
	y /= float64(c / 2)

	return x, y, nil
}

func SelectNthElement(ctx context.Context, sel NthChildSel) error {
	var res interface{}
	if err := chromedp.Evaluate(fmt.Sprintf(selectNthElementJS, sel.Selector, sel.Position), &res).Do(ctx); err != nil {
		if err.Error() == "encountered an undefined value" {
			err = nil
		}
		return err
	}
	return nil
}

func ClipboardCommand(ctx context.Context, command string) error {
	var res interface{}
	if err := chromedp.Evaluate(fmt.Sprintf(clipboardInteractJS, command), &res).Do(ctx); err != nil {
		return err
	}
	return nil
}

// DispatchKeyEventToBrowser takes up the key's value to be dispatched to the browser usually the result of keybinding.InterpretKeycode
// It clicks on the element if isFocussed is false and then sends the appropiate key with the modifiers
// The modifiers here refer to the xproto modifiers not to be confused with the ones in input.DispatchKeyEvent
func DispatchKeyEventToBrowser(ctx context.Context, sel NthChildSel, str string, modifiers uint16, isFocussed bool) error {
	if isFocussed {
		if err := chromedp.Run(ctx, ClickNthElement(sel, chromedp.ByQueryAll)); err != nil {
			return err
		}
	}

	ctx = cdp.WithExecutor(ctx, chromedp.FromContext(ctx).Target)
	keyEvents, err := dispatchKeyEvent(ctx, sel, str, modifiers)
	if err != nil {
		return err
	}
	for _, k := range keyEvents {
		if err := k.Do(ctx); err != nil {
			return err
		}
	}
	return nil
}

func dispatchKeyEvent(ctx context.Context, sel NthChildSel, str string, modifiers uint16) ([]*input.DispatchKeyEventParams, error) {
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
		if err := clipboardAction(ctx, sel, str, keyDown.Modifiers); err != nil {
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

func clipboardAction(ctx context.Context, sel NthChildSel, str string, modifiers input.Modifier) error {
	str = strings.ToLower(str)
	isShift := modifiers & input.ModifierShift

	switch str {
	case "c":
		if err := ClipboardCommand(ctx, "copy"); err != nil {
			return err
		}
	// TODO
	// case "v":
	// 	if err := ClipboardCommand(ctx, "paste"); err != nil {
	// 		return err
	// 	}
	case "x":
		if err := ClipboardCommand(ctx, "cut"); err != nil {
			return err
		}
	case "a":
		if err := SelectNthElement(ctx, sel); err != nil {
			return err
		}
	case "z":
		if isShift > 0 {
			if err := ClipboardCommand(ctx, "redo"); err != nil {
				return err
			}
		} else {
			if err := ClipboardCommand(ctx, "undo"); err != nil {
				return err
			}
		}
	}
	return nil
}
