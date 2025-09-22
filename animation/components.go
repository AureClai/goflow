//go:build js && wasm

package animation

import (
	"fmt"
	"syscall/js"
	"time"

	"github.com/AureClai/goflow/component"
	"github.com/AureClai/goflow/vdom"
)

// AnimatedComponent wraps any component with animation capabilities
type AnimatedComponent struct {
	child      vdom.Component
	animations map[string]*Animation
	engine     *AnimationEngine
	element    js.Value
	mounted    bool
}

// NewAnimatedComponent creates a new animated component wrapper
func NewAnimatedComponent(child vdom.Component, engine *AnimationEngine) *AnimatedComponent {
	return &AnimatedComponent{
		child:      child,
		animations: make(map[string]*Animation),
		engine:     engine,
	}
}

// Render renders the wrapped component with animation classes
func (ac *AnimatedComponent) Render() *vdom.VNode {
	node := ac.child.Render()

	// Add animation-related classes
	if node.Props == nil {
		node.Props = make(map[string]interface{})
	}

	classes := "animated"
	if existing, ok := node.Props["class"].(string); ok {
		classes = existing + " " + classes
	}
	node.Props["class"] = classes

	return node
}

// FadeIn creates a fade-in animation component
type FadeIn struct {
	*AnimatedComponent
	duration time.Duration
	delay    time.Duration
	easing   EasingFunc
}

// NewFadeIn creates a fade-in animated component
func NewFadeIn(child vdom.Component, engine *AnimationEngine) *FadeIn {
	return &FadeIn{
		AnimatedComponent: NewAnimatedComponent(child, engine),
		duration:          500 * time.Millisecond,
		delay:             0,
		easing:            EaseOut,
	}
}

// SetDuration sets the animation duration
func (f *FadeIn) SetDuration(duration time.Duration) *FadeIn {
	f.duration = duration
	return f
}

// SetDelay sets the animation delay
func (f *FadeIn) SetDelay(delay time.Duration) *FadeIn {
	f.delay = delay
	return f
}

// SetEasing sets the easing function
func (f *FadeIn) SetEasing(easing EasingFunc) *FadeIn {
	f.easing = easing
	return f
}

// Render renders the fade-in component
func (f *FadeIn) Render() *vdom.VNode {
	node := f.AnimatedComponent.Render()

	// Add fade-in specific classes
	classes := node.Props["class"].(string) + " fade-in"
	node.Props["class"] = classes
	node.Props["style"] = "opacity: 0;"

	// Start animation when component mounts
	f.engine.OnNextFrame(func() {
		f.startFadeIn()
	})

	return node
}

// startFadeIn starts the fade-in animation
func (f *FadeIn) startFadeIn() {
	anim := &Animation{
		ID:       fmt.Sprintf("fadein_%d", time.Now().UnixNano()),
		Duration: f.duration,
		Delay:    f.delay,
		Easing:   f.easing,
		Properties: []PropertyAnimation{
			{
				Property: "opacity",
				From:     0.0,
				To:       1.0,
				Unit:     "",
			},
		},
	}

	f.engine.AddAnimation(anim)
}

// SlideIn creates a slide-in animation component
type SlideIn struct {
	*AnimatedComponent
	duration  time.Duration
	delay     time.Duration
	easing    EasingFunc
	direction string // "left", "right", "up", "down"
	distance  int    // pixels
}

// NewSlideIn creates a slide-in animated component
func NewSlideIn(child vdom.Component, engine *AnimationEngine) *SlideIn {
	return &SlideIn{
		AnimatedComponent: NewAnimatedComponent(child, engine),
		duration:          600 * time.Millisecond,
		delay:             0,
		easing:            EaseOut,
		direction:         "left",
		distance:          50,
	}
}

// SetDirection sets the slide direction
func (s *SlideIn) SetDirection(direction string) *SlideIn {
	s.direction = direction
	return s
}

// SetDistance sets the slide distance
func (s *SlideIn) SetDistance(distance int) *SlideIn {
	s.distance = distance
	return s
}

// SetDuration sets the animation duration
func (s *SlideIn) SetDuration(duration time.Duration) *SlideIn {
	s.duration = duration
	return s
}

// SetEasing sets the easing function
func (s *SlideIn) SetEasing(easing EasingFunc) *SlideIn {
	s.easing = easing
	return s
}

// Render renders the slide-in component
func (s *SlideIn) Render() *vdom.VNode {
	node := s.AnimatedComponent.Render()

	// Add slide-in specific classes
	classes := node.Props["class"].(string) + " slide-in slide-in-" + s.direction
	node.Props["class"] = classes

	// Set initial transform based on direction
	var initialTransform string
	switch s.direction {
	case "left":
		initialTransform = fmt.Sprintf("translateX(-%dpx)", s.distance)
	case "right":
		initialTransform = fmt.Sprintf("translateX(%dpx)", s.distance)
	case "up":
		initialTransform = fmt.Sprintf("translateY(-%dpx)", s.distance)
	case "down":
		initialTransform = fmt.Sprintf("translateY(%dpx)", s.distance)
	}

	node.Props["style"] = fmt.Sprintf("transform: %s;", initialTransform)

	// Start animation when component mounts
	s.engine.OnNextFrame(func() {
		s.startSlideIn()
	})

	return node
}

// startSlideIn starts the slide-in animation
func (s *SlideIn) startSlideIn() {
	anim := &Animation{
		ID:       fmt.Sprintf("slidein_%d", time.Now().UnixNano()),
		Duration: s.duration,
		Delay:    s.delay,
		Easing:   s.easing,
		Properties: []PropertyAnimation{
			{
				Property: "transform",
				From:     s.getInitialTransform(),
				To:       "translateX(0px) translateY(0px)",
				Unit:     "",
			},
		},
	}

	s.engine.AddAnimation(anim)
}

// getInitialTransform returns the initial transform based on direction
func (s *SlideIn) getInitialTransform() string {
	switch s.direction {
	case "left":
		return fmt.Sprintf("translateX(-%dpx)", s.distance)
	case "right":
		return fmt.Sprintf("translateX(%dpx)", s.distance)
	case "up":
		return fmt.Sprintf("translateY(-%dpx)", s.distance)
	case "down":
		return fmt.Sprintf("translateY(%dpx)", s.distance)
	default:
		return "translateX(0px) translateY(0px)"
	}
}

// ScaleIn creates a scale-in animation component
type ScaleIn struct {
	*AnimatedComponent
	duration  time.Duration
	delay     time.Duration
	easing    EasingFunc
	fromScale float64
	toScale   float64
}

// NewScaleIn creates a scale-in animated component
func NewScaleIn(child vdom.Component, engine *AnimationEngine) *ScaleIn {
	return &ScaleIn{
		AnimatedComponent: NewAnimatedComponent(child, engine),
		duration:          400 * time.Millisecond,
		delay:             0,
		easing:            EaseOutBack,
		fromScale:         0.0,
		toScale:           1.0,
	}
}

// SetScale sets the from and to scale values
func (sc *ScaleIn) SetScale(from, to float64) *ScaleIn {
	sc.fromScale = from
	sc.toScale = to
	return sc
}

// SetDuration sets the animation duration
func (sc *ScaleIn) SetDuration(duration time.Duration) *ScaleIn {
	sc.duration = duration
	return sc
}

// SetEasing sets the easing function
func (sc *ScaleIn) SetEasing(easing EasingFunc) *ScaleIn {
	sc.easing = easing
	return sc
}

// Render renders the scale-in component
func (sc *ScaleIn) Render() *vdom.VNode {
	node := sc.AnimatedComponent.Render()

	// Add scale-in specific classes
	classes := node.Props["class"].(string) + " scale-in"
	node.Props["class"] = classes
	node.Props["style"] = fmt.Sprintf("transform: scale(%f);", sc.fromScale)

	// Start animation when component mounts
	sc.engine.OnNextFrame(func() {
		sc.startScaleIn()
	})

	return node
}

// startScaleIn starts the scale-in animation
func (sc *ScaleIn) startScaleIn() {
	anim := &Animation{
		ID:       fmt.Sprintf("scalein_%d", time.Now().UnixNano()),
		Duration: sc.duration,
		Delay:    sc.delay,
		Easing:   sc.easing,
		Properties: []PropertyAnimation{
			{
				Property: "transform",
				From:     fmt.Sprintf("scale(%f)", sc.fromScale),
				To:       fmt.Sprintf("scale(%f)", sc.toScale),
				Unit:     "",
			},
		},
	}

	sc.engine.AddAnimation(anim)
}

// Stagger creates staggered animations for multiple components
type Stagger struct {
	children    []vdom.Component
	engine      *AnimationEngine
	staggerTime time.Duration
	animation   func(vdom.Component, *AnimationEngine) vdom.Component
}

// NewStagger creates a new stagger animation
func NewStagger(engine *AnimationEngine) *Stagger {
	return &Stagger{
		children:    make([]vdom.Component, 0),
		engine:      engine,
		staggerTime: 100 * time.Millisecond,
		animation: func(child vdom.Component, engine *AnimationEngine) vdom.Component {
			return NewFadeIn(child, engine)
		},
	}
}

// AddChild adds a child component to the stagger
func (s *Stagger) AddChild(child vdom.Component) *Stagger {
	s.children = append(s.children, child)
	return s
}

// SetStaggerTime sets the time between each animation
func (s *Stagger) SetStaggerTime(staggerTime time.Duration) *Stagger {
	s.staggerTime = staggerTime
	return s
}

// SetAnimation sets the animation function to apply to each child
func (s *Stagger) SetAnimation(animation func(vdom.Component, *AnimationEngine) vdom.Component) *Stagger {
	s.animation = animation
	return s
}

// Render renders the staggered components
func (s *Stagger) Render() *vdom.VNode {
	container := component.NewContainer().SetClass("stagger-container")

	for i, child := range s.children {
		// Apply animation with delay
		animatedChild := s.animation(child, s.engine)

		// Add stagger delay
		delay := time.Duration(i) * s.staggerTime
		s.engine.OnNextFrame(func() {
			// Schedule animation with delay
			time.AfterFunc(delay, func() {
				// Animation will be handled by the animated component
			})
		})

		container.AddChild(animatedChild)
	}

	return container.Render()
}

// Helper functions for creating common animated components

// AnimatedButton creates an animated button with hover effects
func AnimatedButton(text string, onClick func(), engine *AnimationEngine) vdom.Component {
	button := component.NewButton(text, onClick).SetClass("animated-button")
	return NewScaleIn(button, engine).SetEasing(EaseOutBack)
}

// AnimatedList creates an animated list with staggered item animations
func AnimatedList(items []string, engine *AnimationEngine) vdom.Component {
	stagger := NewStagger(engine).SetStaggerTime(150 * time.Millisecond)

	for _, item := range items {
		listItem := component.NewText(item).SetClass("list-item")
		stagger.AddChild(listItem)
	}

	return stagger
}

// AnimatedCard creates an animated card component
func AnimatedCard(title, content string, engine *AnimationEngine) vdom.Component {
	card := component.NewContainer().SetClass("card")

	cardTitle := component.NewHeading(title, 3).SetClass("card-title")
	cardContent := component.NewParagraph(content).SetClass("card-content")

	card.AddChild(cardTitle)
	card.AddChild(cardContent)

	return NewSlideIn(card, engine).SetDirection("up").SetDistance(30)
}
