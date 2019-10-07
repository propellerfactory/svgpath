// port of https://github.com/fontello/svgpath
package svgpath

import (
	"testing"

	"github.com/stretchr/testify/assert"
	//"github.com/stretchr/testify/assert"
)

func TestEmptyString(t *testing.T) {
	sp, err := NewSvgPath("")
	assert.Nil(t, err)
	assert.Equal(t, sp.ToString(), "")
}

func TestLineTerminators(t *testing.T) {
	sp, err := NewSvgPath("M0\r 0\n\u1680l2-3\nz")
	assert.Nil(t, err)
	assert.Equal(t, "M0 0l2-3z", sp.ToString())
}

func TestParamsFormats(t *testing.T) {
	sp, err := NewSvgPath("M 0.0 0.0")
	assert.Nil(t, err)
	assert.Equal(t, "M0 0", sp.ToString())
	sp, err = NewSvgPath("M 1e2 0")
	assert.Nil(t, err)
	assert.Equal(t, "M100 0", sp.ToString())
	sp, err = NewSvgPath("M 1e+2 0")
	assert.Nil(t, err)
	assert.Equal(t, "M100 0", sp.ToString())
	sp, err = NewSvgPath("M +1e2 0")
	assert.Nil(t, err)
	assert.Equal(t, "M100 0", sp.ToString())
	sp, err = NewSvgPath("M 1e-2 0")
	assert.Nil(t, err)
	assert.Equal(t, "M0.01 0", sp.ToString())
	sp, err = NewSvgPath("M 0.1e-2 0")
	assert.Nil(t, err)
	assert.Equal(t, "M0.001 0", sp.ToString())
	sp, err = NewSvgPath("M .1e-2 0")
	assert.Nil(t, err)
	assert.Equal(t, "M0.001 0", sp.ToString())
}

func TestRepeated(t *testing.T) {
	sp, err := NewSvgPath("M 0 0 100 100")
	assert.Nil(t, err)
	assert.Equal(t, "M0 0L100 100", sp.ToString())
	sp, err = NewSvgPath("m 0 0 100 100")
	assert.Nil(t, err)
	assert.Equal(t, "M0 0l100 100", sp.ToString())
	sp, err = NewSvgPath("M 0 0 R 1 1 2 2")
	assert.Nil(t, err)
	assert.Equal(t, "M0 0R1 1 2 2", sp.ToString())
	sp, err = NewSvgPath("M 0 0 r 1 1 2 2")
	assert.Nil(t, err)
	assert.Equal(t, "M0 0r1 1 2 2", sp.ToString())
}

func TestErrors(t *testing.T) {
	_, err := NewSvgPath("0")
	assert.Equal(t, "SvgPath: bad command 0 (at pos 0)", err.Error())
	_, err = NewSvgPath("U")
	assert.Equal(t, "SvgPath: bad command U (at pos 0)", err.Error())
	_, err = NewSvgPath("M0 0G 1")
	assert.Equal(t, "SvgPath: bad command G (at pos 4)", err.Error())
	_, err = NewSvgPath("z")
	assert.Equal(t, "SvgPath: string should start with `M` or `m`", err.Error())
	_, err = NewSvgPath("M+")
	assert.Equal(t, "SvgPath: param should start with 0..9 or `.` (at pos 2)", err.Error())
	_, err = NewSvgPath("M00")
	assert.Equal(t, "SvgPath: numbers started with `0` such as `09` are ilegal (at pos 1)", err.Error())
	_, err = NewSvgPath("M0e")
	assert.Equal(t, "SvgPath: invalid float exponent (at pos 3)", err.Error())
	_, err = NewSvgPath("M0")
	assert.Equal(t, "SvgPath: missed param (at pos 2)", err.Error())
	_, err = NewSvgPath("M0,0,")
	assert.Equal(t, "SvgPath: missed param (at pos 5)", err.Error())
	_, err = NewSvgPath("M0 .e3")
	assert.Equal(t, "SvgPath: invalid float exponent (at pos 4)", err.Error())
}
