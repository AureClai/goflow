//go:build js && wasm

package renderer

import (
	"syscall/js"

	"github.com/AureClai/goflow/vdom"
)

type Renderer struct {
	container      js.Value
	animationFrame js.Value
}

func NewRenderer(containerID string) *Renderer {
	document := js.Global().Get("document")
	container := document.Call("getElementById", containerID)
	return &Renderer{
		container: container,
	}
}

func (r *Renderer) Render(vnode *vdom.VNode) {
	// Clear container and render new tree
	r.container.Set("innerHTML", "")
	if vnode != nil {
		domNode := r.createDomNode(vnode)
		r.container.Call("appendChild", domNode)
	}
}

func (r *Renderer) createDomNode(vnode *vdom.VNode) js.Value {
	if vnode == nil {
		return js.Null()
	}

	document := js.Global().Get("document")

	switch vnode.Type {
	case vdom.VNodeText:
		return document.Call("createTextNode", vnode.Text)

	case vdom.VNodeElement:
		element := document.Call("createElement", vnode.Tag)

		// Set properties
		for key, value := range vnode.Props {
			if key == "style" {
				// Handle inline styles
				element.Get("style").Set("cssText", value)
			} else {
				element.Call("setAttribute", key, value)
			}
		}

		// Add event listeners
		for event, handler := range vnode.EventHandlers {
			element.Call("addEventListener", event, js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				handler()
				return nil
			}))
		}

		// Append children
		for _, child := range vnode.Children {
			childNode := r.createDomNode(child)
			if childNode.Truthy() {
				element.Call("appendChild", childNode)
			}
		}

		return element
	}

	return js.Null()
}

// RequestFrame requests an animation frame for smooth rendering
func (r *Renderer) RequestFrame() {
	if r.animationFrame.Truthy() {
		js.Global().Call("cancelAnimationFrame", r.animationFrame)
	}
	
	r.animationFrame = js.Global().Call("requestAnimationFrame", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		// Frame callback - can be used for post-render operations
		return nil
	}))
}

// GetContainer returns the container element
func (r *Renderer) GetContainer() js.Value {
	return r.container
}
