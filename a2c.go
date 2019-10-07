// port of https://github.com/fontello/svgpath
package svgpath

import (
	"math"
)

const TAU = math.Pi * 2

// Calculate an angle between two unit vectors
//
// Since we measure angle between radii of circular arcs,
// we can use simplified math (without length normalization)
//
func unit_vector_angle(ux, uy, vx, vy float64) float64 {
	sign := ux*vy - uy*vx
	if sign < 0 {
		sign = -1
	} else {
		sign = 1
	}

	dot := ux*vx + uy*vy

	// Add this to work with arbitrary vectors:
	// dot /= Math.sqrt(ux * ux + uy * uy) * Math.sqrt(vx * vx + vy * vy);

	// rounding errors, e.g. -1.0000000000000002 can screw up this
	if dot > 1.0 {
		dot = 1.0
	}
	if dot < -1.0 {
		dot = -1.0
	}

	return sign * math.Acos(dot)
}

// Convert from endpoint to center parameterization,
// see http://www.w3.org/TR/SVG11/implnote.html#ArcImplementationNotes
//
// Return [cx, cy, theta1, delta_theta]
//
func get_arc_center(x1, y1, x2, y2, fa, fs, rx, ry, sin_phi, cos_phi float64) []float64 {
	// Step 1.
	//
	// Moving an ellipse so origin will be the middlepoint between our two
	// points. After that, rotate it to line up ellipse axes with coordinate
	// axes.
	//
	x1p := cos_phi*(x1-x2)/2 + sin_phi*(y1-y2)/2
	y1p := -sin_phi*(x1-x2)/2 + cos_phi*(y1-y2)/2

	rx_sq := rx * rx
	ry_sq := ry * ry
	x1p_sq := x1p * x1p
	y1p_sq := y1p * y1p

	// Step 2.
	//
	// Compute coordinates of the centre of this ellipse (cx', cy')
	// in the new coordinate system.
	//
	radicant := (rx_sq * ry_sq) - (rx_sq * y1p_sq) - (ry_sq * x1p_sq)

	if radicant < 0 {
		// due to rounding errors it might be e.g. -1.3877787807814457e-17
		radicant = 0
	}

	radicant /= (rx_sq * y1p_sq) + (ry_sq * x1p_sq)
	sign := 1.0
	if fa == fs {
		sign = -1.0
	}
	radicant = math.Sqrt(radicant) * sign

	cxp := radicant * rx / ry * y1p
	cyp := radicant * -ry / rx * x1p

	// Step 3.
	//
	// Transform back to get centre coordinates (cx, cy) in the original
	// coordinate system.
	//
	cx := cos_phi*cxp - sin_phi*cyp + (x1+x2)/2
	cy := sin_phi*cxp + cos_phi*cyp + (y1+y2)/2

	// Step 4.
	//
	// Compute angles (theta1, delta_theta).
	//
	v1x := (x1p - cxp) / rx
	v1y := (y1p - cyp) / ry
	v2x := (-x1p - cxp) / rx
	v2y := (-y1p - cyp) / ry

	theta1 := unit_vector_angle(1, 0, v1x, v1y)
	delta_theta := unit_vector_angle(v1x, v1y, v2x, v2y)

	if fs == 0 && delta_theta > 0 {
		delta_theta -= TAU
	}
	if fs == 1 && delta_theta < 0 {
		delta_theta += TAU
	}

	return []float64{cx, cy, theta1, delta_theta}
}

//
// Approximate one unit arc segment with bézier curves,
// see http://math.stackexchange.com/questions/873224
//
func approximate_unit_arc(theta1, delta_theta float64) []float64 {
	alpha := 4.0 / 3.0 * math.Tan(delta_theta/4.0)

	x1 := math.Cos(theta1)
	y1 := math.Sin(theta1)
	x2 := math.Cos(theta1 + delta_theta)
	y2 := math.Sin(theta1 + delta_theta)

	return []float64{x1, y1, x1 - y1*alpha, y1 + x1*alpha, x2 + y2*alpha, y2 - x2*alpha, x2, y2}
}

func a2c(x1, y1, x2, y2, fa, fs, rx, ry, phi float64) [][]float64 {
	sin_phi := math.Sin(phi * TAU / 360.0)
	cos_phi := math.Cos(phi * TAU / 360.0)

	// Make sure radii are valid
	//
	x1p := cos_phi*(x1-x2)/2 + sin_phi*(y1-y2)/2
	y1p := -sin_phi*(x1-x2)/2 + cos_phi*(y1-y2)/2

	if x1p == 0.0 && y1p == 0.0 {
		// we're asked to draw line to itself
		return [][]float64{}
	}

	if rx == 0 || ry == 0 {
		// one of the radii is zero
		return [][]float64{}
	}

	// Compensate out-of-range radii
	//
	rx = math.Abs(rx)
	ry = math.Abs(ry)

	lambda := (x1p*x1p)/(rx*rx) + (y1p*y1p)/(ry*ry)
	if lambda > 1 {
		rx *= math.Sqrt(lambda)
		ry *= math.Sqrt(lambda)
	}

	// Get center parameters (cx, cy, theta1, delta_theta)
	//
	cc := get_arc_center(x1, y1, x2, y2, fa, fs, rx, ry, sin_phi, cos_phi)

	result := [][]float64{}
	theta1 := cc[2]
	delta_theta := cc[3]

	// Split an arc to multiple segments, so each segment
	// will be less than τ/4 (= 90°)
	//
	segments := math.Max(math.Ceil(math.Abs(delta_theta)/(TAU/4.0)), 1.0)
	delta_theta /= segments

	for i := 0; i < int(segments); i++ {
		result = append(result, approximate_unit_arc(theta1, delta_theta))
		theta1 += delta_theta
	}

	// We have a bezier approximation of a unit circle,
	// now need to transform back to the original ellipse
	//
	for j := 0; j < len(result); j++ {
		curve := result[j]
		for i := 0; i < len(curve); i += 2 {
			x := curve[i+0]
			y := curve[i+1]

			// scale
			x *= rx
			y *= ry

			// rotate
			xp := cos_phi*x - sin_phi*y
			yp := sin_phi*x + cos_phi*y

			// translate
			curve[i+0] = xp + cc[0]
			curve[i+1] = yp + cc[1]
		}

	}
	return result
}
