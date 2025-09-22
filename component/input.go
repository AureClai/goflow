//go:build js && wasm

package component

import (
	"syscall/js"

	"github.com/AureClai/goflow/vdom"
)

type Input struct {
	Value       string
	Placeholder string
	Type        string
	OnChange    func(string)
	OnInput     func(string)
	Class       string
}

func NewInput(placeholder string) *Input {
	return &Input{
		Placeholder: placeholder,
		Type:        "text",
	}
}

func (i *Input) SetValue(value string) *Input {
	i.Value = value
	return i
}

func (i *Input) SetType(inputType string) *Input {
	i.Type = inputType
	return i
}

func (i *Input) SetClass(class string) *Input {
	i.Class = class
	return i
}

func (i *Input) OnChangeHandler(handler func(string)) *Input {
	i.OnChange = handler
	return i
}

func (i *Input) OnInputHandler(handler func(string)) *Input {
	i.OnInput = handler
	return i
}

func (i *Input) Render() *vdom.VNode {
	props := make(map[string]interface{})
	props["type"] = i.Type
	props["value"] = i.Value

	if i.Placeholder != "" {
		props["placeholder"] = i.Placeholder
	}

	if i.Class != "" {
		props["class"] = i.Class
	}

	eventHandlers := make(map[string]func())

	if i.OnChange != nil {
		eventHandlers["change"] = func() {
			event := js.Global().Get("event")
			target := event.Get("target")
			value := target.Get("value").String()
			i.Value = value
			i.OnChange(value)
		}
	}

	if i.OnInput != nil {
		eventHandlers["input"] = func() {
			event := js.Global().Get("event")
			target := event.Get("target")
			value := target.Get("value").String()
			i.Value = value
			i.OnInput(value)
		}
	}

	return &vdom.VNode{
		Type:          vdom.VNodeElement,
		Tag:           "input",
		Props:         props,
		EventHandlers: eventHandlers,
	}
}
