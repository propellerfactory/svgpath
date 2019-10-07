// port of https://github.com/fontello/svgpath
package svgpath

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTransformTranslate(t *testing.T) {
	sp, err := NewSvgPath("M10 10 L15 15")
	assert.Nil(t, err)
	sp.Transform("translate(20)")
	assert.Equal(t, "M30 10L35 15", sp.ToString(), "translate x only")

	sp, err = NewSvgPath("M10 10 L15 15")
	assert.Nil(t, err)
	sp.Transform("translate(20,10)")
	assert.Equal(t, "M30 20L35 25", sp.ToString(), "translate x and y")

	sp, err = NewSvgPath("M10 10 c15 15, 20 10, 15 15")
	assert.Nil(t, err)
	sp.Transform("translate(20,10)")
	assert.Equal(t, "M30 20c15 15 20 10 15 15", sp.ToString(), "translate x and y with relatives curves")

	sp, err = NewSvgPath("M10 10 C15 15, 20 10, 15 15")
	assert.Nil(t, err)
	sp.Transform("translate(20,10)")
	assert.Equal(t, "M30 20C35 25 40 20 35 25", sp.ToString(), "translate x and y with absolute curves")

	p := "m70 70 l20 20 l-20 0 l0 -20"
	sp, err = NewSvgPath(p)
	assert.Nil(t, err)
	sp.Transform("translate(100,100)")
	assert.Equal(t, "M170 170l20 20-20 0 0-20", sp.ToString(), "translate rel after translate sequence should not break translate if first m (#10)")

	sp, err = NewSvgPath(p)
	assert.Nil(t, err)
	sp.Transform("translate(100,100)")
	sp.Rel()
	assert.Equal(t, "M170 170l20 20-20 0 0-20", sp.ToString(), "translate rel after translate sequence should not break translate if first m (#10)")
}

func TestTransformRotate(t *testing.T) {
	sp, err := NewSvgPath("M10 10L15 10")
	assert.Nil(t, err)
	sp.Transform("rotate(90, 10, 10)")
	sp.Round(0)
	assert.Equal(t, "M10 10L10 15", sp.ToString(), "rotate by 90 degrees about point(10, 10)")

	sp, err = NewSvgPath("M0 10L0 20")
	assert.Nil(t, err)
	sp.Transform("rotate(-90)")
	sp.Round(0)
	assert.Equal(t, "M10 0L20 0", sp.ToString(), "rotate by -90 degrees about point (0,0)")
}

func TestTransformScale(t *testing.T) {
	sp, err := NewSvgPath("M5 5L15 20")
	assert.Nil(t, err)
	sp.Transform("scale(2)")
	assert.Equal(t, "M10 10L30 40", sp.ToString(), "scale picture by 2")

	sp, err = NewSvgPath("M5 5L30 20")
	assert.Nil(t, err)
	sp.Transform("scale(.5, 1.5)")
	assert.Equal(t, "M2.5 7.5L15 30", sp.ToString(), "scale picture with x*0.5 and y*1.5")

	sp, err = NewSvgPath("M5 5c15 15, 20 10, 15 15")
	assert.Nil(t, err)
	sp.Transform("scale(.5, 1.5)")
	assert.Equal(t, "M2.5 7.5c7.5 22.5 10 15 7.5 22.5", sp.ToString(), "scale picture with x*0.5 and y*1.5 with relative elements")
}

func TestTransformSkew(t *testing.T) {
	// SkewX matrix [ 1, 0, 4, 1, 0, 0 ],
	// x = x*1 + y*4 + 0 = x + y*4
	// y = x*0 + y*1 + 0 = y
	sp, err := NewSvgPath("M5 5L15 20")
	assert.Nil(t, err)
	sp.Transform("skewX(75.96)")
	sp.Round(0)
	assert.Equal(t, "M25 5L95 20", sp.ToString(), "skewX")

	// SkewY matrix [ 1, 4, 0, 1, 0, 0 ],
	// x = x*1 + y*0 + 0 = x
	// y = x*4 + y*1 + 0 = y + x*4
	sp, err = NewSvgPath("M5 5L15 20")
	assert.Nil(t, err)
	sp.Transform("skewY(75.96)")
	sp.Round(0)
	assert.Equal(t, "M5 25L15 80", sp.ToString(), "skewY")
}

func TestTransformMatrix(t *testing.T) {
	// x = x*1.5 + y/2 + ( absolute ? 10 : 0)
	// y = x/2 + y*1.5 + ( absolute ? 15 : 0)
	sp, err := NewSvgPath("M5 5 C20 30 10 15 30 15")
	assert.Nil(t, err)
	sp.Transform("matrix(1.5, 0.5, 0.5, 1.5 10, 15)")
	assert.Equal(t, "M20 25C55 70 32.5 42.5 62.5 52.5", sp.ToString(), "path with absolute segments")

	// SkewY matrix [ 1, 4, 0, 1, 0, 0 ],
	// x = x*1 + y*0 + 0 = x
	// y = x*4 + y*1 + 0 = y + x*4
	sp, err = NewSvgPath("M5 5 c10 12 10 15 20 30")
	assert.Nil(t, err)
	sp.Transform("matrix(1.5, 0.5, 0.5, 1.5 10, 15)")
	assert.Equal(t, "M20 25c21 23 22.5 27.5 45 55", sp.ToString(), "path with relative segments")
}

func TestTransformCombinations(t *testing.T) {
	sp, err := NewSvgPath("M0 0 L 10 10 20 10")
	assert.Nil(t, err)
	sp.Transform("translate(100,100) scale(2,3)")
	assert.Equal(t, "M100 100L120 130 140 130", sp.ToString(), "scale + translate")

	sp, err = NewSvgPath("M0 0 L 10 10 20 10")
	assert.Nil(t, err)
	sp.Transform("rotate(90) scale(2,3)")
	sp.Round(0)
	assert.Equal(t, "M0 0L-30 20-30 40", sp.ToString(), "scale + rotate")

	sp, err = NewSvgPath("M0 0 L 10 10 20 10")
	assert.Nil(t, err)
	sp.Transform("skewX(75.96) scale(2,3)")
	sp.Round(0)
	assert.Equal(t, "M0 0L140 30 160 30", sp.ToString(), "rotate + skewX")
}

func TestTransformMisc(t *testing.T) {
	sp, err := NewSvgPath("M0 0 L 10 10 20 10")
	assert.Nil(t, err)
	sp.Transform("rotate(0) scale(1,1) translate(0,0) skewX(0) skewY(0)")
	sp.Round(0)
	assert.Equal(t, "M0 0L10 10 20 10", sp.ToString(), "empty transforms")

	sp, err = NewSvgPath("M0 0 L 10 10 20 10")
	assert.Nil(t, err)
	sp.Transform("rotate(10,0) scale(10,10,1) translate(10,10,0) skewX(10,0) skewY(10,0) matrix(0)")
	sp.Round(0)
	assert.Equal(t, "M0 0L10 10 20 10", sp.ToString(), "wrong params count in transforms")

	sp, err = NewSvgPath("M0 0 H 10 V 10 Z M 100 100 h 15 v -10")
	assert.Nil(t, err)
	sp.Transform("rotate(45)")
	sp.Round(0)
	assert.Equal(t, "M0 0L7 7 0 14ZM0 141l11 11 7-7", sp.ToString(), "segment replacement [H,V] => L")

	sp, err = NewSvgPath("M10 10 L15 15")
	assert.Nil(t, err)
	sp.Transform("    ")
	sp.Round(0)
	assert.Equal(t, "M10 10L15 15", sp.ToString(), "nothing to transform")

	sp, err = NewSvgPath("m70 70 70 70")
	// By default parser force first 'm' to upper case
	// and we don't fall into troubles.
	assert.Nil(t, err)
	sp.Translate(100, 100)
	assert.Equal(t, "M170 170l70 70", sp.ToString(), "first m should be processed as absolute")

	// Emulate first 'm'.
	sp, err = NewSvgPath("m70 70 70 70")
	sp.segments[0][0] = "m"

	// By default parser force first 'm' to upper case
	// and we don't fall into troubles.
	assert.Nil(t, err)
	sp.Translate(100, 100)
	assert.Equal(t, "m170 170l70 70", sp.ToString(), "first m should be processed as absolute")

}
