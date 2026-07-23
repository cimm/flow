package ui

import (
	"time"

	"github.com/diamondburned/gotk4/pkg/cairo"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type DataPoint struct {
	Time  time.Time
	Value float64
}

type Graph struct {
	*gtk.DrawingArea
	data     []DataPoint
	posDummy *gtk.Label
	negDummy *gtk.Label
}

func NewGraph(parentBox *gtk.Box) *Graph {
	pos := gtk.NewLabel("")
	pos.AddCSSClass("power-positive")
	pos.SetVisible(false)

	neg := gtk.NewLabel("")
	neg.AddCSSClass("power-negative")
	neg.SetVisible(false)

	if parentBox != nil {
		parentBox.Append(pos)
		parentBox.Append(neg)
	}

	g := &Graph{
		DrawingArea: gtk.NewDrawingArea(),
		posDummy:    pos,
		negDummy:    neg,
	}
	g.SetDrawFunc(g.draw)
	g.SetVExpand(true)
	g.SetHExpand(true)
	return g
}

func (g *Graph) AddValue(val float64) {
	now := time.Now()
	g.data = append(g.data, DataPoint{Time: now, Value: val})

	// Keep up to 1 minute of data
	cutoff := now.Add(-time.Minute)
	var i int
	for i = 0; i < len(g.data); i++ {
		if g.data[i].Time.After(cutoff) {
			break
		}
	}
	if i > 0 {
		g.data = g.data[i:]
	}

	g.QueueDraw()
}

func (g *Graph) draw(area *gtk.DrawingArea, cr *cairo.Context, width, height int) {
	if len(g.data) == 0 {
		return
	}

	minY := g.data[0].Value
	maxY := g.data[0].Value
	for _, d := range g.data {
		if d.Value < minY {
			minY = d.Value
		}
		if d.Value > maxY {
			maxY = d.Value
		}
	}

	pad := (maxY - minY) * 0.1
	if pad == 0 {
		pad = 10
	}
	minY -= pad
	maxY += pad

	// Find yZero where value == 0
	yZero := float64(height)
	if maxY > minY {
		yZero = float64(height) * (1.0 - (0.0-minY)/(maxY-minY))
	}

	posColor := g.posDummy.Color()
	negColor := g.negDummy.Color()

	pat, _ := cairo.NewPatternLinear(0, 0, 0, float64(height))
	ratio := yZero / float64(height)

	if ratio <= 0 {
		pat.AddColorStopRGBA(0, float64(negColor.Red()), float64(negColor.Green()), float64(negColor.Blue()), float64(negColor.Alpha()))
	} else if ratio >= 1 {
		pat.AddColorStopRGBA(0, float64(posColor.Red()), float64(posColor.Green()), float64(posColor.Blue()), float64(posColor.Alpha()))
	} else {
		pat.AddColorStopRGBA(0, float64(posColor.Red()), float64(posColor.Green()), float64(posColor.Blue()), float64(posColor.Alpha()))
		pat.AddColorStopRGBA(ratio, float64(posColor.Red()), float64(posColor.Green()), float64(posColor.Blue()), float64(posColor.Alpha()))
		pat.AddColorStopRGBA(ratio, float64(negColor.Red()), float64(negColor.Green()), float64(negColor.Blue()), float64(negColor.Alpha()))
		pat.AddColorStopRGBA(1, float64(negColor.Red()), float64(negColor.Green()), float64(negColor.Blue()), float64(negColor.Alpha()))
	}

	cr.SetSource(pat)
	cr.SetLineWidth(2.0)

	now := time.Now()
	started := false
	for _, d := range g.data {
		age := now.Sub(d.Time).Seconds()
		if age > 60.0 {
			age = 60.0
		} else if age < 0 {
			age = 0
		}

		// Newest (age=0) on the right (x=width)
		// Oldest (age=60) on the left (x=0)
		x := float64(width) * (1.0 - (age / 60.0))
		y := float64(height)
		if maxY > minY {
			y = float64(height) * (1.0 - (d.Value-minY)/(maxY-minY))
		}

		if !started {
			cr.MoveTo(x, y)
			started = true
		} else {
			cr.LineTo(x, y)
		}
	}
	cr.Stroke()
}
