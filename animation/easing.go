//go:build js && wasm

package animation

import "math"

// Easing functions for smooth animations
// All functions take a value from 0.0 to 1.0 and return the eased value

// Linear easing (no easing)
func Linear(t float64) float64 {
	return t
}

// Quadratic easing functions
func EaseInQuad(t float64) float64 {
	return t * t
}

func EaseOutQuad(t float64) float64 {
	return 1 - (1-t)*(1-t)
}

func EaseInOutQuad(t float64) float64 {
	if t < 0.5 {
		return 2 * t * t
	}
	return 1 - math.Pow(-2*t+2, 2)/2
}

// Cubic easing functions
func EaseInCubic(t float64) float64 {
	return t * t * t
}

func EaseOutCubic(t float64) float64 {
	return 1 - math.Pow(1-t, 3)
}

func EaseInOutCubic(t float64) float64 {
	if t < 0.5 {
		return 4 * t * t * t
	}
	return 1 - math.Pow(-2*t+2, 3)/2
}

// Quartic easing functions
func EaseInQuart(t float64) float64 {
	return t * t * t * t
}

func EaseOutQuart(t float64) float64 {
	return 1 - math.Pow(1-t, 4)
}

func EaseInOutQuart(t float64) float64 {
	if t < 0.5 {
		return 8 * t * t * t * t
	}
	return 1 - math.Pow(-2*t+2, 4)/2
}

// Quintic easing functions
func EaseInQuint(t float64) float64 {
	return t * t * t * t * t
}

func EaseOutQuint(t float64) float64 {
	return 1 - math.Pow(1-t, 5)
}

func EaseInOutQuint(t float64) float64 {
	if t < 0.5 {
		return 16 * t * t * t * t * t
	}
	return 1 - math.Pow(-2*t+2, 5)/2
}

// Sine easing functions
func EaseInSine(t float64) float64 {
	return 1 - math.Cos(t*math.Pi/2)
}

func EaseOutSine(t float64) float64 {
	return math.Sin(t * math.Pi / 2)
}

func EaseInOutSine(t float64) float64 {
	return -(math.Cos(math.Pi*t) - 1) / 2
}

// Exponential easing functions
func EaseInExpo(t float64) float64 {
	if t == 0 {
		return 0
	}
	return math.Pow(2, 10*(t-1))
}

func EaseOutExpo(t float64) float64 {
	if t == 1 {
		return 1
	}
	return 1 - math.Pow(2, -10*t)
}

func EaseInOutExpo(t float64) float64 {
	if t == 0 {
		return 0
	}
	if t == 1 {
		return 1
	}
	if t < 0.5 {
		return math.Pow(2, 20*t-10) / 2
	}
	return (2 - math.Pow(2, -20*t+10)) / 2
}

// Circular easing functions
func EaseInCirc(t float64) float64 {
	return 1 - math.Sqrt(1-t*t)
}

func EaseOutCirc(t float64) float64 {
	return math.Sqrt(1 - math.Pow(t-1, 2))
}

func EaseInOutCirc(t float64) float64 {
	if t < 0.5 {
		return (1 - math.Sqrt(1-math.Pow(2*t, 2))) / 2
	}
	return (math.Sqrt(1-math.Pow(-2*t+2, 2)) + 1) / 2
}

// Back easing functions (overshoot)
func EaseInBack(t float64) float64 {
	c1 := 1.70158
	c3 := c1 + 1
	return c3*t*t*t - c1*t*t
}

func EaseOutBack(t float64) float64 {
	c1 := 1.70158
	c3 := c1 + 1
	return 1 + c3*math.Pow(t-1, 3) + c1*math.Pow(t-1, 2)
}

func EaseInOutBack(t float64) float64 {
	c1 := 1.70158
	c2 := c1 * 1.525
	if t < 0.5 {
		return (math.Pow(2*t, 2) * ((c2+1)*2*t - c2)) / 2
	}
	return (math.Pow(2*t-2, 2)*((c2+1)*(t*2-2)+c2) + 2) / 2
}

// Elastic easing functions (spring effect)
func EaseInElastic(t float64) float64 {
	c4 := (2 * math.Pi) / 3
	if t == 0 {
		return 0
	}
	if t == 1 {
		return 1
	}
	return -math.Pow(2, 10*t-10) * math.Sin((t*10-10.75)*c4)
}

func EaseOutElastic(t float64) float64 {
	c4 := (2 * math.Pi) / 3
	if t == 0 {
		return 0
	}
	if t == 1 {
		return 1
	}
	return math.Pow(2, -10*t)*math.Sin((t*10-0.75)*c4) + 1
}

func EaseInOutElastic(t float64) float64 {
	c5 := (2 * math.Pi) / 4.5
	if t == 0 {
		return 0
	}
	if t == 1 {
		return 1
	}
	if t < 0.5 {
		return -(math.Pow(2, 20*t-10) * math.Sin((20*t-11.125)*c5)) / 2
	}
	return (math.Pow(2, -20*t+10)*math.Sin((20*t-11.125)*c5))/2 + 1
}

// Bounce easing functions
func EaseInBounce(t float64) float64 {
	return 1 - EaseOutBounce(1-t)
}

func EaseOutBounce(t float64) float64 {
	n1 := 7.5625
	d1 := 2.75

	if t < 1/d1 {
		return n1 * t * t
	} else if t < 2/d1 {
		t -= 1.5 / d1
		return n1*t*t + 0.75
	} else if t < 2.5/d1 {
		t -= 2.25 / d1
		return n1*t*t + 0.9375
	} else {
		t -= 2.625 / d1
		return n1*t*t + 0.984375
	}
}

func EaseInOutBounce(t float64) float64 {
	if t < 0.5 {
		return (1 - EaseOutBounce(1-2*t)) / 2
	}
	return (1 + EaseOutBounce(2*t-1)) / 2
}

// Custom easing function builders
func CreateBezier(x1, y1, x2, y2 float64) EasingFunc {
	// Simplified cubic bezier approximation
	return func(t float64) float64 {
		// This is a simplified version - a full implementation would
		// require solving the cubic bezier equation
		return EaseInOutCubic(t) // Fallback for now
	}
}

// Spring easing with custom parameters
func CreateSpring(tension, friction float64) EasingFunc {
	return func(t float64) float64 {
		// Simplified spring physics
		omega := math.Sqrt(tension)
		zeta := friction / (2 * math.Sqrt(tension))

		if zeta < 1 {
			// Underdamped
			omegaD := omega * math.Sqrt(1-zeta*zeta)
			return 1 - math.Exp(-zeta*omega*t)*math.Cos(omegaD*t)
		} else if zeta == 1 {
			// Critically damped
			return 1 - math.Exp(-omega*t)*(1+omega*t)
		} else {
			// Overdamped
			r1 := -omega * (zeta + math.Sqrt(zeta*zeta-1))
			r2 := -omega * (zeta - math.Sqrt(zeta*zeta-1))
			return 1 - (r2*math.Exp(r1*t)-r1*math.Exp(r2*t))/(r2-r1)
		}
	}
}

// Predefined common easing functions
var (
	// Most commonly used easing functions
	EaseIn    = EaseInCubic
	EaseOut   = EaseOutCubic
	EaseInOut = EaseInOutCubic

	// Smooth and natural feeling
	Smooth = EaseInOutSine

	// Quick and snappy
	Snappy = EaseOutBack

	// Bouncy and playful
	Bouncy = EaseOutBounce

	// Elastic and spring-like
	Elastic = EaseOutElastic
)
