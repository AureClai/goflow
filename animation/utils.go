//go:build js && wasm

package animation

import (
	"fmt"
	"math"
	"syscall/js"
	"time"
)

// AnimationBuilder provides a fluent interface for creating animations
type AnimationBuilder struct {
	animation *Animation
}

// NewAnimation creates a new animation builder
func NewAnimation() *AnimationBuilder {
	return &AnimationBuilder{
		animation: &Animation{
			Duration:   500 * time.Millisecond,
			Easing:     EaseOut,
			Properties: make([]PropertyAnimation, 0),
			State:      AnimationPending,
		},
	}
}

// SetID sets the animation ID
func (ab *AnimationBuilder) SetID(id string) *AnimationBuilder {
	ab.animation.ID = id
	return ab
}

// SetDuration sets the animation duration
func (ab *AnimationBuilder) SetDuration(duration time.Duration) *AnimationBuilder {
	ab.animation.Duration = duration
	return ab
}

// SetDelay sets the animation delay
func (ab *AnimationBuilder) SetDelay(delay time.Duration) *AnimationBuilder {
	ab.animation.Delay = delay
	return ab
}

// SetEasing sets the easing function
func (ab *AnimationBuilder) SetEasing(easing EasingFunc) *AnimationBuilder {
	ab.animation.Easing = easing
	return ab
}

// SetElement sets the target DOM element
func (ab *AnimationBuilder) SetElement(element js.Value) *AnimationBuilder {
	ab.animation.element = element
	return ab
}

// OnStart sets the onStart callback
func (ab *AnimationBuilder) OnStart(callback func()) *AnimationBuilder {
	ab.animation.OnStart = callback
	return ab
}

// OnUpdate sets the onUpdate callback
func (ab *AnimationBuilder) OnUpdate(callback func(progress float64)) *AnimationBuilder {
	ab.animation.OnUpdate = callback
	return ab
}

// OnComplete sets the onComplete callback
func (ab *AnimationBuilder) OnComplete(callback func()) *AnimationBuilder {
	ab.animation.OnComplete = callback
	return ab
}

// OnCancel sets the onCancel callback
func (ab *AnimationBuilder) OnCancel(callback func()) *AnimationBuilder {
	ab.animation.OnCancel = callback
	return ab
}

// Animate adds a property animation
func (ab *AnimationBuilder) Animate(property string, from, to interface{}, unit string) *AnimationBuilder {
	ab.animation.Properties = append(ab.animation.Properties, PropertyAnimation{
		Property: property,
		From:     from,
		To:       to,
		Unit:     unit,
	})
	return ab
}

// FadeTo animates opacity
func (ab *AnimationBuilder) FadeTo(opacity float64) *AnimationBuilder {
	return ab.Animate("opacity", 0.0, opacity, "")
}

// FadeIn animates opacity from 0 to 1
func (ab *AnimationBuilder) FadeIn() *AnimationBuilder {
	return ab.FadeTo(1.0)
}

// FadeOut animates opacity from current to 0
func (ab *AnimationBuilder) FadeOut() *AnimationBuilder {
	return ab.Animate("opacity", 1.0, 0.0, "")
}

// MoveTo animates position
func (ab *AnimationBuilder) MoveTo(x, y float64) *AnimationBuilder {
	ab.Animate("left", 0.0, x, "px")
	ab.Animate("top", 0.0, y, "px")
	return ab
}

// ScaleTo animates scale
func (ab *AnimationBuilder) ScaleTo(scale float64) *AnimationBuilder {
	return ab.Animate("transform", "scale(1)", fmt.Sprintf("scale(%f)", scale), "")
}

// RotateTo animates rotation
func (ab *AnimationBuilder) RotateTo(degrees float64) *AnimationBuilder {
	return ab.Animate("transform", "rotate(0deg)", fmt.Sprintf("rotate(%fdeg)", degrees), "")
}

// SlideLeft slides element to the left
func (ab *AnimationBuilder) SlideLeft(distance float64) *AnimationBuilder {
	return ab.Animate("transform", "translateX(0px)", fmt.Sprintf("translateX(-%fpx)", distance), "")
}

// SlideRight slides element to the right
func (ab *AnimationBuilder) SlideRight(distance float64) *AnimationBuilder {
	return ab.Animate("transform", "translateX(0px)", fmt.Sprintf("translateX(%fpx)", distance), "")
}

// SlideUp slides element up
func (ab *AnimationBuilder) SlideUp(distance float64) *AnimationBuilder {
	return ab.Animate("transform", "translateY(0px)", fmt.Sprintf("translateY(-%fpx)", distance), "")
}

// SlideDown slides element down
func (ab *AnimationBuilder) SlideDown(distance float64) *AnimationBuilder {
	return ab.Animate("transform", "translateY(0px)", fmt.Sprintf("translateY(%fpx)", distance), "")
}

// Build returns the constructed animation
func (ab *AnimationBuilder) Build() *Animation {
	return ab.animation
}

// Color utilities for color interpolation
type Color struct {
	R, G, B, A float64
}

// ParseColor parses a color string (hex, rgb, rgba)
func ParseColor(colorStr string) Color {
	// Simple implementation - in a real system, you'd want more robust parsing
	if len(colorStr) == 7 && colorStr[0] == '#' {
		// Hex color
		r := hexToFloat(colorStr[1:3])
		g := hexToFloat(colorStr[3:5])
		b := hexToFloat(colorStr[5:7])
		return Color{R: r, G: g, B: b, A: 1.0}
	}

	// Default to white
	return Color{R: 255, G: 255, B: 255, A: 1.0}
}

// hexToFloat converts hex string to float64
func hexToFloat(hex string) float64 {
	if len(hex) != 2 {
		return 0
	}

	val := 0
	for _, char := range hex {
		val *= 16
		if char >= '0' && char <= '9' {
			val += int(char - '0')
		} else if char >= 'a' && char <= 'f' {
			val += int(char - 'a' + 10)
		} else if char >= 'A' && char <= 'F' {
			val += int(char - 'A' + 10)
		}
	}

	return float64(val)
}

// ToString converts color to string
func (c Color) ToString() string {
	if c.A == 1.0 {
		return fmt.Sprintf("rgb(%d,%d,%d)", int(c.R), int(c.G), int(c.B))
	}
	return fmt.Sprintf("rgba(%d,%d,%d,%f)", int(c.R), int(c.G), int(c.B), c.A)
}

// Lerp interpolates between two colors
func (c Color) Lerp(target Color, t float64) Color {
	return Color{
		R: c.R + (target.R-c.R)*t,
		G: c.G + (target.G-c.G)*t,
		B: c.B + (target.B-c.B)*t,
		A: c.A + (target.A-c.A)*t,
	}
}

// Vector2 represents a 2D vector for position/movement animations
type Vector2 struct {
	X, Y float64
}

// Lerp interpolates between two vectors
func (v Vector2) Lerp(target Vector2, t float64) Vector2 {
	return Vector2{
		X: v.X + (target.X-v.X)*t,
		Y: v.Y + (target.Y-v.Y)*t,
	}
}

// Distance calculates distance between two vectors
func (v Vector2) Distance(other Vector2) float64 {
	dx := v.X - other.X
	dy := v.Y - other.Y
	return math.Sqrt(dx*dx + dy*dy)
}

// Normalize returns a normalized vector
func (v Vector2) Normalize() Vector2 {
	length := math.Sqrt(v.X*v.X + v.Y*v.Y)
	if length == 0 {
		return Vector2{0, 0}
	}
	return Vector2{X: v.X / length, Y: v.Y / length}
}

// Spring physics for natural animations
type Spring struct {
	Position     float64
	Velocity     float64
	Target       float64
	Stiffness    float64
	Damping      float64
	Mass         float64
	RestDistance float64
}

// NewSpring creates a new spring with default values
func NewSpring(target float64) *Spring {
	return &Spring{
		Position:     0,
		Velocity:     0,
		Target:       target,
		Stiffness:    100,
		Damping:      10,
		Mass:         1,
		RestDistance: 0.01,
	}
}

// Update updates the spring physics
func (s *Spring) Update(deltaTime float64) {
	force := s.Stiffness * (s.Target - s.Position)
	damping := s.Damping * s.Velocity
	acceleration := (force - damping) / s.Mass

	s.Velocity += acceleration * deltaTime
	s.Position += s.Velocity * deltaTime
}

// IsAtRest returns true if the spring is at rest
func (s *Spring) IsAtRest() bool {
	return math.Abs(s.Position-s.Target) < s.RestDistance && math.Abs(s.Velocity) < s.RestDistance
}

// Utility functions for common animation patterns

// Pulse creates a pulsing scale animation
func Pulse(engine *AnimationEngine, element js.Value, scale float64, duration time.Duration) *Animation {
	return NewAnimation().
		SetElement(element).
		SetDuration(duration).
		SetEasing(EaseInOutSine).
		ScaleTo(scale).
		OnComplete(func() {
			// Scale back down
			NewAnimation().
				SetElement(element).
				SetDuration(duration).
				SetEasing(EaseInOutSine).
				ScaleTo(1.0).
				Build()
		}).
		Build()
}

// Shake creates a shake animation
func Shake(engine *AnimationEngine, element js.Value, intensity float64, duration time.Duration) *Timeline {
	timeline := NewTimeline(engine)

	// Create multiple quick movements
	steps := 8
	stepDuration := duration / time.Duration(steps)

	for i := 0; i < steps; i++ {
		direction := 1.0
		if i%2 == 0 {
			direction = -1.0
		}

		offset := intensity * direction * (1.0 - float64(i)/float64(steps))

		anim := NewAnimation().
			SetElement(element).
			SetDuration(stepDuration).
			SetEasing(EaseInOutQuad).
			SlideRight(offset).
			Build()

		timeline.AddAnimation(time.Duration(i)*stepDuration, anim)
	}

	return timeline
}

// Bounce creates a bouncing animation
func Bounce(engine *AnimationEngine, element js.Value, height float64, duration time.Duration) *Timeline {
	timeline := NewTimeline(engine)

	// Up
	upAnim := NewAnimation().
		SetElement(element).
		SetDuration(duration / 2).
		SetEasing(EaseOutQuad).
		SlideUp(height).
		Build()

	// Down
	downAnim := NewAnimation().
		SetElement(element).
		SetDuration(duration / 2).
		SetEasing(EaseInQuad).
		SlideDown(height).
		Build()

	timeline.AddAnimation(0, upAnim)
	timeline.AddAnimation(duration/2, downAnim)

	return timeline
}

// Wiggle creates a wiggling rotation animation
func Wiggle(engine *AnimationEngine, element js.Value, angle float64, duration time.Duration) *Timeline {
	timeline := NewTimeline(engine)

	steps := 6
	stepDuration := duration / time.Duration(steps)

	for i := 0; i < steps; i++ {
		direction := 1.0
		if i%2 == 0 {
			direction = -1.0
		}

		rotation := angle * direction * (1.0 - float64(i)/float64(steps*2))

		anim := NewAnimation().
			SetElement(element).
			SetDuration(stepDuration).
			SetEasing(EaseInOutSine).
			RotateTo(rotation).
			Build()

		timeline.AddAnimation(time.Duration(i)*stepDuration, anim)
	}

	return timeline
}

// TypeWriter creates a typewriter text animation
func TypeWriter(engine *AnimationEngine, element js.Value, text string, charDelay time.Duration) *Timeline {
	timeline := NewTimeline(engine)

	for i := 0; i <= len(text); i++ {
		currentText := text[:i]

		anim := &Animation{
			ID:       fmt.Sprintf("typewriter_%d_%d", time.Now().UnixNano(), i),
			Duration: 1 * time.Millisecond, // Instant
			OnStart: func() {
				element.Set("textContent", currentText)
			},
		}

		timeline.AddAnimation(time.Duration(i)*charDelay, anim)
	}

	return timeline
}
