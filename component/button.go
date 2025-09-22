//go:build js && wasm

package component

import (
	"github.com/AureClai/goflow/vdom"
)

type Button struct {
	Text     string
	OnClick  func()
	Class    string
	Disabled bool
}

func NewButton(text string, onClick func()) *Button {
	return &Button{
		Text:    text,
		OnClick: onClick,
	}
}

func (b *Button) SetClass(class string) *Button {
	b.Class = class
	return b
}

func (b *Button) SetDisabled(disabled bool) *Button {
	b.Disabled = disabled
	return b
}

func (b *Button) Render() *vdom.VNode {
	props := make(map[string]interface{})

	if b.Class != "" {
		props["class"] = b.Class
	}

	if b.Disabled {
		props["disabled"] = true
	}

	eventHandlers := make(map[string]func())
	if b.OnClick != nil && !b.Disabled {
		eventHandlers["click"] = b.OnClick
	}

	return &vdom.VNode{
		Type:          vdom.VNodeElement,
		Tag:           "button",
		Props:         props,
		Children:      []*vdom.VNode{{Type: vdom.VNodeText, Text: b.Text}},
		EventHandlers: eventHandlers,
	}
}
