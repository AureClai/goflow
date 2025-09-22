//go:build js && wasm

package animation

import (
	"sort"
	"time"
)

// Timeline manages complex animation sequences
type Timeline struct {
	animations []TimelineAnimation
	duration   time.Duration
	playhead   time.Duration
	startTime  time.Time
	state      AnimationState
	loop       bool
	onComplete func()
	onUpdate   func(progress float64)
	engine     *AnimationEngine
}

// TimelineAnimation represents an animation within a timeline
type TimelineAnimation struct {
	StartTime time.Duration
	Animation *Animation
}

// NewTimeline creates a new animation timeline
func NewTimeline(engine *AnimationEngine) *Timeline {
	return &Timeline{
		animations: make([]TimelineAnimation, 0),
		engine:     engine,
		state:      AnimationPending,
	}
}

// AddAnimation adds an animation to the timeline at a specific time
func (t *Timeline) AddAnimation(startTime time.Duration, anim *Animation) *Timeline {
	t.animations = append(t.animations, TimelineAnimation{
		StartTime: startTime,
		Animation: anim,
	})

	// Update total duration
	animEnd := startTime + anim.Duration + anim.Delay
	if animEnd > t.duration {
		t.duration = animEnd
	}

	// Sort animations by start time for efficient processing
	sort.Slice(t.animations, func(i, j int) bool {
		return t.animations[i].StartTime < t.animations[j].StartTime
	})

	return t
}

// AddAnimationAfter adds an animation after the previous one completes
func (t *Timeline) AddAnimationAfter(anim *Animation) *Timeline {
	return t.AddAnimation(t.duration, anim)
}

// AddAnimationWith adds an animation to run simultaneously with the last added animation
func (t *Timeline) AddAnimationWith(anim *Animation) *Timeline {
	if len(t.animations) == 0 {
		return t.AddAnimation(0, anim)
	}

	lastAnim := t.animations[len(t.animations)-1]
	return t.AddAnimation(lastAnim.StartTime, anim)
}

// SetLoop enables or disables timeline looping
func (t *Timeline) SetLoop(loop bool) *Timeline {
	t.loop = loop
	return t
}

// OnComplete sets a callback for when the timeline completes
func (t *Timeline) OnComplete(callback func()) *Timeline {
	t.onComplete = callback
	return t
}

// OnUpdate sets a callback for timeline progress updates
func (t *Timeline) OnUpdate(callback func(progress float64)) *Timeline {
	t.onUpdate = callback
	return t
}

// Play starts the timeline
func (t *Timeline) Play() {
	t.startTime = time.Now()
	t.state = AnimationRunning
	t.playhead = 0

	// Start all animations that should be active at the beginning
	t.updateActiveAnimations()
}

// Pause pauses the timeline
func (t *Timeline) Pause() {
	if t.state == AnimationRunning {
		t.state = AnimationPaused

		// Pause all active animations
		for _, anim := range t.animations {
			if anim.Animation.State == AnimationRunning {
				t.engine.PauseAnimation(anim.Animation.ID)
			}
		}
	}
}

// Resume resumes the timeline
func (t *Timeline) Resume() {
	if t.state == AnimationPaused {
		t.state = AnimationRunning
		t.startTime = time.Now().Add(-t.playhead)

		// Resume all paused animations
		for _, anim := range t.animations {
			if anim.Animation.State == AnimationPaused {
				t.engine.ResumeAnimation(anim.Animation.ID)
			}
		}

		t.updateActiveAnimations()
	}
}

// Stop stops the timeline
func (t *Timeline) Stop() {
	t.state = AnimationCancelled
	t.playhead = 0

	// Stop all animations
	for _, anim := range t.animations {
		t.engine.RemoveAnimation(anim.Animation.ID)
	}
}

// Seek moves the playhead to a specific time
func (t *Timeline) Seek(seekTime time.Duration) {
	if seekTime < 0 {
		seekTime = 0
	}
	if seekTime > t.duration {
		seekTime = t.duration
	}

	t.playhead = seekTime
	t.startTime = time.Now().Add(-t.playhead)

	// Update all animations based on new playhead position
	t.updateActiveAnimations()
}

// GetProgress returns the current timeline progress (0.0 to 1.0)
func (t *Timeline) GetProgress() float64 {
	if t.duration == 0 {
		return 0
	}
	return float64(t.playhead) / float64(t.duration)
}

// GetDuration returns the total timeline duration
func (t *Timeline) GetDuration() time.Duration {
	return t.duration
}

// GetPlayhead returns the current playhead position
func (t *Timeline) GetPlayhead() time.Duration {
	return t.playhead
}

// IsPlaying returns true if the timeline is currently playing
func (t *Timeline) IsPlaying() bool {
	return t.state == AnimationRunning
}

// IsPaused returns true if the timeline is paused
func (t *Timeline) IsPaused() bool {
	return t.state == AnimationPaused
}

// IsComplete returns true if the timeline has completed
func (t *Timeline) IsComplete() bool {
	return t.state == AnimationComplete
}

// Update updates the timeline (called by the animation engine)
func (t *Timeline) Update() {
	if t.state != AnimationRunning {
		return
	}

	// Update playhead
	t.playhead = time.Since(t.startTime)

	// Check if timeline is complete
	if t.playhead >= t.duration {
		if t.loop {
			// Restart timeline
			t.playhead = 0
			t.startTime = time.Now()
			t.updateActiveAnimations()
		} else {
			// Complete timeline
			t.state = AnimationComplete
			t.playhead = t.duration

			if t.onComplete != nil {
				t.onComplete()
			}
			return
		}
	}

	// Update active animations
	t.updateActiveAnimations()

	// Call update callback
	if t.onUpdate != nil {
		t.onUpdate(t.GetProgress())
	}
}

// updateActiveAnimations starts/stops animations based on current playhead
func (t *Timeline) updateActiveAnimations() {
	for _, timelineAnim := range t.animations {
		animStartTime := timelineAnim.StartTime
		animEndTime := animStartTime + timelineAnim.Animation.Duration + timelineAnim.Animation.Delay

		// Check if animation should be active
		shouldBeActive := t.playhead >= animStartTime && t.playhead < animEndTime
		isCurrentlyActive := timelineAnim.Animation.State == AnimationRunning ||
			timelineAnim.Animation.State == AnimationPending

		if shouldBeActive && !isCurrentlyActive {
			// Start animation
			anim := timelineAnim.Animation
			anim.StartTime = t.startTime.Add(animStartTime)
			anim.State = AnimationPending
			t.engine.AddAnimation(anim)
		} else if !shouldBeActive && isCurrentlyActive {
			// Stop animation
			t.engine.RemoveAnimation(timelineAnim.Animation.ID)
		}
	}
}

// Sequence creates a timeline with animations running one after another
func Sequence(engine *AnimationEngine, animations ...*Animation) *Timeline {
	timeline := NewTimeline(engine)
	currentTime := time.Duration(0)

	for _, anim := range animations {
		timeline.AddAnimation(currentTime, anim)
		currentTime += anim.Duration + anim.Delay
	}

	return timeline
}

// Parallel creates a timeline with all animations running simultaneously
func Parallel(engine *AnimationEngine, animations ...*Animation) *Timeline {
	timeline := NewTimeline(engine)

	for _, anim := range animations {
		timeline.AddAnimation(0, anim)
	}

	return timeline
}

// StaggerTimeline creates a timeline with animations starting at staggered intervals
func StaggerTimeline(engine *AnimationEngine, staggerDelay time.Duration, animations ...*Animation) *Timeline {
	timeline := NewTimeline(engine)
	currentTime := time.Duration(0)

	for _, anim := range animations {
		timeline.AddAnimation(currentTime, anim)
		currentTime += staggerDelay
	}

	return timeline
}

// TimelineBuilder provides a fluent interface for building complex timelines
type TimelineBuilder struct {
	timeline *Timeline
}

// NewTimelineBuilder creates a new timeline builder
func NewTimelineBuilder(engine *AnimationEngine) *TimelineBuilder {
	return &TimelineBuilder{
		timeline: NewTimeline(engine),
	}
}

// At adds an animation at a specific time
func (tb *TimelineBuilder) At(time time.Duration, anim *Animation) *TimelineBuilder {
	tb.timeline.AddAnimation(time, anim)
	return tb
}

// Then adds an animation after the previous one
func (tb *TimelineBuilder) Then(anim *Animation) *TimelineBuilder {
	tb.timeline.AddAnimationAfter(anim)
	return tb
}

// With adds an animation alongside the previous one
func (tb *TimelineBuilder) With(anim *Animation) *TimelineBuilder {
	tb.timeline.AddAnimationWith(anim)
	return tb
}

// Loop enables looping
func (tb *TimelineBuilder) Loop() *TimelineBuilder {
	tb.timeline.SetLoop(true)
	return tb
}

// Build returns the constructed timeline
func (tb *TimelineBuilder) Build() *Timeline {
	return tb.timeline
}
