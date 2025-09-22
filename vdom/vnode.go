package vdom

type VNodeType int

const (
	VNodeElement VNodeType = iota
	VNodeText
)

// VNode represents a virtual node in the DOM
type VNode struct {
	Type          VNodeType
	Tag           string                 // HTML tag name
	Text          string                 // Text content
	Props         map[string]interface{} // Attributes and properties
	Children      []*VNode               // Child nodes
	EventHandlers map[string]func()      // Event handlers
}

type Component interface {
	Render() *VNode
}
