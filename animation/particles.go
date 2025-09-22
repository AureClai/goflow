//go:build js && wasm

package animation

import (
	"fmt"
	"math"
	"math/rand"
	"syscall/js"
	"time"
)

// Particle represents a single particle in the system
type Particle struct {
	ID       string
	Element  js.Value
	Position Vector2
	Velocity Vector2
	Size     float64
	Color    Color
	Life     float64 // 0.0 to 1.0
	MaxLife  float64 // in seconds
	Gravity  float64
	Fade     bool
	Scale    bool
}

// ParticleSystem manages a collection of particles
type ParticleSystem struct {
	particles []*Particle
	engine    *AnimationEngine
	container js.Value
}

// NewParticleSystem creates a new particle system
func NewParticleSystem(engine *AnimationEngine) *ParticleSystem {
	document := js.Global().Get("document")
	container := document.Call("createElement", "div")
	container.Get("style").Set("position", "fixed")
	container.Get("style").Set("top", "0")
	container.Get("style").Set("left", "0")
	container.Get("style").Set("width", "100%")
	container.Get("style").Set("height", "100%")
	container.Get("style").Set("pointerEvents", "none")
	container.Get("style").Set("zIndex", "9999")
	container.Set("id", "particle-container")

	document.Get("body").Call("appendChild", container)

	return &ParticleSystem{
		particles: make([]*Particle, 0),
		engine:    engine,
		container: container,
	}
}

// CreateParticle creates a new particle at the specified position
func (ps *ParticleSystem) CreateParticle(x, y float64, config ParticleConfig) *Particle {
	document := js.Global().Get("document")
	element := document.Call("createElement", "div")

	// Set up particle element
	style := element.Get("style")
	style.Set("position", "absolute")
	style.Set("borderRadius", "50%")
	style.Set("pointerEvents", "none")
	style.Set("willChange", "transform, opacity")

	// Random velocity based on config
	angle := rand.Float64() * 2 * math.Pi
	speed := config.MinSpeed + rand.Float64()*(config.MaxSpeed-config.MinSpeed)

	velocity := Vector2{
		X: math.Cos(angle) * speed,
		Y: math.Sin(angle) * speed,
	}

	// Random size
	size := config.MinSize + rand.Float64()*(config.MaxSize-config.MinSize)

	// Random color from palette
	color := config.Colors[rand.Intn(len(config.Colors))]

	particle := &Particle{
		ID:       fmt.Sprintf("particle_%d_%d", time.Now().UnixNano(), rand.Intn(10000)),
		Element:  element,
		Position: Vector2{X: x, Y: y},
		Velocity: velocity,
		Size:     size,
		Color:    color,
		Life:     1.0,
		MaxLife:  config.LifeTime,
		Gravity:  config.Gravity,
		Fade:     config.Fade,
		Scale:    config.Scale,
	}

	// Apply initial styles
	ps.updateParticleElement(particle)

	// Add to container
	ps.container.Call("appendChild", element)

	return particle
}

// ParticleConfig defines the configuration for particle creation
type ParticleConfig struct {
	Count    int
	MinSize  float64
	MaxSize  float64
	MinSpeed float64
	MaxSpeed float64
	LifeTime float64
	Gravity  float64
	Colors   []Color
	Fade     bool
	Scale    bool
}

// DefaultParticleConfig returns a default particle configuration
func DefaultParticleConfig() ParticleConfig {
	return ParticleConfig{
		Count:    20,
		MinSize:  4,
		MaxSize:  12,
		MinSpeed: 50,
		MaxSpeed: 200,
		LifeTime: 2.0,
		Gravity:  100,
		Colors: []Color{
			{R: 255, G: 100, B: 100, A: 1.0}, // Red
			{R: 100, G: 255, B: 100, A: 1.0}, // Green
			{R: 100, G: 100, B: 255, A: 1.0}, // Blue
			{R: 255, G: 255, B: 100, A: 1.0}, // Yellow
			{R: 255, G: 100, B: 255, A: 1.0}, // Magenta
			{R: 100, G: 255, B: 255, A: 1.0}, // Cyan
			{R: 255, G: 255, B: 255, A: 1.0}, // White
		},
		Fade:  true,
		Scale: true,
	}
}

// CreateParticleBurst creates a burst of particles at the specified position
func (ps *ParticleSystem) CreateParticleBurst(x, y float64, config ParticleConfig) {
	for i := 0; i < config.Count; i++ {
		particle := ps.CreateParticle(x, y, config)
		ps.particles = append(ps.particles, particle)

		// Start particle animation
		ps.animateParticle(particle)
	}
}

// animateParticle creates and starts the animation for a particle
func (ps *ParticleSystem) animateParticle(particle *Particle) {
	startTime := time.Now()
	duration := time.Duration(particle.MaxLife * float64(time.Second))

	anim := &Animation{
		ID:        particle.ID,
		StartTime: startTime,
		Duration:  duration,
		State:     AnimationPending,
		OnUpdate: func(progress float64) {
			ps.updateParticle(particle, progress)
		},
		OnComplete: func() {
			ps.removeParticle(particle)
		},
	}

	ps.engine.AddAnimation(anim)
}

// updateParticle updates a particle's position and appearance
func (ps *ParticleSystem) updateParticle(particle *Particle, progress float64) {
	deltaTime := particle.MaxLife * progress / 60.0 // Approximate frame time

	// Update position with physics
	particle.Velocity.Y += particle.Gravity * deltaTime
	particle.Position.X += particle.Velocity.X * deltaTime
	particle.Position.Y += particle.Velocity.Y * deltaTime

	// Update life
	particle.Life = 1.0 - progress

	// Update element
	ps.updateParticleElement(particle)
}

// updateParticleElement updates the DOM element for a particle
func (ps *ParticleSystem) updateParticleElement(particle *Particle) {
	style := particle.Element.Get("style")

	// Position
	style.Set("left", fmt.Sprintf("%.2fpx", particle.Position.X-particle.Size/2))
	style.Set("top", fmt.Sprintf("%.2fpx", particle.Position.Y-particle.Size/2))

	// Size
	currentSize := particle.Size
	if particle.Scale {
		currentSize *= particle.Life // Scale down as it dies
	}
	style.Set("width", fmt.Sprintf("%.2fpx", currentSize))
	style.Set("height", fmt.Sprintf("%.2fpx", currentSize))

	// Color and opacity
	color := particle.Color
	if particle.Fade {
		color.A = particle.Life // Fade out as it dies
	}
	style.Set("backgroundColor", color.ToString())

	// Add some glow effect
	glowColor := fmt.Sprintf("rgba(%d,%d,%d,%.2f)",
		int(color.R), int(color.G), int(color.B), color.A*0.5)
	style.Set("boxShadow", fmt.Sprintf("0 0 %.0fpx %s", currentSize*0.5, glowColor))
}

// removeParticle removes a particle from the system
func (ps *ParticleSystem) removeParticle(particle *Particle) {
	// Remove from DOM
	if particle.Element.Truthy() {
		ps.container.Call("removeChild", particle.Element)
	}

	// Remove from particles slice
	for i, p := range ps.particles {
		if p.ID == particle.ID {
			ps.particles = append(ps.particles[:i], ps.particles[i+1:]...)
			break
		}
	}
}

// GetActiveParticleCount returns the number of active particles
func (ps *ParticleSystem) GetActiveParticleCount() int {
	return len(ps.particles)
}

// Clear removes all particles from the system
func (ps *ParticleSystem) Clear() {
	for _, particle := range ps.particles {
		ps.engine.RemoveAnimation(particle.ID)
		if particle.Element.Truthy() {
			ps.container.Call("removeChild", particle.Element)
		}
	}
	ps.particles = ps.particles[:0]
}

// SetupMouseClickHandler sets up a global mouse click handler for particle bursts
func (ps *ParticleSystem) SetupMouseClickHandler(config ParticleConfig) {
	document := js.Global().Get("document")

	clickHandler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			event := args[0]
			x := event.Get("clientX").Float()
			y := event.Get("clientY").Float()

			// Create particle burst at click position
			ps.CreateParticleBurst(x, y, config)
		}
		return nil
	})

	document.Call("addEventListener", "click", clickHandler)
}

// Preset configurations for different effects

// FireworkConfig creates a firework-like particle burst
func FireworkConfig() ParticleConfig {
	return ParticleConfig{
		Count:    30,
		MinSize:  3,
		MaxSize:  8,
		MinSpeed: 80,
		MaxSpeed: 300,
		LifeTime: 3.0,
		Gravity:  50,
		Colors: []Color{
			{R: 255, G: 215, B: 0, A: 1.0},  // Gold
			{R: 255, G: 69, B: 0, A: 1.0},   // Red-Orange
			{R: 255, G: 20, B: 147, A: 1.0}, // Deep Pink
			{R: 138, G: 43, B: 226, A: 1.0}, // Blue Violet
			{R: 0, G: 191, B: 255, A: 1.0},  // Deep Sky Blue
		},
		Fade:  true,
		Scale: true,
	}
}

// SparkleConfig creates a gentle sparkle effect
func SparkleConfig() ParticleConfig {
	return ParticleConfig{
		Count:    15,
		MinSize:  2,
		MaxSize:  6,
		MinSpeed: 30,
		MaxSpeed: 80,
		LifeTime: 1.5,
		Gravity:  20,
		Colors: []Color{
			{R: 255, G: 255, B: 255, A: 1.0}, // White
			{R: 255, G: 255, B: 224, A: 1.0}, // Light Yellow
			{R: 255, G: 248, B: 220, A: 1.0}, // Cornsilk
			{R: 240, G: 248, B: 255, A: 1.0}, // Alice Blue
		},
		Fade:  true,
		Scale: false,
	}
}

// MagicConfig creates a magical particle effect
func MagicConfig() ParticleConfig {
	return ParticleConfig{
		Count:    25,
		MinSize:  4,
		MaxSize:  10,
		MinSpeed: 60,
		MaxSpeed: 150,
		LifeTime: 2.5,
		Gravity:  -20, // Negative gravity for floating effect
		Colors: []Color{
			{R: 186, G: 85, B: 211, A: 1.0},  // Medium Orchid
			{R: 147, G: 112, B: 219, A: 1.0}, // Medium Slate Blue
			{R: 123, G: 104, B: 238, A: 1.0}, // Medium Slate Blue
			{R: 72, G: 61, B: 139, A: 1.0},   // Dark Slate Blue
			{R: 138, G: 43, B: 226, A: 1.0},  // Blue Violet
		},
		Fade:  true,
		Scale: true,
	}
}
