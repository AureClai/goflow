//go:build js && wasm

package animation

import (
	"fmt"
	"math"
	"math/rand"
	"syscall/js"
	"time"
)

// TextMorphEffect creates a text morphing animation
type TextMorphEffect struct {
	engine *AnimationEngine
}

// NewTextMorphEffect creates a new text morph effect
func NewTextMorphEffect(engine *AnimationEngine) *TextMorphEffect {
	return &TextMorphEffect{engine: engine}
}

// MorphTextToShape morphs text characters into a shape with particles
func (tme *TextMorphEffect) MorphTextToShape(textElement js.Value, targetShape string, config ParticleConfig) {
	if !textElement.Truthy() {
		return
	}

	// Get text content and create individual character elements
	text := textElement.Get("textContent").String()
	textElement.Set("innerHTML", "")

	// Create character spans
	for i, char := range text {
		span := js.Global().Get("document").Call("createElement", "span")
		span.Set("textContent", string(char))
		span.Get("style").Set("display", "inline-block")
		span.Get("style").Set("transition", "all 0.3s ease")
		span.Set("data-char-index", i)
		textElement.Call("appendChild", span)
	}

	// Animate each character
	chars := textElement.Get("children")
	for i := 0; i < chars.Get("length").Int(); i++ {
		char := chars.Index(i)
		delay := time.Duration(i*100) * time.Millisecond

		// Create explosion effect for each character
		tme.animateCharacterExplosion(char, delay, config)
	}
}

// animateCharacterExplosion creates an explosion effect for a character
func (tme *TextMorphEffect) animateCharacterExplosion(char js.Value, delay time.Duration, config ParticleConfig) {
	// Get character position
	rect := char.Call("getBoundingClientRect")
	x := rect.Get("left").Float() + rect.Get("width").Float()/2
	y := rect.Get("top").Float() + rect.Get("height").Float()/2

	// Character explosion animation
	explodeAnim := NewAnimation().
		SetElement(char).
		SetDuration(600*time.Millisecond).
		SetDelay(delay).
		SetEasing(EaseOutExpo).
		Animate("transform", "translateY(0px) scale(1) rotate(0deg)",
			fmt.Sprintf("translateY(-40px) scale(1.8) rotate(%ddeg)", rand.Intn(360)), "").
		Animate("opacity", 1.0, 0.0, "").
		OnStart(func() {
			// Create particle burst at character position
			particleSystem := NewParticleSystem(tme.engine)
			particleSystem.CreateParticleBurst(x, y, config)
		}).
		Build()

	tme.engine.AddAnimation(explodeAnim)
}

// GlowPulseEffect creates a pulsing glow effect
type GlowPulseEffect struct {
	engine *AnimationEngine
}

// NewGlowPulseEffect creates a new glow pulse effect
func NewGlowPulseEffect(engine *AnimationEngine) *GlowPulseEffect {
	return &GlowPulseEffect{engine: engine}
}

// StartGlowPulse starts a pulsing glow animation
func (gpe *GlowPulseEffect) StartGlowPulse(element js.Value, color Color, intensity float64, duration time.Duration) {
	if !element.Truthy() {
		return
	}

	timeline := NewTimeline(gpe.engine)

	// Create pulsing glow effect
	for i := 0; i < 3; i++ {
		glowIn := NewAnimation().
			SetElement(element).
			SetDuration(duration/2).
			SetEasing(EaseInOutSine).
			Animate("boxShadow", "0 0 0px rgba(0,0,0,0)",
				fmt.Sprintf("0 0 %.0fpx rgba(%d,%d,%d,%.2f), 0 0 %.0fpx rgba(%d,%d,%d,%.2f)",
					intensity, int(color.R), int(color.G), int(color.B), color.A,
					intensity*2, int(color.R), int(color.G), int(color.B), color.A*0.5), "").
			Build()

		glowOut := NewAnimation().
			SetElement(element).
			SetDuration(duration/2).
			SetEasing(EaseInOutSine).
			Animate("boxShadow",
				fmt.Sprintf("0 0 %.0fpx rgba(%d,%d,%d,%.2f), 0 0 %.0fpx rgba(%d,%d,%d,%.2f)",
					intensity, int(color.R), int(color.G), int(color.B), color.A,
					intensity*2, int(color.R), int(color.G), int(color.B), color.A*0.5),
				"0 0 0px rgba(0,0,0,0)", "").
			Build()

		timeline.AddAnimation(time.Duration(i)*duration, glowIn)
		timeline.AddAnimation(time.Duration(i)*duration+duration/2, glowOut)
	}

	timeline.Play()
}

// RippleEffect creates a ripple animation
type RippleEffect struct {
	engine *AnimationEngine
}

// NewRippleEffect creates a new ripple effect
func NewRippleEffect(engine *AnimationEngine) *RippleEffect {
	return &RippleEffect{engine: engine}
}

// CreateRipple creates a ripple effect at the specified position
func (re *RippleEffect) CreateRipple(x, y float64, color Color, maxRadius float64, duration time.Duration) {
	document := js.Global().Get("document")

	// Create ripple element
	ripple := document.Call("createElement", "div")
	ripple.Get("style").Set("position", "fixed")
	ripple.Get("style").Set("left", fmt.Sprintf("%.2fpx", x))
	ripple.Get("style").Set("top", fmt.Sprintf("%.2fpx", y))
	ripple.Get("style").Set("width", "0px")
	ripple.Get("style").Set("height", "0px")
	ripple.Get("style").Set("borderRadius", "50%")
	ripple.Get("style").Set("border", fmt.Sprintf("2px solid %s", color.ToString()))
	ripple.Get("style").Set("pointerEvents", "none")
	ripple.Get("style").Set("zIndex", "10000")
	ripple.Get("style").Set("transform", "translate(-50%, -50%)")

	document.Get("body").Call("appendChild", ripple)

	// Animate ripple expansion
	expandAnim := NewAnimation().
		SetElement(ripple).
		SetDuration(duration).
		SetEasing(EaseOutQuad).
		Animate("width", 0.0, maxRadius*2, "px").
		Animate("height", 0.0, maxRadius*2, "px").
		Animate("opacity", color.A, 0.0, "").
		OnComplete(func() {
			document.Get("body").Call("removeChild", ripple)
		}).
		Build()

	re.engine.AddAnimation(expandAnim)
}

// MagneticEffect creates a magnetic attraction effect
type MagneticEffect struct {
	engine *AnimationEngine
}

// NewMagneticEffect creates a new magnetic effect
func NewMagneticEffect(engine *AnimationEngine) *MagneticEffect {
	return &MagneticEffect{engine: engine}
}

// AttractElements attracts elements to a central point
func (me *MagneticEffect) AttractElements(elements []js.Value, centerX, centerY float64, strength float64, duration time.Duration) {
	for i, element := range elements {
		if !element.Truthy() {
			continue
		}

		rect := element.Call("getBoundingClientRect")
		currentX := rect.Get("left").Float() + rect.Get("width").Float()/2
		currentY := rect.Get("top").Float() + rect.Get("height").Float()/2

		// Calculate attraction vector
		dx := centerX - currentX
		dy := centerY - currentY
		distance := math.Sqrt(dx*dx + dy*dy)

		if distance > 0 {
			// Normalize and apply strength
			attractX := (dx / distance) * strength
			attractY := (dy / distance) * strength

			delay := time.Duration(i*50) * time.Millisecond

			attractAnim := NewAnimation().
				SetElement(element).
				SetDuration(duration).
				SetDelay(delay).
				SetEasing(EaseOutElastic).
				Animate("transform", "translate(0px, 0px)",
					fmt.Sprintf("translate(%.2fpx, %.2fpx)", attractX, attractY), "").
				Build()

			me.engine.AddAnimation(attractAnim)
		}
	}
}

// FloatingEffect creates floating animations
type FloatingEffect struct {
	engine *AnimationEngine
}

// NewFloatingEffect creates a new floating effect
func NewFloatingEffect(engine *AnimationEngine) *FloatingEffect {
	return &FloatingEffect{engine: engine}
}

// StartFloating starts a continuous floating animation
func (fe *FloatingEffect) StartFloating(element js.Value, amplitude float64, period time.Duration) {
	if !element.Truthy() {
		return
	}

	// Create infinite floating animation
	floatUp := NewAnimation().
		SetElement(element).
		SetDuration(period/2).
		SetEasing(EaseInOutSine).
		Animate("transform", "translateY(0px)", fmt.Sprintf("translateY(-%.2fpx)", amplitude), "").
		Build()

	floatDown := NewAnimation().
		SetElement(element).
		SetDuration(period/2).
		SetEasing(EaseInOutSine).
		Animate("transform", fmt.Sprintf("translateY(-%.2fpx)", amplitude), "translateY(0px)", "").
		Build()

	timeline := NewTimeline(fe.engine)
	timeline.AddAnimation(0, floatUp)
	timeline.AddAnimation(period/2, floatDown)
	timeline.SetLoop(true)
	timeline.Play()
}

// ParticleTrailEffect creates particle trails following mouse movement
type ParticleTrailEffect struct {
	engine         *AnimationEngine
	particleSystem *ParticleSystem
	isActive       bool
}

// NewParticleTrailEffect creates a new particle trail effect
func NewParticleTrailEffect(engine *AnimationEngine, particleSystem *ParticleSystem) *ParticleTrailEffect {
	return &ParticleTrailEffect{
		engine:         engine,
		particleSystem: particleSystem,
		isActive:       false,
	}
}

// StartTrail starts the particle trail effect
func (pte *ParticleTrailEffect) StartTrail(config ParticleConfig) {
	if pte.isActive {
		return
	}

	pte.isActive = true
	document := js.Global().Get("document")

	mouseMoveHandler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if !pte.isActive {
			return nil
		}

		if len(args) > 0 {
			event := args[0]
			x := event.Get("clientX").Float()
			y := event.Get("clientY").Float()

			// Create small particle burst at mouse position
			trailConfig := config
			trailConfig.Count = 3
			trailConfig.LifeTime = 0.8
			trailConfig.MinSpeed = 10
			trailConfig.MaxSpeed = 30

			pte.particleSystem.CreateParticleBurst(x, y, trailConfig)
		}
		return nil
	})

	document.Call("addEventListener", "mousemove", mouseMoveHandler)
}

// StopTrail stops the particle trail effect
func (pte *ParticleTrailEffect) StopTrail() {
	pte.isActive = false
}
