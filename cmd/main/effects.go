package main

import (
	"image/color"
	"math"
	"math/rand/v2"
)

// Animation timings, in seconds.
const (
	pieceDuration     = 0.22 // a symbol being drawn in its cell
	highlightDuration = 0.18 // playable boards fading in and out
	boardDelay        = 0.18 // from the winning move to the small board animation
	strikeDuration    = 0.25 // line through the three winning cells
	bigSymbolDuration = 0.40 // large symbol covering a won small board
	endLineDelay      = 0.55 // from the last move to the line through the winning boards
	endLineDuration   = 0.50
	outcomeDelay      = endLineDelay + endLineDuration // confetti, shake or fade
	shakeDuration     = 0.45
	fadeDuration      = 0.50
	cardDelay         = outcomeDelay + 0.35 // result card
	cardDuration      = 0.35
	confettiLifetime  = 2.6
)

// progress returns how far (0 to 1) an animation of the given duration is, delay seconds after start.
func progress(now, start, delay, duration float64) float64 {
	return clamp01((now - start - delay) / duration)
}

func clamp01(t float64) float64 { return math.Max(0, math.Min(t, 1)) }

func easeOutCubic(t float64) float64 { return 1 - math.Pow(1-t, 3) }

func easeInOutCubic(t float64) float64 {
	if t < 0.5 {
		return 4 * t * t * t
	}
	return 1 - math.Pow(-2*t+2, 3)/2
}

// approach moves value towards target, covering the whole distance in duration seconds.
func approach(value, target, dt, duration float64) float64 {
	step := dt / duration
	if value < target {
		return math.Min(value+step, target)
	}
	return math.Max(value-step, target)
}

// shakeOffset returns the horizontal offset of a damped shake t seconds after it started.
func shakeOffset(t float64) float64 {
	if t < 0 || t > shakeDuration {
		return 0
	}
	damping := 1 - t/shakeDuration
	return 9 * damping * damping * math.Sin(t*48)
}

type confetto struct {
	x, y, vx, vy float64
	angle, spin  float64
	w, h         float64
	age          float64
	clr          color.NRGBA
}

// confetti is a short burst of paper pieces falling over the board.
type confetti []confetto

func (c *confetti) burst(x, y float64, count int) {
	for range count {
		angle := -math.Pi/2 + (rand.Float64()-0.5)*1.6
		speed := 420 + rand.Float64()*520
		*c = append(*c, confetto{
			x: x, y: y,
			vx:    math.Cos(angle) * speed,
			vy:    math.Sin(angle) * speed,
			angle: rand.Float64() * math.Pi,
			spin:  (rand.Float64() - 0.5) * 14,
			w:     7 + rand.Float64()*6,
			h:     4 + rand.Float64()*4,
			clr:   colorConfetti[rand.IntN(len(colorConfetti))],
		})
	}
}

func (c *confetti) update(dt float64) {
	const gravity, drag = 1100.0, 1.6
	alive := (*c)[:0]
	for _, p := range *c {
		p.age += dt
		p.vy += gravity * dt
		p.vx -= p.vx * drag * dt
		p.vy -= p.vy * drag * dt
		p.x += p.vx * dt
		p.y += p.vy * dt
		p.angle += p.spin * dt
		if p.age < confettiLifetime && p.y < viewHeight+20 {
			alive = append(alive, p)
		}
	}
	*c = alive
}
