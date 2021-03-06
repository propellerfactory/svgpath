// port of https://github.com/fontello/svgpath
package svgpath

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToString(t *testing.T) {
	sp, err := NewSvgPath("M 10 10 M 10 100 M 100 100 M 100 10 Z")
	assert.Nil(t, err)
	assert.Equal(t, "M10 10M10 100M100 100M100 10Z", sp.ToString(), "should not collapse multiple M")

	sp, err = NewSvgPath("m 10 10 m 10 100 m 100 100 m 100 10 z")
	assert.Nil(t, err)
	assert.Equal(t, "M10 10m10 100m100 100m100 10z", sp.ToString(), "should not collapse multiple m")
}

func TestUnshort(t *testing.T) {
	sp, err := NewSvgPath("M10 10 C 20 20, 40 20, 50 10")
	assert.Nil(t, err)
	sp.Unshort()
	assert.Equal(t, "M10 10C20 20 40 20 50 10", sp.ToString(), "shouldn't change full arc")

	sp, err = NewSvgPath("M10 10 C 20 20, 40 20, 50 10 S 80 0, 90 10")
	assert.Nil(t, err)
	sp.Unshort()
	assert.Equal(t, "M10 10C20 20 40 20 50 10 60 0 80 0 90 10", sp.ToString(), "should reflect control point after full path")

	sp, err = NewSvgPath("M10 10 S 50 50, 90 10")
	assert.Nil(t, err)
	sp.Unshort()
	assert.Equal(t, "M10 10C10 10 50 50 90 10", sp.ToString(), "should copy starting point if not followed by a path")

	sp, err = NewSvgPath("M30 50 c 10 30, 30 30, 40 0 s 30 -30, 40 0")
	assert.Nil(t, err)
	sp.Unshort()
	assert.Equal(t, "M30 50c10 30 30 30 40 0 10-30 30-30 40 0", sp.ToString(), "should handle relative paths")
}

func TestUnshortQuadratic(t *testing.T) {
	sp, err := NewSvgPath("M10 10 Q 50 50, 90 10")
	assert.Nil(t, err)
	sp.Unshort()
	assert.Equal(t, "M10 10Q50 50 90 10", sp.ToString(), "shouldn't change full arc")

	sp, err = NewSvgPath("M30 50 Q 50 90, 90 50 T 150 50")
	assert.Nil(t, err)
	sp.Unshort()
	assert.Equal(t, "M30 50Q50 90 90 50 130 10 150 50", sp.ToString(), "should reflect control point after full path")

	sp, err = NewSvgPath("M10 30 T150 50")
	assert.Nil(t, err)
	sp.Unshort()
	assert.Equal(t, "M10 30Q10 30 150 50", sp.ToString(), "should copy starting point if not followed by a path")

	sp, err = NewSvgPath("M30 50 q 20 20, 40 0 t 40 0")
	assert.Nil(t, err)
	sp.Unshort()
	assert.Equal(t, "M30 50q20 20 40 0 20-20 40 0", sp.ToString(), "should handle relative paths")
}

func TestAbs(t *testing.T) {
	sp, err := NewSvgPath("M10 10 l 30 30")
	assert.Nil(t, err)
	sp.Abs()
	assert.Equal(t, "M10 10L40 40", sp.ToString(), "should convert line")

	sp, err = NewSvgPath("M10 10 L30 30")
	assert.Nil(t, err)
	sp.Abs()
	assert.Equal(t, "M10 10L30 30", sp.ToString(), "shouldn't process existing line")

	sp, err = NewSvgPath("M10 10 c 10 30 30 30 40, 0 10 -30 20 -30 40 0")
	assert.Nil(t, err)
	sp.Abs()
	assert.Equal(t, "M10 10C20 40 40 40 50 10 60-20 70-20 90 10", sp.ToString(), "should convert multi-segment curve")

	sp, err = NewSvgPath("M10 10H40h50")
	assert.Nil(t, err)
	sp.Abs()
	assert.Equal(t, "M10 10H40 90", sp.ToString(), "should handle horizontal lines")

	sp, err = NewSvgPath("M10 10V40v50")
	assert.Nil(t, err)
	sp.Abs()
	assert.Equal(t, "M10 10V40 90", sp.ToString(), "should handle vertical line")

	sp, err = NewSvgPath("M40 30a20 40 -45 0 1 20 50")
	assert.Nil(t, err)
	sp.Abs()
	assert.Equal(t, "M40 30A20 40-45 0 1 60 80", sp.ToString(), "should handle arcs")

	sp, err = NewSvgPath("M10 10 l10 0 l0 10 Z l 0 10 l 10 0 z l-1-1")
	assert.Nil(t, err)
	sp.Abs()
	assert.Equal(t, "M10 10L20 10 20 20ZL10 20 20 20ZL9 9", sp.ToString(), "should track position after z")
}

func TestRel(t *testing.T) {
	sp, err := NewSvgPath("M10 10 L30 30")
	assert.Nil(t, err)
	sp.Rel()
	assert.Equal(t, "M10 10l20 20", sp.ToString(), "should convert line")

	sp, err = NewSvgPath("m10 10 l30 30")
	assert.Nil(t, err)
	sp.Rel()
	assert.Equal(t, "M10 10l30 30", sp.ToString(), "shouldn't process existing line")

	sp, err = NewSvgPath("M10 10 C 20 40 40 40 50 10 60 -20 70 -20 90 10")
	assert.Nil(t, err)
	sp.Rel()
	assert.Equal(t, "M10 10c10 30 30 30 40 0 10-30 20-30 40 0", sp.ToString(), "should convert multi-segment curve")

	sp, err = NewSvgPath("M10 10H40h50")
	assert.Nil(t, err)
	sp.Rel()
	assert.Equal(t, "M10 10h30 50", sp.ToString(), "should handle horizontal lines")

	sp, err = NewSvgPath("M10 10V40v50")
	assert.Nil(t, err)
	sp.Rel()
	assert.Equal(t, "M10 10v30 50", sp.ToString(), "should handle vertical lines")

	sp, err = NewSvgPath("M40 30A20 40 -45 0 1 60 80")
	assert.Nil(t, err)
	sp.Rel()
	assert.Equal(t, "M40 30a20 40-45 0 1 20 50", sp.ToString(), "should handle arcs")

	sp, err = NewSvgPath("M10 10 L20 10 L20 20 Z L10 20 L20 20 z L9 9")
	assert.Nil(t, err)
	sp.Rel()
	assert.Equal(t, "M10 10l10 0 0 10zl0 10 10 0zl-1-1", sp.ToString(), "should track position after z")
}

func TestScale(t *testing.T) {
	sp, err := NewSvgPath("M10 10 C 20 40 40 40 50 10")
	assert.Nil(t, err)
	sp.Scale(2, 1.5)
	assert.Equal(t, "M20 15C40 60 80 60 100 15", sp.ToString(), "should scale abs curve")

	sp, err = NewSvgPath("M10 10 c 10 30 30 30 40 0")
	assert.Nil(t, err)
	sp.Scale(2, 1.5)
	assert.Equal(t, "M20 15c20 45 60 45 80 0", sp.ToString(), "should scale rel curve")

	/* not supported in golang API (by choice during port)
	   it('second argument defaults to the first', function () {
	     assert.equal(
	       svgpath('M10 10l20 30').scale(2).toString(),
	       'M20 20l40 60'
	     );
	   });
	*/

	sp, err = NewSvgPath("M10 10H40h50")
	assert.Nil(t, err)
	sp.Scale(2, 1.5)
	assert.Equal(t, "M20 15H80h100", sp.ToString(), "should handle horizontal lines")

	sp, err = NewSvgPath("M10 10V40v50")
	assert.Nil(t, err)
	sp.Scale(2, 1.5)
	assert.Equal(t, "M20 15V60v75", sp.ToString(), "should handle vertical lines")

	sp, err = NewSvgPath("M40 30a20 40 -45 0 1 20 50")
	assert.Nil(t, err)
	sp.Scale(2, 1.5)
	sp.Round(0)
	assert.Equal(t, "M80 45a72 34 32.04 0 1 40 75", sp.ToString(), "should handle arcs")

	sp, err = NewSvgPath("M40 30A20 40 -45 0 1 20 50")
	assert.Nil(t, err)
	sp.Scale(2, 1.5)
	sp.Round(0)
	assert.Equal(t, "M80 45A72 34 32.04 0 1 40 75", sp.ToString(), "should handle horizontal lines")
}

func TestRotate(t *testing.T) {
	sp, err := NewSvgPath("M10 10L15 10")
	assert.Nil(t, err)
	sp.Rotate(90, 10, 10)
	sp.Round(0)
	assert.Equal(t, "M10 10L10 15", sp.ToString(), "otate by 90 degrees about point(10, 10)")
	/*
			it('rotate by -90 degrees about point (0,0)', function () {
		      assert.equal(
		        svgpath('M0 10L0 20').rotate(-90).round(0).toString(),
		        'M10 0L20 0'
		      );
		    });

		    it('rotate abs arc', function () {
		      assert.equal(
		        svgpath('M 100 100 A 90 30 0 1 1 200 200').rotate(45).round(0).toString(),
		        'M0 141A90 30 45 1 1 0 283'
		      );
		    });

		    it('rotate rel arc', function () {
		      assert.equal(
		        svgpath('M 100 100 a 90 30 15 1 1 200 200').rotate(20).round(0).toString(),
		        'M60 128a90 30 35 1 1 119 257'
		      );
		    });
		  });
	*/
}

func TestSkew(t *testing.T) {
	/*

	   describe('skew', function () {
	     // SkewX matrix [ 1, 0, 4, 1, 0, 0 ],
	     // x = x*1 + y*4 + 0 = x + y*4
	     // y = x*0 + y*1 + 0 = y
	     it('skewX', function () {
	       assert.equal(
	         svgpath('M5 5L15 20').skewX(75.96).round(0).toString(),
	         'M25 5L95 20'
	       );
	     });

	     // SkewY matrix [ 1, 4, 0, 1, 0, 0 ],
	     // x = x*1 + y*0 + 0 = x
	     // y = x*4 + y*1 + 0 = y + x*4
	     it('skewY', function () {
	       assert.equal(
	         svgpath('M5 5L15 20').skewY(75.96).round(0).toString(),
	         'M5 25L15 80'
	       );
	     });
	   });
	*/
}

func TestMatrix(t *testing.T) {
	/*
	   describe('matrix', function () {
	     // x = x*1.5 + y/2 + ( absolute ? 10 : 0)
	     // y = x/2 + y*1.5 + ( absolute ? 15 : 0)
	     it('path with absolute segments', function () {
	       assert.equal(
	         svgpath('M5 5 C20 30 10 15 30 15').matrix([ 1.5, 0.5, 0.5, 1.5, 10, 15 ]).toString(),
	         'M20 25C55 70 32.5 42.5 62.5 52.5'
	       );
	     });

	     it('path with relative segments', function () {
	       assert.equal(
	         svgpath('M5 5 c10 12 10 15 20 30').matrix([ 1.5, 0.5, 0.5, 1.5, 10, 15 ]).toString(),
	         'M20 25c21 23 22.5 27.5 45 55'
	       );
	     });

	     it('no change', function () {
	       assert.equal(
	         svgpath('M5 5 C20 30 10 15 30 15').matrix([ 1, 0, 0, 1, 0, 0 ]).toString(),
	         'M5 5C20 30 10 15 30 15'
	       );
	     });

	     it('should handle arcs', function () {
	       assert.equal(
	         svgpath('M40 30a20 40 -45 0 1 20 50').matrix([ 1.5, 0.5, 0.5, 1.5, 10, 15 ]).round(0).toString(),
	         'M85 80a80 20 45 0 1 55 85'
	       );

	       assert.equal(
	         svgpath('M40 30A20 40 -45 0 1 20 50').matrix([ 1.5, 0.5, 0.5, 1.5, 10, 15 ]).round(0).toString(),
	         'M85 80A80 20 45 0 1 65 100'
	       );
	     });
	   });
	*/
}

func TestCombinations(t *testing.T) {
	/*
	   describe('combinations', function () {
	     it('scale + translate', function () {
	       assert.equal(
	         svgpath('M0 0 L 10 10 20 10').scale(2, 3).translate(100, 100).toString(),
	         'M100 100L120 130 140 130'
	       );
	     });

	     it('scale + rotate', function () {
	       assert.equal(
	         svgpath('M0 0 L 10 10 20 10').scale(2, 3).rotate(90).round(0).toString(),
	         'M0 0L-30 20-30 40'
	       );
	     });

	     it('empty', function () {
	       assert.equal(
	         svgpath('M0 0 L 10 10 20 10').translate(0).scale(1).rotate(0, 10, 10).round(0).toString(),
	         'M0 0L10 10 20 10'
	       );
	     });
	   });
	*/
}

func TestTranslate(t *testing.T) {
	/*
	   describe('translate', function () {
	     it('should translate abs curve', function () {
	       assert.equal(
	         svgpath('M10 10 C 20 40 40 40 50 10').translate(5, 15).toString(),
	         'M15 25C25 55 45 55 55 25'
	       );
	     });

	     it('should translate rel curve', function () {
	       assert.equal(
	         svgpath('M10 10 c 10 30 30 30 40 0').translate(5, 15).toString(),
	         'M15 25c10 30 30 30 40 0'
	       );
	     });

	     it('second argument defaults to zero', function () {
	       assert.equal(
	         svgpath('M10 10L20 30').translate(10).toString(),
	         'M20 10L30 30'
	       );
	     });

	     it('should handle horizontal lines', function () {
	       assert.equal(
	         svgpath('M10 10H40h50').translate(10, 15).toString(),
	         'M20 25H50h50'
	       );
	     });

	     it('should handle vertical lines', function () {
	       assert.equal(
	         svgpath('M10 10V40v50').translate(10, 15).toString(),
	         'M20 25V55v50'
	       );
	     });

	     it('should handle arcs', function () {
	       assert.equal(
	         svgpath('M40 30a20 40 -45 0 1 20 50').translate(10, 15).round(0).toString(),
	         'M50 45a40 20 45 0 1 20 50'
	       );

	       assert.equal(
	         svgpath('M40 30A20 40 -45 0 1 20 50').translate(10, 15).round(0).toString(),
	         'M50 45A40 20 45 0 1 30 65'
	       );
	     });
	   });
	*/
}

func TestRound(t *testing.T) {

	/*

	   describe('round', function () {
	     it('should round arcs', function () {
	       assert.equal(
	         svgpath('M10 10 A12.5 17.5 45.5 0 0 15.5 19.5').round(0).toString(),
	         'M10 10A13 18 45.5 0 0 16 20'
	       );
	     });

	     it('should round curves', function () {
	       assert.equal(
	         svgpath('M10 10 c 10.12 30.34 30.56 30 40.00 0.12').round(0).toString(),
	         'M10 10c10 30 31 30 40 0'
	       );
	     });

	     it('set precision', function () {
	       assert.equal(
	         svgpath('M10.123 10.456L20.4351 30.0000').round(2).toString(),
	         'M10.12 10.46L20.44 30'
	       );
	     });

	     it('should track errors', function () {
	       assert.equal(
	         svgpath('M1.2 1.4l1.2 1.4 l1.2 1.4').round(0).toString(),
	         'M1 1l1 2 2 1'
	       );
	     });

	     it('should track errors #2', function () {
	       assert.equal(
	         svgpath('M1.2 1.4 H2.4 h1.2 v2.4 h-2.4 V2.4 v-1.2').round(0).toString(),
	         'M1 1H2h2v3h-3V2v-1'
	       );
	     });

	     it('should track errors for contour start', function () {
	       assert.equal(
	         svgpath('m0.4 0.2zm0.4 0.2m0.4 0.2m0.4 0.2zm0.4 0.2').round(0).abs().toString(),
	         'M0 0ZM1 0M1 1M2 1ZM2 1'
	       );
	     });

	     it('reset delta error on contour end', function () {
	       assert.equal(
	         svgpath('m.1 .1l.3 .3zm.1 .1l.3 .3zm0 0z').round(0).abs().toString(),
	         'M0 0L0 0ZM0 0L1 1ZM0 0Z'
	       );
	     });
	   });
	*/
}

func TestUnarc(t *testing.T) {
	sp, err := NewSvgPath("M100 100 A30 50 0 1 1 110 110")
	assert.Nil(t, err)
	sp.Unarc()
	sp.Round(0)
	assert.Equal(t, "M100 100C89 83 87 54 96 33 105 12 122 7 136 20 149 33 154 61 147 84 141 108 125 119 110 110", sp.ToString(), "almost complete arc gets expanded to 4 curves")

	/*
	   describe('unarc', function () {
	     it('', function () {
	       assert.equal(
	         svgpath('').unarc().round(0).toString(),
	         ''
	       );
	     });

	     it('small arc gets expanded to one curve', function () {
	       assert.equal(
	         svgpath('M100 100 a30 50 0 0 1 30 30').unarc().round(0).toString(),
	         'M100 100C113 98 125 110 130 130'
	       );
	     });

	     it('unarc a circle', function () {
	       assert.equal(
	         svgpath('M 100, 100 m -75, 0 a 75,75 0 1,0 150,0 a 75,75 0 1,0 -150,0').unarc().round(0).toString(),
	         'M100 100m-75 0C25 141 59 175 100 175 141 175 175 141 175 100 175 59 141 25 100 25 59 25 25 59 25 100'
	       );
	     });

	     it('rounding errors', function () {
	       // Coverage
	       //
	       // Due to rounding errors, with these exact arguments radicant
	       // will be -9.974659986866641e-17, causing Math.sqrt() of that to be NaN
	       //
	       assert.equal(
	         svgpath('M-0.5 0 A 0.09188163040671497 0.011583783896639943 0 0 1 0 0.5').unarc().round(5).toString(),
	         'M-0.5 0C0.59517-0.01741 1.59491 0.08041 1.73298 0.21848 1.87105 0.35655 1.09517 0.48259 0 0.5'
	       );
	     });

	     it('rounding errors #2', function () {
	       // Coverage
	       //
	       // Due to rounding errors this will compute Math.acos(-1.0000000000000002)
	       // and fail when calculating vector between angles
	       //
	       assert.equal(
	         svgpath('M-0.07467194809578359 -0.3862391309812665' +
	             'A1.2618792965076864 0.2013618852943182 90 0 1 -0.7558937461581081 -0.8010219619609416')
	           .unarc().round(5).toString(),

	         'M-0.07467-0.38624C-0.09295 0.79262-0.26026 1.65542-0.44838 1.54088' +
	         '-0.63649 1.42634-0.77417 0.37784-0.75589-0.80102'
	       );
	     });

	     it("we're already there", function () {
	       // Asked to draw a curve between a point and itself. According to spec,
	       // nothing shall be drawn in this case.
	       //
	       assert.equal(
	         svgpath('M100 100A123 456 90 0 1 100 100').unarc().round(0).toString(),
	         'M100 100L100 100'
	       );

	       assert.equal(
	         svgpath('M100 100a123 456 90 0 1 0 0').unarc().round(0).toString(),
	         'M100 100l0 0'
	       );
	     });

	     it('radii are zero', function () {
	       // both rx and ry are zero
	       assert.equal(
	         svgpath('M100 100A0 0 0 0 1 110 110').unarc().round(0).toString(),
	         'M100 100L110 110'
	       );

	       // rx is zero
	       assert.equal(
	         svgpath('M100 100A0 100 0 0 1 110 110').unarc().round(0).toString(),
	         'M100 100L110 110'
	       );
	     });
	   });

	*/
}

func TestUncubic(t *testing.T) {
	sp, err := NewSvgPath("M81.016,63.155c-0.992-0.004-1.838,0.78-1.868,1.787l-0.006,0.143c-0.1,1.4-0.728,3.061-3.145,3.061   c-0.982,0-2.861-2.336-4.233-4.041c-2.625-3.262-5.6-6.959-9.767-6.959c-5.1,0-10.089,1.966-12.006,2.814   c-1.917-0.849-6.906-2.814-12.006-2.814c-4.167,0-7.142,3.697-9.766,6.959c-1.372,1.705-3.251,4.041-4.234,4.041   c-2.417,0-3.045-1.66-3.145-3.062l-0.006-0.139c-0.027-1.003-0.848-1.791-1.849-1.791c-0.005,0-0.011,0-0.017,0   c-1.008,0.009-1.824,0.833-1.834,1.841c0,0-0.001,0.11,0.012,0.306c0.109,2.079,1.011,12.356,8.108,14.479   c2.467,0.738,4.944,1.112,7.363,1.112c6.136,0,11.648-2.396,15.123-6.572c0.951-1.143,1.683-2.166,2.251-3.077   c0.568,0.911,1.299,1.934,2.25,3.077c3.475,4.177,8.987,6.572,15.124,6.572c2.418,0,4.896-0.374,7.362-1.112   c7.097-2.123,7.998-12.397,8.107-14.479c0.014-0.196,0.012-0.307,0.012-0.307C82.837,63.987,82.022,63.165,81.016,63.155zM51.997,77.146h-4c-1.022,0-1.85,0.828-1.85,1.85s0.828,1.85,1.85,1.85h4c1.021,0,1.85-0.828,1.85-1.85   S53.019,77.146,51.997,77.146zM67.601,23.057c0.287,0,0.578-0.067,0.851-0.209l0.797-0.414c0.907-0.471,1.26-1.588,0.789-2.495   c-0.471-0.906-1.587-1.259-2.494-0.789l-0.797,0.414c-0.907,0.471-1.26,1.588-0.789,2.495   C66.286,22.693,66.932,23.057,67.601,23.057zM30.745,22.433l0.796,0.414c0.273,0.142,0.565,0.209,0.853,0.209c0.668,0,1.313-0.363,1.643-0.997   c0.472-0.907,0.119-2.023-0.787-2.495l-0.796-0.414c-0.908-0.472-2.024-0.119-2.495,0.787   C29.486,20.845,29.839,21.961,30.745,22.433zM26.167,22.891c1.021-0.046,1.81-0.912,1.764-1.933c-0.047-1.021-0.917-1.817-1.933-1.764   c-7.542,0.344-14.493,4.667-18.142,11.281c-0.494,0.895-0.168,2.02,0.726,2.513c0.283,0.157,0.59,0.23,0.892,0.23   c0.652,0,1.284-0.345,1.621-0.957C14.128,26.768,19.903,23.177,26.167,22.891zM72.997,25.146c-7.249,0-13.363,4.898-15.242,11.555c-2.275-1.016-4.748-1.555-7.258-1.555   c-2.864,0-5.639,0.685-8.148,1.977c-1.735-6.87-7.95-11.977-15.352-11.977c-8.74,0-15.85,7.11-15.85,15.85   c0,8.739,7.11,15.849,15.85,15.849c8.702,0,15.784-7.05,15.845-15.738c2.285-1.475,4.918-2.261,7.655-2.261   c2.335,0,4.628,0.586,6.674,1.687c-0.005,0.155-0.023,0.307-0.023,0.463c0,8.739,7.11,15.849,15.85,15.849   s15.85-7.109,15.85-15.849C88.847,32.256,81.736,25.146,72.997,25.146z M26.997,53.146c-6.7,0-12.15-5.45-12.15-12.149   c0-6.7,5.45-12.15,12.15-12.15s12.15,5.45,12.15,12.15C39.147,47.695,33.697,53.146,26.997,53.146z M72.997,50.095   c-5.018,0-9.1-4.082-9.1-9.099c0-5.018,4.082-9.1,9.1-9.1s9.1,4.083,9.1,9.1C82.097,46.013,78.015,50.095,72.997,50.095zM91.991,30.21c-3.707-6.503-10.38-10.618-17.851-11.008c-1.014-0.047-1.891,0.73-1.944,1.751   c-0.053,1.02,0.73,1.891,1.751,1.944c6.206,0.324,11.75,3.742,14.829,9.145c0.342,0.599,0.966,0.934,1.609,0.934   c0.311,0,0.625-0.078,0.914-0.243C92.188,32.227,92.497,31.097,91.991,30.21zM72.997,34.096c-3.805,0-6.9,3.095-6.9,6.9c0,3.804,3.096,6.899,6.9,6.899s6.9-3.095,6.9-6.899   C79.897,37.191,76.802,34.096,72.997,34.096z M72.997,43.096c-1.16,0-2.1-0.939-2.1-2.1s0.939-2.1,2.1-2.1s2.1,0.939,2.1,2.1   S74.157,43.096,72.997,43.096z") //"M100,100 c10,10 50,0 10,-10")
	assert.Nil(t, err)
	sp.Uncubic()
	sp.Round(0)

	/*
	   describe('unarc', function () {
	     it('', function () {
	       assert.equal(
	         svgpath('').unarc().round(0).toString(),
	         ''
	       );
	     });

	     it('small arc gets expanded to one curve', function () {
	       assert.equal(
	         svgpath('M100 100 a30 50 0 0 1 30 30').unarc().round(0).toString(),
	         'M100 100C113 98 125 110 130 130'
	       );
	     });

	     it('unarc a circle', function () {
	       assert.equal(
	         svgpath('M 100, 100 m -75, 0 a 75,75 0 1,0 150,0 a 75,75 0 1,0 -150,0').unarc().round(0).toString(),
	         'M100 100m-75 0C25 141 59 175 100 175 141 175 175 141 175 100 175 59 141 25 100 25 59 25 25 59 25 100'
	       );
	     });

	     it('rounding errors', function () {
	       // Coverage
	       //
	       // Due to rounding errors, with these exact arguments radicant
	       // will be -9.974659986866641e-17, causing Math.sqrt() of that to be NaN
	       //
	       assert.equal(
	         svgpath('M-0.5 0 A 0.09188163040671497 0.011583783896639943 0 0 1 0 0.5').unarc().round(5).toString(),
	         'M-0.5 0C0.59517-0.01741 1.59491 0.08041 1.73298 0.21848 1.87105 0.35655 1.09517 0.48259 0 0.5'
	       );
	     });

	     it('rounding errors #2', function () {
	       // Coverage
	       //
	       // Due to rounding errors this will compute Math.acos(-1.0000000000000002)
	       // and fail when calculating vector between angles
	       //
	       assert.equal(
	         svgpath('M-0.07467194809578359 -0.3862391309812665' +
	             'A1.2618792965076864 0.2013618852943182 90 0 1 -0.7558937461581081 -0.8010219619609416')
	           .unarc().round(5).toString(),

	         'M-0.07467-0.38624C-0.09295 0.79262-0.26026 1.65542-0.44838 1.54088' +
	         '-0.63649 1.42634-0.77417 0.37784-0.75589-0.80102'
	       );
	     });

	     it("we're already there", function () {
	       // Asked to draw a curve between a point and itself. According to spec,
	       // nothing shall be drawn in this case.
	       //
	       assert.equal(
	         svgpath('M100 100A123 456 90 0 1 100 100').unarc().round(0).toString(),
	         'M100 100L100 100'
	       );

	       assert.equal(
	         svgpath('M100 100a123 456 90 0 1 0 0').unarc().round(0).toString(),
	         'M100 100l0 0'
	       );
	     });

	     it('radii are zero', function () {
	       // both rx and ry are zero
	       assert.equal(
	         svgpath('M100 100A0 0 0 0 1 110 110').unarc().round(0).toString(),
	         'M100 100L110 110'
	       );

	       // rx is zero
	       assert.equal(
	         svgpath('M100 100A0 100 0 0 1 110 110').unarc().round(0).toString(),
	         'M100 100L110 110'
	       );
	     });
	   });

	*/
}

func TestArcTransformEdgeCases(t *testing.T) {

	/*
	     describe('arc transform edge cases', function () {
	       it('replace arcs rx/ry = 0 with lines', function () {
	         assert.equal(
	           svgpath('M40 30a0 40 -45 0 1 20 50Z M40 30A20 0 -45 0 1 20 50Z').scale(2, 2).toString(),
	           'M80 60l40 100ZM80 60L40 100Z'
	         );
	       });

	       it('drop arcs with end point === start point', function () {
	         assert.equal(
	           svgpath('M40 30a20 40 -45 0 1 0 0').scale(2, 2).toString(),
	           'M80 60l0 0'
	         );

	         assert.equal(
	           svgpath('M40 30A20 40 -45 0 1 40 30').scale(2, 2).toString(),
	           'M80 60L80 60'
	         );
	       });

	       it('to line at scale x|y = 0 ', function () {
	         assert.equal(
	           svgpath('M40 30a20 40 -45 0 1 20 50').scale(0, 1).toString(),
	           'M0 30l0 50'
	         );

	         assert.equal(
	           svgpath('M40 30A20 40 -45 0 1 20 50').scale(1, 0).toString(),
	           'M40 0L20 0'
	         );
	       });

	       it('rotate to +/- 90 degree', function () {
	         assert.equal(
	           svgpath('M40 30a20 40 -45 0 1 20 50').rotate(90).round(0).toString(),
	           'M-30 40a20 40 45 0 1-50 20'
	         );

	         assert.equal(
	           svgpath('M40 30a20 40 -45 0 1 20 50').matrix([ 0, 1, -1, 0, 0, 0 ]).round(0).toString(),
	           'M-30 40a20 40 45 0 1-50 20'
	         );

	         assert.equal(
	           svgpath('M40 30a20 40 -45 0 1 20 50').rotate(-90).round(0).toString(),
	           'M30-40a20 40 45 0 1 50-20'
	         );

	         assert.equal(
	           svgpath('M40 30a20 40 -45 0 1 20 50').matrix([ 0, -1, 1, 0, 0, 0 ]).round(0).toString(),
	           'M30-40a20 40 45 0 1 50-20'
	         );
	       });

	       it('process circle-like segments', function () {
	         assert.equal(
	           svgpath('M50 50A30 30 -45 0 1 100 100').scale(0.5).round(0).toString(),
	           'M25 25A15 15 0 0 1 50 50'
	         );
	       });

	       it('almost zero eigen values', function () {
	         assert.equal(
	           svgpath('M148.7 277.9A228.7 113.2 90 1 0 159.3 734.8').translate(10).round(1).toString(),
	           'M158.7 277.9A228.7 113.2 90 1 0 169.3 734.8'
	         );
	       });

	       it('should flip sweep flag if image is flipped', function () {
	         assert.equal(
	           svgpath('M10 10A20 15 90 0 1 30 10').scale(1, -1).translate(0, 40).toString(),
	           'M10 30A20 15 90 0 0 30 30'
	         );

	         assert.equal(
	           svgpath('M10 10A20 15 90 0 1 30 10').scale(-1, -1).translate(40, 40).toString(),
	           'M30 30A20 15 90 0 1 10 30'
	         );
	       });
	     });
	   });
	*/
}
