//go:build js && wasm

package component

import (
	"github.com/AureClai/goflow/vdom"
)

type Container struct {
	Children []vdom.Component
	Class    string
	ID       string
	Style    string
}

func NewContainer() *Container {
	return &Container{
		Children: make([]vdom.Component, 0),
	}
}

func (c *Container) AddChild(child vdom.Component) *Container {
	c.Children = append(c.Children, child)
	return c
}

func (c *Container) SetClass(class string) *Container {
	c.Class = class
	return c
}

func (c *Container) SetID(id string) *Container {
	c.ID = id
	return c
}

func (c *Container) SetStyle(style string) *Container {
	c.Style = style
	return c
}

func (c *Container) Render() *vdom.VNode {
	props := make(map[string]interface{})

	if c.Class != "" {
		props["class"] = c.Class
	}

	if c.ID != "" {
		props["id"] = c.ID
	}

	if c.Style != "" {
		props["style"] = c.Style
	}

	children := make([]*vdom.VNode, len(c.Children))
	for i, child := range c.Children {
		children[i] = child.Render()
	}

	return &vdom.VNode{
		Type:     vdom.VNodeElement,
		Tag:      "div",
		Props:    props,
		Children: children,
	}
}
