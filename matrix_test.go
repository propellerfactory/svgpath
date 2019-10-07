// port of https://github.com/fontello/svgpath
package svgpath

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMatrixNops(t *testing.T) {
	m := NewMatrix()

	m.Matrix([]float64{1, 0, 0, 1, 0, 0})
	assert.Equal(t, 0, len(m.queue))

	m.Translate(0, 0)
	assert.Equal(t, 0, len(m.queue))

	m.Scale(1, 1)
	assert.Equal(t, 0, len(m.queue))

	m.Rotate(0, 0, 0)
	assert.Equal(t, 0, len(m.queue))

	m.SkewX(0)
	assert.Equal(t, 0, len(m.queue))

	m.SkewY(0)
	assert.Equal(t, 0, len(m.queue))
}

func TestMatrixEmptyQueue(t *testing.T) {
	m := NewMatrix()
	assert.EqualValues(t, []float64{10, 11}, m.Calc(10, 11, false))
	assert.EqualValues(t, []float64{1, 0, 0, 1, 0, 0}, m.ToArray())
}

func TestMatrixCompose(t *testing.T) {
	m := NewMatrix()
	m.Translate(10, 10)
	m.Translate(-10, -10)
	m.Rotate(180, 10, 10)
	m.Rotate(180, 10, 10)
	results := m.ToArray()

	// Need to round errors prior to compare
	assert.Equal(t, 100.0, math.Round(results[0]*100))
	assert.Equal(t, 0.0, math.Round(results[1]*100))
	assert.Equal(t, 0.0, math.Round(results[2]*100))
	assert.Equal(t, 100.0, math.Round(results[3]*100))
	assert.Equal(t, 0.0, math.Round(results[4]*100))
	assert.Equal(t, 0.0, math.Round(results[5]*100))
}

func TestMatrixCache(t *testing.T) {
	m := NewMatrix()
	m.Translate(10, 20)
	m.Scale(2, 3)

	assert.Nil(t, m.cache)
	assert.EqualValues(t, []float64{2, 0, 0, 3, 10, 20}, m.ToArray())
	assert.EqualValues(t, []float64{2, 0, 0, 3, 10, 20}, m.cache)
	m.cache = []float64{1, 2, 3, 4, 5, 6}
	assert.EqualValues(t, []float64{1, 2, 3, 4, 5, 6}, m.ToArray())
}
