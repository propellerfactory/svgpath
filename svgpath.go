// port of https://github.com/fontello/svgpath
package svgpath

import (
	"fmt"
	"math"
	"regexp"
	"strings"
)

// SVG Path transformations library
//
// Usage:
//
//    SvgPath('...')
//      .translate(-150, -100)
//      .scale(0.5)
//      .translate(-150, -100)
//      .toFixed(1)
//      .toString()
//

type SvgPath struct {
	segments [][]interface{}
	stack    []*Matrix
}

// Class constructor
//
func NewSvgPath(path string) (*SvgPath, error) {
	segments, err := PathParse(path)
	if err != nil {
		return nil, err
	}
	return &SvgPath{
		segments: segments,
		// Transforms stack for lazy evaluation
		stack: []*Matrix{},
	}, nil
}

func (sp *SvgPath) matrix(m *Matrix) {

	// Quick leave for empty matrix
	if len(m.queue) == 0 {
		return
	}

	sp.iterate(func(s []interface{}, index int, x float64, y float64) [][]interface{} {
		var result []interface{}
		arc := func() [][]interface{} {
			// ARC is: ['A', rx, ry, x-axis-rotation, large-arc-flag, sweep-flag, x, y]

			// Drop segment if arc is empty (end point === start point)
			if (s[0].(string) == "A" && s[6].(float64) == x && s[7].(float64) == y) ||
				(s[0].(string) == "A" && s[6].(float64) == 0.0 && s[7].(float64) == 0.0) {
				return [][]interface{}{}
			}

			// Transform rx, ry and the x-axis-rotation
			ma := m.ToArray()
			e := NewEllipse(s[1].(float64), s[2].(float64), s[3].(float64))
			e.Transform(ma)

			// flip sweep-flag if matrix is not orientation-preserving
			if ma[0]*ma[3]-ma[1]*ma[2] < 0 {
				if s[5].(float64) != 0.0 {
					s[5] = "0"
				} else {
					s[5] = "1"
				}
			}

			// Transform end point as usual (without translation for relative notation)
			p := m.Calc(s[6].(float64), s[7].(float64), s[0].(string) == "a")

			// Empty arcs can be ignored by renderer, but should not be dropped
			// to avoid collisions with `S A S` and so on. Replace with empty line.
			if (s[0].(string) == "A" && s[6].(float64) == x && s[7].(float64) == y) ||
				(s[0].(string) == "a" && s[6].(float64) == 0.0 && s[7].(float64) == 0.0) {
				if s[0].(float64) == 'a' {
					result = []interface{}{"l", p[0], p[1]}
				} else {
					result = []interface{}{"L", p[0], p[1]}
				}
				return nil
			}

			// if the resulting ellipse is (almost) a segment ...
			if e.IsDegenerate() {
				// replace the arc by a line
				if s[0].(float64) == 'a' {
					result = []interface{}{"l", p[0], p[1]}
				} else {
					result = []interface{}{"L", p[0], p[1]}
				}
			} else {
				// if it is a real ellipse
				// s[0], s[4] and s[5] are not modified
				result = []interface{}{s[0], e.rx, e.ry, e.ax, s[4], s[5], p[0], p[1]}
			}
			return nil
		}

		switch s[0] {
		// Process 'assymetric' commands separately
		case "v":
			p := m.Calc(0, s[1].(float64), true)
			if p[0] == 0 {
				result = []interface{}{"v", p[1]}
			} else {
				result = []interface{}{"l", p[0], p[1]}
			}
		case "V":
			p := m.Calc(x, s[1].(float64), false)
			if p[0] == m.Calc(x, y, false)[0] {
				result = []interface{}{"V", p[1]}
			} else {
				result = []interface{}{"L", p[0], p[1]}
			}

		case "h":
			p := m.Calc(s[1].(float64), 0, true)
			if p[1] == 0 {
				result = []interface{}{"h", p[0]}
			} else {
				result = []interface{}{"l", p[0], p[1]}
			}

		case "H":
			p := m.Calc(s[1].(float64), y, false)
			if p[1] == m.Calc(x, y, false)[1] {
				result = []interface{}{"H", p[0]}
			} else {
				result = []interface{}{"L", p[0], p[1]}
			}
		case "a":
			if r := arc(); r != nil {
				return r
			}
		case "A":
			if r := arc(); r != nil {
				return r
			}
		case "m":
			// Edge case. The very first `m` should be processed as absolute, if happens.
			// Make sense for coord shift transforms.
			isRelative := index > 0

			p := m.Calc(s[1].(float64), s[2].(float64), isRelative)
			result = []interface{}{"m", p[0], p[1]}

		default:
			name := s[0].(string)
			result = []interface{}{name}
			isRelative := strings.ToLower(name) == name

			// Apply transformations to the segment
			for i := 1; i < len(s); i += 2 {
				p := m.Calc(s[i].(float64), s[i+1].(float64), isRelative)
				result = append(result, p[0], p[1])
			}
		}

		sp.segments[index] = result
		return nil
	}, true)
}

// Apply stacked commands
//
func (sp *SvgPath) evaluateStack() {
	if len(sp.stack) == 0 {
		return
	}

	if len(sp.stack) == 1 {
		sp.matrix(sp.stack[0])
		sp.stack = []*Matrix{}
		return
	}

	m := NewMatrix()
	for i := len(sp.stack) - 1; i >= 0; i-- {
		m.Matrix(sp.stack[i].ToArray())
	}

	sp.matrix(m)
	sp.stack = []*Matrix{}
}

// Convert processed SVG Path back to string
//
func (sp *SvgPath) ToString() string {
	elements := []string{}

	sp.evaluateStack()

	for i := 0; i < len(sp.segments); i++ {
		// remove repeating commands names
		cmd := sp.segments[i][0].(string)
		skipCmd := i > 0 && cmd != "m" && cmd != "M" && cmd == sp.segments[i-1][0]
		if !skipCmd {
			elements = append(elements, cmd)
		}
		for _, param := range sp.segments[i][1:] {
			elements = append(elements, fmt.Sprintf("%g", param))
		}
	}

	// TODO: Optimizations: remove spaces around commands & before `-`
	//
	// We could also remove leading zeros for `0.5`-like values,
	// but their count is too small to spend time for.
	result := strings.Join(elements, " ")
	result = regexp.MustCompile("(?i) ?([achlmqrstvz]) ?").ReplaceAllString(result, "$1")
	result = regexp.MustCompile(" \\-").ReplaceAllString(result, "-")
	return result
}

func toFixed(value float64, precision int) float64 {
	s := math.Pow10(precision)
	return math.Round(value*s) / s
}

// Round coords with given decimal precision.
// 0 by default (to integers)
//
func (sp *SvgPath) Round(d int) {
	contourStartDeltaX := 0.0
	contourStartDeltaY := 0.0
	deltaX := 0.0
	deltaY := 0.0
	l := 0

	sp.evaluateStack()

	for _, s := range sp.segments {
		cmd := s[0].(string)
		isRelative := strings.ToLower(cmd) == s[0].(string)

		switch strings.ToLower(cmd) {
		case "h":
			if isRelative {
				s[1] = s[1].(float64) + deltaX
			}
			deltaX = s[1].(float64) - toFixed(s[1].(float64), d)
			s[1] = toFixed(s[1].(float64), d)

		case "v":
			if isRelative {
				s[1] = s[1].(float64) + deltaY
			}
			deltaY = s[1].(float64) - toFixed(s[1].(float64), d)
			s[1] = toFixed(s[1].(float64), d)

		case "z":
			deltaX = contourStartDeltaX
			deltaY = contourStartDeltaY

		case "m":
			if isRelative {
				s[1] = s[1].(float64) + deltaX
				s[2] = s[2].(float64) + deltaY
			}

			deltaX = s[1].(float64) - toFixed(s[1].(float64), d)
			deltaY = s[2].(float64) - toFixed(s[2].(float64), d)

			contourStartDeltaX = deltaX
			contourStartDeltaY = deltaY

			s[1] = toFixed(s[1].(float64), d)
			s[2] = toFixed(s[2].(float64), d)

		case "a":
			// [cmd, rx, ry, x-axis-rotation, large-arc-flag, sweep-flag, x, y]
			if isRelative {
				s[6] = s[6].(float64) + deltaX
				s[7] = s[7].(float64) + deltaY
			}

			deltaX = s[6].(float64) - toFixed(s[6].(float64), d)
			deltaY = s[7].(float64) - toFixed(s[7].(float64), d)

			s[1] = toFixed(s[1].(float64), d)
			s[2] = toFixed(s[2].(float64), d)
			s[3] = toFixed(s[3].(float64), d+2) // better precision for rotation
			s[6] = toFixed(s[6].(float64), d)
			s[7] = toFixed(s[7].(float64), d)

		default:
			// a c l q s t
			l = len(s)

			if isRelative {
				s[l-2] = s[l-2].(float64) + deltaX
				s[l-1] = s[l-1].(float64) + deltaY
			}

			deltaX = s[l-2].(float64) - toFixed(s[l-2].(float64), d)
			deltaY = s[l-1].(float64) - toFixed(s[l-1].(float64), d)

			for i := 1; i < len(s); i++ {
				s[i] = toFixed(s[i].(float64), d)
			}
		}
	}
}

type IteratorFn func(segment []interface{}, index int, lastX float64, lastY float64) [][]interface{}

// Apply iterator function to all segments. If function returns result,
// current segment will be replaced to array of returned segments.
// If empty array is returned, current segment will be deleted.
func (sp *SvgPath) iterate(iterator IteratorFn, keepLazyStack bool) {
	segments := sp.segments
	replacements := map[int][][]interface{}{}
	needReplace := false
	lastX := 0.0
	lastY := 0.0
	countourStartX := 0.0
	countourStartY := 0.0

	if !keepLazyStack {
		sp.evaluateStack()
	}

	for index, s := range segments {

		res := iterator(s, index, lastX, lastY)

		if res != nil {
			replacements[index] = res
			needReplace = true
		}

		var isRelative = s[0] == strings.ToLower(s[0].(string))

		// calculate absolute X and Y
		switch strings.ToLower(s[0].(string)) {
		case "m":
			s1 := s[1].(float64)
			s2 := s[2].(float64)
			if isRelative {
				lastX += s1
				lastY += s2
			} else {
				lastX = s1
				lastY = s2
			}
			countourStartX = lastX
			countourStartY = lastY
		case "h":
			s1 := s[1].(float64)
			if isRelative {
				lastX += s1
			} else {
				lastX = s1
			}
		case "v":
			s1 := s[1].(float64)
			if isRelative {
				lastY += s1
			} else {
				lastY = s1
			}
		case "z":
			// That make sence for multiple contours
			lastX = countourStartX
			lastY = countourStartY

		default:
			s1 := s[len(s)-2].(float64)
			s2 := s[len(s)-1].(float64)
			if isRelative {
				lastX += s1
				lastY += s2
			} else {
				lastX = s1
				lastY = s2
			}
		}
	}

	// Replace segments if iterator return results

	if !needReplace {
		return
	}

	newSegments := [][]interface{}{}

	for i := 0; i < len(segments); i++ {
		if replacements[i] != nil {
			for j := 0; j < len(replacements[i]); j++ {
				newSegments = append(newSegments, replacements[i][j])
			}
		} else {
			newSegments = append(newSegments, segments[i])
		}
	}

	sp.segments = newSegments
}

// Translate path to (x  y)
//
func (sp *SvgPath) Translate(x, y float64) {
	m := NewMatrix()
	m.Translate(x, y)
	sp.stack = append(sp.stack, m)
}

// Scale path to (sx , sy)
//
func (sp *SvgPath) Scale(sx, sy float64) {
	m := NewMatrix()
	m.Scale(sx, sy)
	sp.stack = append(sp.stack, m)
}

// Rotate path around point (sx , sy)
//
func (sp *SvgPath) Rotate(angle, rx, ry float64) {
	m := NewMatrix()
	m.Rotate(angle, rx, ry)
	sp.stack = append(sp.stack, m)
}

// Skew path along the X axis by `degrees` angle
//
func (sp *SvgPath) SkewX(degrees float64) {
	m := NewMatrix()
	m.SkewX(degrees)
	sp.stack = append(sp.stack, m)
}

// Skew path along the Y axis by `degrees` angle
//
func (sp *SvgPath) SkewY(degrees float64) {
	m := NewMatrix()
	m.SkewY(degrees)
	sp.stack = append(sp.stack, m)
}

// Apply matrix transform (array of 6 elements)
//
func (sp *SvgPath) Matrix(mv []float64) {
	m := NewMatrix()
	m.Matrix(mv)
	sp.stack = append(sp.stack, m)
}

// Transform path according to "transform" attr of SVG spec
//
func (sp *SvgPath) Transform(transformString string) {
	transformString = strings.Trim(transformString, " ")
	if len(transformString) == 0 {
		return
	}
	sp.stack = append(sp.stack, TransformParse(transformString))
}

// Converts segments from relative to absolute
//
func (sp *SvgPath) Abs() {

	sp.iterate(func(s []interface{}, index int, x float64, y float64) [][]interface{} {
		name := s[0].(string)
		nameUC := strings.ToUpper(name)

		// Skip absolute commands
		if name == nameUC {
			return nil
		}

		s[0] = nameUC

		switch name {
		case "v":
			// v has shifted coords parity
			s[1] = s[1].(float64) + y

		case "a":
			// ARC is: ['A', rx, ry, x-axis-rotation, large-arc-flag, sweep-flag, x, y]
			// touch x, y only
			s[6] = s[6].(float64) + x
			s[7] = s[7].(float64) + y

		default:
			for i := 1; i < len(s); i++ {
				// odd values are X, even - Y
				if i%2 == 0 {
					s[i] = s[i].(float64) + y
				} else {
					s[i] = s[i].(float64) + x
				}
			}
		}
		return nil
	}, true)
}

// Converts segments from absolute to relative
//
func (sp *SvgPath) Rel() {

	sp.iterate(func(s []interface{}, index int, x float64, y float64) [][]interface{} {
		name := s[0].(string)
		nameLC := strings.ToLower(name)

		// Skip relative commands
		if name == nameLC {
			return nil
		}

		// Don't touch the first M to avoid potential confusions.
		if index == 0 && name == "M" {
			return nil
		}

		s[0] = nameLC

		switch name {
		case "V":
			// V has shifted coords parity
			s[1] = s[1].(float64) - y

		case "A":
			// ARC is: ['A', rx, ry, x-axis-rotation, large-arc-flag, sweep-flag, x, y]
			// touch x, y only
			s[6] = s[6].(float64) - x
			s[7] = s[7].(float64) - y

		default:
			for i := 1; i < len(s); i++ {
				// odd values are X, even - Y
				if i%2 == 0 {
					s[i] = s[i].(float64) - y
				} else {
					s[i] = s[i].(float64) - x
				}
			}
		}
		return nil
	}, true)
}

// Converts arcs to cubic bézier curves
//
func (sp *SvgPath) Unarc() {
	sp.iterate(func(s []interface{}, index int, x float64, y float64) [][]interface{} {
		result := [][]interface{}{}
		name := s[0].(string)

		// Skip anything except arcs
		if name != "A" && name != "a" {
			return nil
		}

		var nextX, nextY float64
		if name == "a" {
			// convert relative arc coordinates to absolute
			nextX = x + s[6].(float64)
			nextY = y + s[7].(float64)
		} else {
			nextX = s[6].(float64)
			nextY = s[7].(float64)
		}

		new_segments := a2c(x, y, nextX, nextY, s[4].(float64), s[5].(float64), s[1].(float64), s[2].(float64), s[3].(float64))

		// Degenerated arcs can be ignored by renderer, but should not be dropped
		// to avoid collisions with `S A S` and so on. Replace with empty line.
		if len(new_segments) == 0 {
			if s[0].(string) == "a" {
				return [][]interface{}{[]interface{}{"l", s[6], s[7]}}

			} else {
				return [][]interface{}{[]interface{}{"L", s[6], s[7]}}
			}
		}

		for _, s := range new_segments {
			result = append(result, []interface{}{"C", s[2], s[3], s[4], s[5], s[6], s[7]})
		}

		return result
	}, false)
}

// Converts smooth curves (with missed control point) to generic curves
//
func (sp *SvgPath) Unshort() {
	segments := sp.segments
	var prevControlX, prevControlY float64
	var curControlX, curControlY float64

	// TODO: add lazy evaluation flag when relative commands supported

	sp.iterate(func(s []interface{}, index int, x float64, y float64) [][]interface{} {
		name := s[0].(string)
		nameUC := strings.ToUpper(name)

		// First command MUST be M|m, it's safe to skip.
		// Protect from access to [-1] for sure.
		if index == 0 {
			return nil
		}

		isRelative := false
		if nameUC == "T" { // quadratic curve
			isRelative = name == "t"

			prevSegment := segments[index-1]

			if prevSegment[0].(string) == "Q" {
				prevControlX = prevSegment[1].(float64) - x
				prevControlY = prevSegment[2].(float64) - y
			} else if prevSegment[0].(string) == "q" {
				prevControlX = prevSegment[1].(float64) - prevSegment[3].(float64)
				prevControlY = prevSegment[2].(float64) - prevSegment[4].(float64)
			} else {
				prevControlX = 0.0
				prevControlY = 0.0
			}

			curControlX = -prevControlX
			curControlY = -prevControlY

			if !isRelative {
				curControlX += x
				curControlY += y
			}

			if isRelative {
				segments[index] = []interface{}{
					"q",
					curControlX, curControlY,
					s[1], s[2],
				}
			} else {
				segments[index] = []interface{}{
					"Q",
					curControlX, curControlY,
					s[1], s[2],
				}

			}

		} else if nameUC == "S" { // cubic curve
			isRelative := name == "s"

			prevSegment := segments[index-1]

			if prevSegment[0].(string) == "C" {
				prevControlX = prevSegment[3].(float64) - x
				prevControlY = prevSegment[4].(float64) - y
			} else if prevSegment[0].(string) == "c" {
				prevControlX = prevSegment[3].(float64) - prevSegment[5].(float64)
				prevControlY = prevSegment[4].(float64) - prevSegment[6].(float64)
			} else {
				prevControlX = 0.0
				prevControlY = 0.0
			}

			curControlX = -prevControlX
			curControlY = -prevControlY

			if !isRelative {
				curControlX += x
				curControlY += y
			}

			if isRelative {
				segments[index] = []interface{}{
					"c",
					curControlX, curControlY,
					s[1], s[2], s[3], s[4],
				}
			} else {
				segments[index] = []interface{}{
					"C",
					curControlX, curControlY,
					s[1], s[2], s[3], s[4],
				}
			}
		}
		return nil
	}, false)
}
