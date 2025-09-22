//go:build js && wasm

package component

import (
	"fmt"

	"github.com/AureClai/goflow/renderer"
	"github.com/AureClai/goflow/vdom"
)

type Counter struct {
	count    int
	renderer *renderer.Renderer
}

func NewCounter(renderer *renderer.Renderer) *Counter {
	return &Counter{count: 0, renderer: renderer}
}

func (c *Counter) Render() *vdom.VNode {
	return &vdom.VNode{
		Type: vdom.VNodeElement,
		Tag:  "div",
		Children: []*vdom.VNode{
			{Type: vdom.VNodeText, Text: fmt.Sprintf("Count: %d", c.count)},
			{Type: vdom.VNodeElement,
				Tag: "button",
				Children: []*vdom.VNode{
					{Type: vdom.VNodeText, Text: "Increment"},
				},
				EventHandlers: map[string]func(){
					"click": c.increment,
				},
			},
		},
	}
}

func (c *Counter) increment() {
	c.count++
	c.renderer.Render(c.Render())
}
