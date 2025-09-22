//go:build js && wasm

package animation

import (
	"fmt"
	"sync"
	"syscall/js"
	"time"

	"github.com/AureClai/goflow/renderer"
)

// EasingFunc defines the signature for easing functions
type EasingFunc func(float64) float64

// AnimationState represents the current state of an animation
type AnimationState int

const (
	AnimationPending AnimationState = iota
	AnimationRunning
	AnimationPaused
	AnimationComplete
	AnimationCancelled
)

// AnimationEngine manages all active animations
type AnimationEngine struct {
	activeAnimations map[string]*Animation
	timeline         *Timeline
	renderer         *renderer.Renderer
	running          bool
	mutex            sync.RWMutex
	frameCallbacks   []func()
}

// Animation represents a single animation instance
type Animation struct {
	ID         string
	StartTime  time.Time
	Duration   time.Duration
	Delay      time.Duration
	Easing     EasingFunc
	Properties []PropertyAnimation
	State      AnimationState
	Progress   float64
	OnStart    func()
	OnUpdate   func(progress float64)
	OnComplete func()
	OnCancel   func()
	element    js.Value
}

// PropertyAnimation represents an animated property
type PropertyAnimation struct {
	Property    string
	From        interface{}
	To          interface{}
	Current     interface{}
	Unit        string // e.g., "px", "%", "deg"
	Interpolate func(from, to interface{}, progress float64) interface{}
}

// NewAnimationEngine creates a new animation engine
func NewAnimationEngine(r *renderer.Renderer) *AnimationEngine {
	return &AnimationEngine{
		activeAnimations: make(map[string]*Animation),
		renderer:         r,
		frameCallbacks:   make([]func(), 0),
	}
}

// Start begins the animation loop
func (e *AnimationEngine) Start() {
	e.mutex.Lock()
	if e.running {
		e.mutex.Unlock()
		return
	}
	e.running = true
	e.mutex.Unlock()

	// Use requestAnimationFrame for smooth 60fps animations
	e.scheduleFrame()
}

// Stop stops the animation loop
func (e *AnimationEngine) Stop() {
	e.mutex.Lock()
	e.running = false
	e.mutex.Unlock()
}

// scheduleFrame uses requestAnimationFrame for optimal performance
func (e *AnimationEngine) scheduleFrame() {
	if !e.running {
		return
	}

	js.Global().Call("requestAnimationFrame", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		e.updateFrame()
		e.scheduleFrame() // Schedule next frame
		return nil
	}))
}

// updateFrame processes all active animations
func (e *AnimationEngine) updateFrame() {
	now := time.Now()

	e.mutex.Lock()
	defer e.mutex.Unlock()

	// Execute frame callbacks
	for _, callback := range e.frameCallbacks {
		callback()
	}
	e.frameCallbacks = e.frameCallbacks[:0]

	// Update all active animations
	for id, anim := range e.activeAnimations {
		if e.updateAnimation(anim, now) {
			delete(e.activeAnimations, id)
		}
	}
}

// updateAnimation updates a single animation and returns true if complete
func (e *AnimationEngine) updateAnimation(anim *Animation, now time.Time) bool {
	// Handle delay
	if now.Before(anim.StartTime.Add(anim.Delay)) {
		return false
	}

	// Start animation if pending
	if anim.State == AnimationPending {
		anim.State = AnimationRunning
		if anim.OnStart != nil {
			anim.OnStart()
		}
	}

	// Skip if not running
	if anim.State != AnimationRunning {
		return anim.State == AnimationComplete || anim.State == AnimationCancelled
	}

	// Calculate progress
	elapsed := now.Sub(anim.StartTime.Add(anim.Delay))
	progress := float64(elapsed) / float64(anim.Duration)

	if progress >= 1.0 {
		// Animation complete
		progress = 1.0
		anim.State = AnimationComplete
		anim.Progress = progress

		// Update properties to final values
		e.updateAnimationProperties(anim, progress)

		if anim.OnComplete != nil {
			anim.OnComplete()
		}
		return true
	}

	// Apply easing
	if anim.Easing != nil {
		progress = anim.Easing(progress)
	}

	anim.Progress = progress

	// Update properties
	e.updateAnimationProperties(anim, progress)

	if anim.OnUpdate != nil {
		anim.OnUpdate(progress)
	}

	return false
}

// updateAnimationProperties updates all animated properties
func (e *AnimationEngine) updateAnimationProperties(anim *Animation, progress float64) {
	for i := range anim.Properties {
		prop := &anim.Properties[i]

		if prop.Interpolate != nil {
			prop.Current = prop.Interpolate(prop.From, prop.To, progress)
		} else {
			prop.Current = e.defaultInterpolate(prop.From, prop.To, progress)
		}

		// Apply to DOM element if available
		if anim.element.Truthy() {
			e.applyPropertyToElement(anim.element, prop)
		}
	}
}

// defaultInterpolate provides default interpolation for common types
func (e *AnimationEngine) defaultInterpolate(from, to interface{}, progress float64) interface{} {
	switch f := from.(type) {
	case float64:
		if t, ok := to.(float64); ok {
			return f + (t-f)*progress
		}
	case int:
		if t, ok := to.(int); ok {
			return int(float64(f) + float64(t-f)*progress)
		}
	case string:
		// Handle color interpolation, etc.
		return e.interpolateString(f, to.(string), progress)
	}

	// Fallback: return target value when progress >= 0.5
	if progress >= 0.5 {
		return to
	}
	return from
}

// interpolateString handles string interpolation (colors, etc.)
func (e *AnimationEngine) interpolateString(from, to string, progress float64) string {
	// For now, simple threshold-based switching
	// TODO: Implement proper color interpolation
	if progress >= 0.5 {
		return to
	}
	return from
}

// applyPropertyToElement applies an animated property to a DOM element
func (e *AnimationEngine) applyPropertyToElement(element js.Value, prop *PropertyAnimation) {
	style := element.Get("style")

	switch prop.Property {
	case "opacity":
		style.Set("opacity", fmt.Sprintf("%v", prop.Current))
	case "transform":
		style.Set("transform", fmt.Sprintf("%v", prop.Current))
	case "left", "top", "width", "height":
		value := fmt.Sprintf("%v%s", prop.Current, prop.Unit)
		style.Set(prop.Property, value)
	case "backgroundColor", "color", "borderColor":
		style.Set(prop.Property, fmt.Sprintf("%v", prop.Current))
	default:
		// Generic property setting
		if prop.Unit != "" {
			value := fmt.Sprintf("%v%s", prop.Current, prop.Unit)
			style.Set(prop.Property, value)
		} else {
			style.Set(prop.Property, fmt.Sprintf("%v", prop.Current))
		}
	}
}

// AddAnimation adds a new animation to the engine
func (e *AnimationEngine) AddAnimation(anim *Animation) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if anim.ID == "" {
		anim.ID = fmt.Sprintf("anim_%d", time.Now().UnixNano())
	}

	if anim.StartTime.IsZero() {
		anim.StartTime = time.Now()
	}

	if anim.State == AnimationState(0) {
		anim.State = AnimationPending
	}

	e.activeAnimations[anim.ID] = anim
}

// RemoveAnimation removes an animation by ID
func (e *AnimationEngine) RemoveAnimation(id string) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if anim, exists := e.activeAnimations[id]; exists {
		anim.State = AnimationCancelled
		if anim.OnCancel != nil {
			anim.OnCancel()
		}
		delete(e.activeAnimations, id)
	}
}

// PauseAnimation pauses an animation by ID
func (e *AnimationEngine) PauseAnimation(id string) {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	if anim, exists := e.activeAnimations[id]; exists {
		anim.State = AnimationPaused
	}
}

// ResumeAnimation resumes a paused animation by ID
func (e *AnimationEngine) ResumeAnimation(id string) {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	if anim, exists := e.activeAnimations[id]; exists && anim.State == AnimationPaused {
		anim.State = AnimationRunning
	}
}

// GetActiveAnimations returns a copy of active animation IDs
func (e *AnimationEngine) GetActiveAnimations() []string {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	ids := make([]string, 0, len(e.activeAnimations))
	for id := range e.activeAnimations {
		ids = append(ids, id)
	}
	return ids
}

// OnNextFrame schedules a callback for the next animation frame
func (e *AnimationEngine) OnNextFrame(callback func()) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	e.frameCallbacks = append(e.frameCallbacks, callback)
}
