//go:build js && wasm

package component

import (
	"github.com/AureClai/goflow/vdom"
)

type Text struct {
	Content string
	Tag     string
	Class   string
}

func NewText(content string) *Text {
	return &Text{
		Content: content,
		Tag:     "span",
	}
}

func NewHeading(content string, level int) *Text {
	tag := "h1"
	switch level {
	case 1:
		tag = "h1"
	case 2:
		tag = "h2"
	case 3:
		tag = "h3"
	case 4:
		tag = "h4"
	case 5:
		tag = "h5"
	case 6:
		tag = "h6"
	}

	return &Text{
		Content: content,
		Tag:     tag,
	}
}

func NewParagraph(content string) *Text {
	return &Text{
		Content: content,
		Tag:     "p",
	}
}

func (t *Text) SetClass(class string) *Text {
	t.Class = class
	return t
}

func (t *Text) Render() *vdom.VNode {
	props := make(map[string]interface{})

	if t.Class != "" {
		props["class"] = t.Class
	}

	return &vdom.VNode{
		Type:     vdom.VNodeElement,
		Tag:      t.Tag,
		Props:    props,
		Children: []*vdom.VNode{{Type: vdom.VNodeText, Text: t.Content}},
	}
}
