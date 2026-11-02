// Package graphics draws resolution independent shapes and text on an ebiten image.
package graphics

import (
	"bytes"
	"image/color"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/gofont/gomedium"
	"golang.org/x/image/font/gofont/goregular"
)

// Font selects a typeface.
type Font int

const (
	Regular Font = iota
	Medium
)

var fontSources = [...]*text.GoTextFaceSource{
	Regular: mustLoadFont(goregular.TTF),
	Medium:  mustLoadFont(gomedium.TTF),
}

func mustLoadFont(ttf []byte) *text.GoTextFaceSource {
	src, err := text.NewGoTextFaceSource(bytes.NewReader(ttf))
	if err != nil {
		log.Fatal(err)
	}
	return src
}

// Align positions text horizontally relative to its anchor.
type Align = text.Align

const (
	AlignStart  = text.AlignStart
	AlignCenter = text.AlignCenter
	AlignEnd    = text.AlignEnd
)

// Viewport maps a fixed virtual coordinate space, centered and scaled to fit, onto a screen.
type Viewport struct {
	scale, offX, offY float64
}

// Fit returns the viewport showing a virtual area of width x height on a screen of screenW x screenH pixels.
func Fit(screenW, screenH, width, height float64) Viewport {
	scale := math.Min(screenW/width, screenH/height)
	return Viewport{scale: scale, offX: (screenW - width*scale) / 2, offY: (screenH - height*scale) / 2}
}

// ToVirtual converts screen pixels to virtual coordinates.
func (v Viewport) ToVirtual(x, y float64) (float64, float64) {
	return (x - v.offX) / v.scale, (y - v.offY) / v.scale
}

// Canvas draws in virtual coordinates on a destination image.
type Canvas struct {
	Viewport
	dst            *ebiten.Image
	shakeX, shakeY float64
}

// NewCanvas returns a canvas drawing on dst through v.
func NewCanvas(dst *ebiten.Image, v Viewport) *Canvas {
	return &Canvas{Viewport: v, dst: dst}
}

// Translate moves everything drawn afterwards by (dx, dy) virtual units, used for shaking.
func (c *Canvas) Translate(dx, dy float64) {
	c.shakeX, c.shakeY = dx, dy
}

func (c *Canvas) pt(x, y float64) (float32, float32) {
	return float32(c.offX + (x+c.shakeX)*c.scale), float32(c.offY + (y+c.shakeY)*c.scale)
}

func (c *Canvas) len(l float64) float32 { return float32(l * c.scale) }

func (c *Canvas) fill(p *vector.Path, clr color.Color) {
	op := &vector.DrawPathOptions{AntiAlias: true}
	op.ColorScale.ScaleWithColor(clr)
	vector.FillPath(c.dst, p, nil, op)
}

func (c *Canvas) stroke(p *vector.Path, width float64, clr color.Color) {
	op := &vector.DrawPathOptions{AntiAlias: true}
	op.ColorScale.ScaleWithColor(clr)
	vector.StrokePath(c.dst, p, &vector.StrokeOptions{
		Width:    c.len(width),
		LineCap:  vector.LineCapRound,
		LineJoin: vector.LineJoinRound,
	}, op)
}

func (c *Canvas) roundRectPath(x, y, w, h, r float64) *vector.Path {
	r = math.Min(r, math.Min(w, h)/2)
	var p vector.Path
	x0, y0 := c.pt(x, y)
	x1, y1 := c.pt(x+w, y+h)
	rr := c.len(r)
	p.MoveTo(x0+rr, y0)
	p.ArcTo(x1, y0, x1, y1, rr)
	p.ArcTo(x1, y1, x0, y1, rr)
	p.ArcTo(x0, y1, x0, y0, rr)
	p.ArcTo(x0, y0, x1, y0, rr)
	p.Close()
	return &p
}

// FillRect fills a rectangle with rounded corners of radius r.
func (c *Canvas) FillRect(x, y, w, h, r float64, clr color.Color) {
	c.fill(c.roundRectPath(x, y, w, h, r), clr)
}

// StrokeRect outlines a rectangle with rounded corners of radius r.
func (c *Canvas) StrokeRect(x, y, w, h, r, width float64, clr color.Color) {
	c.stroke(c.roundRectPath(x, y, w, h, r), width, clr)
}

// Line draws the first fraction progress (0 to 1) of the segment from (x0, y0) to (x1, y1).
func (c *Canvas) Line(x0, y0, x1, y1, width, progress float64, clr color.Color) {
	if progress <= 0 {
		return
	}
	progress = math.Min(progress, 1)
	var p vector.Path
	p.MoveTo(c.pt(x0, y0))
	p.LineTo(c.pt(x0+(x1-x0)*progress, y0+(y1-y0)*progress))
	c.stroke(&p, width, clr)
}

// FillQuad fills a w x h rectangle centered on (cx, cy) and rotated by angle radians.
func (c *Canvas) FillQuad(cx, cy, w, h, angle float64, clr color.Color) {
	sin, cos := math.Sincos(angle)
	var p vector.Path
	for i, corner := range [4][2]float64{{-1, -1}, {1, -1}, {1, 1}, {-1, 1}} {
		dx, dy := corner[0]*w/2, corner[1]*h/2
		x, y := c.pt(cx+dx*cos-dy*sin, cy+dx*sin+dy*cos)
		if i == 0 {
			p.MoveTo(x, y)
		} else {
			p.LineTo(x, y)
		}
	}
	p.Close()
	c.fill(&p, clr)
}

// Nought draws a circle of radius r centered on (cx, cy), traced clockwise from the top up to progress.
func (c *Canvas) Nought(cx, cy, r, width, progress float64, clr color.Color) {
	if progress <= 0 {
		return
	}
	start := -math.Pi / 2
	end := start + 2*math.Pi*math.Min(progress, 1)
	x, y := c.pt(cx, cy)
	var p vector.Path
	p.MoveTo(c.pt(cx+r*math.Cos(start), cy+r*math.Sin(start)))
	p.Arc(x, y, c.len(r), float32(start), float32(end), vector.Clockwise)
	c.stroke(&p, width, clr)
}

// Cross draws a cross fitting in a circle of radius r: the first stroke during the first half of progress, the second after.
func (c *Canvas) Cross(cx, cy, r, width, progress float64, clr color.Color) {
	d := r * math.Sqrt2 / 2
	c.Line(cx-d, cy-d, cx+d, cy+d, width, progress*2, clr)
	c.Line(cx+d, cy-d, cx-d, cy+d, width, progress*2-1, clr)
}

// Text draws a single line of text whose vertical center is y.
func (c *Canvas) Text(s string, x, y, size float64, font Font, align Align, clr color.Color) {
	face := &text.GoTextFace{Source: fontSources[font], Size: float64(c.len(size))}
	op := &text.DrawOptions{}
	op.PrimaryAlign = align
	op.SecondaryAlign = text.AlignCenter
	sx, sy := c.pt(x, y)
	op.GeoM.Translate(float64(sx), float64(sy))
	op.ColorScale.ScaleWithColor(clr)
	text.Draw(c.dst, s, face, op)
}

// TextWidth returns the width of s in virtual units.
func TextWidth(s string, size float64, font Font) float64 {
	w, _ := text.Measure(s, &text.GoTextFace{Source: fontSources[font], Size: size}, 0)
	return w
}

// Fade returns clr with its alpha multiplied by alpha.
func Fade(clr color.NRGBA, alpha float64) color.NRGBA {
	clr.A = uint8(float64(clr.A) * math.Max(0, math.Min(alpha, 1)))
	return clr
}

// Mix returns the linear interpolation between a and b, t being between 0 and 1.
func Mix(a, b color.NRGBA, t float64) color.NRGBA {
	t = math.Max(0, math.Min(t, 1))
	lerp := func(x, y uint8) uint8 { return uint8(float64(x) + (float64(y)-float64(x))*t) }
	return color.NRGBA{R: lerp(a.R, b.R), G: lerp(a.G, b.G), B: lerp(a.B, b.B), A: lerp(a.A, b.A)}
}
