package chrometool

import (
	"context"
	"fmt"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/chromedp"
)

const (
	windowDimensionsJS  = `[window.innerWidth, window.innerHeight];`
	windowScrollJS      = `window.scrollBy(%d, %d);`
	clipboardInteractJS = `document.execCommand("%s")`
	selectNthElementJS  = `document.querySelectorAll("%s")[%d].select()`
)

// SendKeysToNthElement is an element query action that synthesizes the key up, char, and down
// events as needed for the runes in v, sending them to the nth element node
// matching the selector.
func SendKeysToNthElement(sel interface{}, n int, v string, opts ...chromedp.QueryOption) chromedp.QueryAction {
	return chromedp.QueryAfter(sel, func(ctx context.Context, nodes ...*cdp.Node) error {
		if len(nodes) < 1 {
			return fmt.Errorf("selector %q did not return any nodes", sel)
		}
		if len(nodes) < n {
			return fmt.Errorf("invalid n for selector %q", sel)
		}

		node := nodes[n]

		return KeyEventNode(node, v).Do(ctx)
	}, append(opts, chromedp.NodeReady)...)
}

// KeyEventNode is a key action that dispatches a key event on a element node.
func KeyEventNode(n *cdp.Node, keys string, opts ...chromedp.KeyOption) chromedp.KeyAction {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		return chromedp.KeyEvent(keys, opts...).Do(ctx)
	})
}

// ClickNthElement is an element query action that sends a mouse click event to the nth element
// node matching the selector.
func ClickNthElement(sel interface{}, n int, opts ...chromedp.QueryOption) chromedp.QueryAction {
	return chromedp.QueryAfter(sel, func(ctx context.Context, nodes ...*cdp.Node) error {
		if len(nodes) < 1 {
			return fmt.Errorf("selector %q did not return any nodes", sel)
		}
		if len(nodes) < n {
			return fmt.Errorf("invalid n for selector %q", sel)
		}
		return MouseClickNode(nodes[n]).Do(ctx)
	}, append(opts, chromedp.NodeReady)...)
}

// MouseClickNode is an action that dispatches a mouse left button click event
// at the center of a specified node.
//
// Note that the window will be scrolled if the node is not within the window's
// viewport.
func MouseClickNode(n *cdp.Node, opts ...chromedp.MouseOption) chromedp.MouseAction {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		var windowDim []float64
		if err := chromedp.Evaluate(windowDimensionsJS, &windowDim).Do(ctx); err != nil {
			return err
		}

		box, err := dom.GetBoxModel().WithNodeID(n.NodeID).Do(ctx)
		if err != nil {
			return err
		}

		x, y, err := EvalNodeCenter(box)
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

			x, y, err = EvalNodeCenter(box)
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

// EvalNodeCenter evaluates a node's central coordinates and returns x, y
func EvalNodeCenter(box *dom.BoxModel) (float64, float64, error) {
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

func ClipboardCommand(ctx context.Context, command string) error {
	var res interface{}
	if err := chromedp.Evaluate(fmt.Sprintf(clipboardInteractJS, command), &res).Do(ctx); err != nil {
		return err
	}
	return nil
}

func SelectNthElement(ctx context.Context, sel interface{}, n int) error {
	var res interface{}
	if err := chromedp.Evaluate(fmt.Sprintf(selectNthElementJS, sel, n), &res).Do(ctx); err != nil {
		if err.Error() == "encountered an undefined value" {
			err = nil
		}
		return err
	}
	return nil
}
