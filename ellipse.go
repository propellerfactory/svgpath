// port of https://github.com/fontello/svgpath
package svgpath

import (
	"math"
)

const epsilon = 0.0000000001

// To convert degree in radians
//
const torad = math.Pi / 180

type Ellipse struct {
	rx, ry, ax float64
}

// Class constructor :
//  an ellipse centred at 0 with radii rx,ry and x - axis - angle ax.
//
func NewEllipse(rx, ry, ax float64) *Ellipse {
	return &Ellipse{
		rx: rx,
		ry: ry,
		ax: ax,
	}
}

func (e *Ellipse) Transform(m []float64) {
	// We consider the current ellipse as image of the unit circle
	// by first scale(rx,ry) and then rotate(ax) ...
	// So we apply ma =  m x rotate(ax) x scale(rx,ry) to the unit circle.
	c := math.Cos(e.ax * torad)
	s := math.Sin(e.ax * torad)
	ma := []float64{
		e.rx * (m[0]*c + m[2]*s),
		e.rx * (m[1]*c + m[3]*s),
		e.ry * (-m[0]*s + m[2]*c),
		e.ry * (-m[1]*s + m[3]*c),
	}

	// ma * transpose(ma) = [ J L ]
	//                      [ L K ]
	// L is calculated later (if the image is not a circle)
	J := ma[0]*ma[0] + ma[2]*ma[2]
	K := ma[1]*ma[1] + ma[3]*ma[3]

	// the discriminant of the characteristic polynomial of ma * transpose(ma)
	D := ((ma[0]-ma[3])*(ma[0]-ma[3]) + (ma[2]+ma[1])*(ma[2]+ma[1])) *
		((ma[0]+ma[3])*(ma[0]+ma[3]) + (ma[2]-ma[1])*(ma[2]-ma[1]))

	// the "mean eigenvalue"
	JK := (J + K) / 2

	// check if the image is (almost) a circle
	if D < epsilon*JK {
		// if it is
		e.rx = math.Sqrt(JK)
		e.ry = e.rx
		e.ax = 0
		return
	}

	// if it is not a circle
	L := ma[0]*ma[1] + ma[2]*ma[3]

	D = math.Sqrt(D)

	// {l1,l2} = the two eigen values of ma * transpose(ma)
	l1 := JK + D/2
	l2 := JK - D/2
	// the x - axis - rotation angle is the argument of the l1 - eigenvector

	if math.Abs(L) < epsilon && math.Abs(l1-K) < epsilon {
		e.ax = 90 * 180 / math.Pi
	} else {
		if math.Abs(L) > math.Abs(l1-K) {
			e.ax = math.Atan((l1-J)/L) * 180.0 / math.Pi
		} else {
			e.ax = math.Atan(L/(l1-K)) * 180.0 / math.Pi
		}
	}

	// if ax > 0 => rx = sqrt(l1), ry = sqrt(l2), else exchange axes and ax += 90
	if e.ax >= 0.0 {
		// if ax in [0,90]
		e.rx = math.Sqrt(l1)
		e.ry = math.Sqrt(l2)
	} else {
		// if ax in ]-90,0[ => exchange axes
		e.ax += 90.0
		e.rx = math.Sqrt(l2)
		e.ry = math.Sqrt(l1)
	}
}

// Check if the ellipse is (almost) degenerate, i.e. rx = 0 or ry = 0
//
func (e *Ellipse) IsDegenerate() bool {
	return (e.rx < epsilon*e.ry || e.ry < epsilon*e.rx)
}
