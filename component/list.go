//go:build js && wasm

package component

import (
	"github.com/AureClai/goflow/vdom"
)

type List struct {
	Items       []string
	Ordered     bool
	Class       string
	OnItemClick func(int, string)
}

func NewList(items []string) *List {
	return &List{
		Items:   items,
		Ordered: false,
	}
}

func NewOrderedList(items []string) *List {
	return &List{
		Items:   items,
		Ordered: true,
	}
}

func (l *List) SetClass(class string) *List {
	l.Class = class
	return l
}

func (l *List) OnItemClickHandler(handler func(int, string)) *List {
	l.OnItemClick = handler
	return l
}

func (l *List) Render() *vdom.VNode {
	tag := "ul"
	if l.Ordered {
		tag = "ol"
	}

	props := make(map[string]interface{})
	if l.Class != "" {
		props["class"] = l.Class
	}

	children := make([]*vdom.VNode, len(l.Items))
	for i, item := range l.Items {
		eventHandlers := make(map[string]func())
		if l.OnItemClick != nil {
			index := i
			text := item
			eventHandlers["click"] = func() {
				l.OnItemClick(index, text)
			}
		}

		children[i] = &vdom.VNode{
			Type:          vdom.VNodeElement,
			Tag:           "li",
			Children:      []*vdom.VNode{{Type: vdom.VNodeText, Text: item}},
			EventHandlers: eventHandlers,
		}
	}

	return &vdom.VNode{
		Type:     vdom.VNodeElement,
		Tag:      tag,
		Props:    props,
		Children: children,
	}
}
