package chrometool

import (
	"context"
	"fmt"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/chromedp"
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
		dom.ScrollIntoViewIfNeeded().WithNodeID(n.NodeID).Do(ctx)

		boxes, err := dom.GetContentQuads().WithNodeID(n.NodeID).Do(ctx)
		if err != nil {
			return err
		}
		content := boxes[0]

		c := len(content)
		if c%2 != 0 || c < 1 {
			return chromedp.ErrInvalidDimensions
		}

		var x, y float64
		for i := 0; i < c; i += 2 {
			x += content[i]
			y += content[i+1]
		}
		x /= float64(c / 2)
		y /= float64(c / 2)

		return chromedp.MouseClickXY(x, y, opts...).Do(ctx)
	})
}
