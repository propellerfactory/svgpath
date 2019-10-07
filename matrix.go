// port of https://github.com/fontello/svgpath
package svgpath

import (
	"math"
)

// combine 2 matrixes
// m1, m2 - [a, b, c, d, e, g]
//
func combine(m1, m2 []float64) []float64 {
	return []float64{
		m1[0]*m2[0] + m1[2]*m2[1],
		m1[1]*m2[0] + m1[3]*m2[1],
		m1[0]*m2[2] + m1[2]*m2[3],
		m1[1]*m2[2] + m1[3]*m2[3],
		m1[0]*m2[4] + m1[2]*m2[5] + m1[4],
		m1[1]*m2[4] + m1[3]*m2[5] + m1[5],
	}
}

type Matrix struct {
	queue [][]float64
	cache []float64
}

func NewMatrix() *Matrix {
	return &Matrix{
		// list of matrixes to apply
		queue: [][]float64{},
		cache: nil,
	}
}

func (mx *Matrix) Matrix(m []float64) {
	if m[0] == 1.0 && m[1] == 0.0 && m[2] == 0.0 && m[3] == 1.0 && m[4] == 0.0 && m[5] == 0.0 {
		return
	}
	mx.queue = append(mx.queue, m)
	mx.cache = nil
}

func (mx *Matrix) Translate(tx, ty float64) {
	if tx != 0.0 || ty != 0.0 {
		mx.queue = append(mx.queue, []float64{1.0, 0.0, 0.0, 1.0, tx, ty})
		mx.cache = nil
	}
}

func (mx *Matrix) Scale(sx, sy float64) {
	if sx != 1.0 || sy != 1.0 {
		mx.queue = append(mx.queue, []float64{sx, 0, 0, sy, 0, 0})
		mx.cache = nil
	}
}

func (mx *Matrix) Rotate(angle, rx, ry float64) {
	if angle != 0.0 {
		mx.Translate(rx, ry)

		rad := angle * math.Pi / 180.0
		cos := math.Cos(rad)
		sin := math.Sin(rad)

		mx.queue = append(mx.queue, []float64{cos, sin, -sin, cos, 0.0, 0.0})
		mx.cache = nil

		mx.Translate(-rx, -ry)
	}
}

func (mx *Matrix) SkewX(angle float64) {
	if angle != 0.0 {
		mx.queue = append(mx.queue, []float64{1.0, 0.0, math.Tan(angle * math.Pi / 180.0), 1.0, 0.0, 0.0})
		mx.cache = nil
	}
}

func (mx *Matrix) SkewY(angle float64) {
	if angle != 0.0 {
		mx.queue = append(mx.queue, []float64{1.0, math.Tan(angle * math.Pi / 180.0), 0.0, 1.0, 0.0, 0.0})
		mx.cache = nil
	}
}

// Flatten queue
//
func (mx *Matrix) ToArray() []float64 {
	if mx.cache != nil {
		return mx.cache
	}
	if len(mx.queue) == 0 {
		mx.cache = []float64{1.0, 0.0, 0.0, 1.0, 0.0, 0.0}
		return mx.cache
	}

	mx.cache = mx.queue[0]

	if len(mx.queue) == 1 {
		return mx.cache
	}

	for i := 1; i < len(mx.queue); i++ {
		mx.cache = combine(mx.cache, mx.queue[i])
	}

	return mx.cache
}

// Apply list of matrixes to (x,y) point.
// If `isRelative` set, `translate` component of matrix will be skipped
//
func (mx *Matrix) Calc(x float64, y float64, isRelative bool) []float64 {

	// Don't change point on empty transforms queue
	if len(mx.queue) == 0 {
		return []float64{x, y}
	}

	// Calculate final matrix, if not exists
	//
	// NB. if you deside to apply transforms to point one-by-one,
	// they should be taken in reverse order

	if mx.cache == nil {
		mx.cache = mx.ToArray()
	}

	m := mx.cache

	// Apply matrix to point
	if isRelative {
		return []float64{
			x*m[0] + y*m[2],
			x*m[1] + y*m[3],
		}
	}
	return []float64{
		x*m[0] + y*m[2] + m[4],
		x*m[1] + y*m[3] + m[5],
	}
}
