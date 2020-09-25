// port of https://github.com/fontello/svgpath
package svgpath

import (
	"fmt"
	"math"
	"regexp"
	"strings"

	"github.com/propellerfactory/cubic2quad"
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

type Segment struct {
	Command string
	Params  []float64
}

type SvgPath struct {
	segments []*Segment
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

func (sp *SvgPath) Segments() []*Segment {
	return sp.segments
}

func (sp *SvgPath) matrix(m *Matrix) {

	// Quick leave for empty matrix
	if len(m.queue) == 0 {
		return
	}

	sp.iterate(func(s *Segment, index int, x float64, y float64) []*Segment {
		var result *Segment
		arc := func() []*Segment {
			// ARC is: ['A', rx, ry, x-axis-rotation, large-arc-flag, sweep-flag, x, y]

			// Drop segment if arc is empty (end point === start point)
			if (s.Command == "A" && s.Params[5] == x && s.Params[6] == y) ||
				(s.Command == "A" && s.Params[5] == 0.0 && s.Params[6] == 0.0) {
				return []*Segment{}
			}

			// Transform rx, ry and the x-axis-rotation
			ma := m.ToArray()
			e := NewEllipse(s.Params[0], s.Params[1], s.Params[2])
			e.Transform(ma)

			// flip sweep-flag if matrix is not orientation-preserving
			if ma[0]*ma[3]-ma[1]*ma[2] < 0 {
				if s.Params[4] != 0.0 {
					s.Params[4] = 0.0
				} else {
					s.Params[4] = 1.0
				}
			}

			// Transform end point as usual (without translation for relative notation)
			p := m.Calc(s.Params[5], s.Params[6], s.Command == "a")

			// Empty arcs can be ignored by renderer, but should not be dropped
			// to avoid collisions with `S A S` and so on. Replace with empty line.
			if (s.Command == "A" && s.Params[5] == x && s.Params[6] == y) ||
				(s.Command == "a" && s.Params[5] == 0.0 && s.Params[6] == 0.0) {
				if s.Command == "a" {
					result = &Segment{Command: "l", Params: []float64{p[0], p[1]}}
				} else {
					result = &Segment{Command: "L", Params: []float64{p[0], p[1]}}
				}
				return nil
			}

			// if the resulting ellipse is (almost) a segment ...
			if e.IsDegenerate() {
				// replace the arc by a line
				if s.Command == "a" {
					result = &Segment{Command: "l", Params: []float64{p[0], p[1]}}
				} else {
					result = &Segment{Command: "L", Params: []float64{p[0], p[1]}}
				}
			} else {
				// if it is a real ellipse
				// s[0], s.Params[3] and s.Params[4] are not modified
				result = &Segment{Command: s.Command, Params: []float64{e.rx, e.ry, e.ax, s.Params[3], s.Params[4], p[0], p[1]}}
			}
			return nil
		}

		switch s.Command {
		// Process 'assymetric' commands separately
		case "v":
			p := m.Calc(0, s.Params[0], true)
			if p[0] == 0 {
				result = &Segment{Command: "v", Params: []float64{p[1]}}
			} else {
				result = &Segment{Command: "l", Params: []float64{p[0], p[1]}}
			}
		case "V":
			p := m.Calc(x, s.Params[0], false)
			if p[0] == m.Calc(x, y, false)[0] {
				result = &Segment{Command: "V", Params: []float64{p[1]}}
			} else {
				result = &Segment{Command: "L", Params: []float64{p[0], p[1]}}
			}

		case "h":
			p := m.Calc(s.Params[0], 0, true)
			if p[1] == 0 {
				result = &Segment{Command: "h", Params: []float64{p[0]}}
			} else {
				result = &Segment{Command: "l", Params: []float64{p[0], p[1]}}
			}

		case "H":
			p := m.Calc(s.Params[0], y, false)
			if p[1] == m.Calc(x, y, false)[1] {
				result = &Segment{Command: "H", Params: []float64{p[0]}}
			} else {
				result = &Segment{Command: "L", Params: []float64{p[0], p[1]}}
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

			p := m.Calc(s.Params[0], s.Params[1], isRelative)
			result = &Segment{Command: "m", Params: []float64{p[0], p[1]}}

		default:
			name := s.Command
			result = &Segment{Command: name}
			isRelative := strings.ToLower(name) == name

			// Apply transformations to the segment
			for i := 0; i < len(s.Params); i += 2 {
				p := m.Calc(s.Params[i], s.Params[i+1], isRelative)
				result.Params = append(result.Params, p[0], p[1])
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
		cmd := sp.segments[i].Command
		skipCmd := i > 0 && cmd != "m" && cmd != "M" && cmd == sp.segments[i-1].Command
		if !skipCmd {
			elements = append(elements, cmd)
		}
		for _, param := range sp.segments[i].Params {
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
		cmd := s.Command
		isRelative := strings.ToLower(cmd) == s.Command

		switch strings.ToLower(cmd) {
		case "h":
			if isRelative {
				s.Params[0] = s.Params[0] + deltaX
			}
			deltaX = s.Params[0] - toFixed(s.Params[0], d)
			s.Params[0] = toFixed(s.Params[0], d)

		case "v":
			if isRelative {
				s.Params[0] = s.Params[0] + deltaY
			}
			deltaY = s.Params[0] - toFixed(s.Params[0], d)
			s.Params[0] = toFixed(s.Params[0], d)

		case "z":
			deltaX = contourStartDeltaX
			deltaY = contourStartDeltaY

		case "m":
			if isRelative {
				s.Params[0] = s.Params[0] + deltaX
				s.Params[1] = s.Params[1] + deltaY
			}

			deltaX = s.Params[0] - toFixed(s.Params[0], d)
			deltaY = s.Params[1] - toFixed(s.Params[1], d)

			contourStartDeltaX = deltaX
			contourStartDeltaY = deltaY

			s.Params[0] = toFixed(s.Params[0], d)
			s.Params[1] = toFixed(s.Params[1], d)

		case "a":
			// [cmd, rx, ry, x-axis-rotation, large-arc-flag, sweep-flag, x, y]
			if isRelative {
				s.Params[5] = s.Params[5] + deltaX
				s.Params[6] = s.Params[6] + deltaY
			}

			deltaX = s.Params[5] - toFixed(s.Params[5], d)
			deltaY = s.Params[6] - toFixed(s.Params[6], d)

			s.Params[0] = toFixed(s.Params[0], d)
			s.Params[1] = toFixed(s.Params[1], d)
			s.Params[2] = toFixed(s.Params[2], d+2) // better precision for rotation
			s.Params[5] = toFixed(s.Params[5], d)
			s.Params[6] = toFixed(s.Params[6], d)

		default:
			// a c l q s t
			l = len(s.Params)

			if isRelative {
				s.Params[l-2] = s.Params[l-2] + deltaX
				s.Params[l-1] = s.Params[l-1] + deltaY
			}

			deltaX = s.Params[l-2] - toFixed(s.Params[l-2], d)
			deltaY = s.Params[l-1] - toFixed(s.Params[l-1], d)

			for i := 0; i < len(s.Params); i++ {
				s.Params[i] = toFixed(s.Params[i], d)
			}
		}
	}
}

type IteratorFn func(segment *Segment, index int, lastX float64, lastY float64) []*Segment

// Apply iterator function to all segments. If function returns result,
// current segment will be replaced to array of returned segments.
// If empty array is returned, current segment will be deleted.
func (sp *SvgPath) iterate(iterator IteratorFn, keepLazyStack bool) {
	segments := sp.segments
	replacements := map[int][]*Segment{}
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

		var isRelative = s.Command == strings.ToLower(s.Command)

		// calculate absolute X and Y
		switch strings.ToLower(s.Command) {
		case "m":
			s1 := s.Params[0]
			s2 := s.Params[1]
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
			s1 := s.Params[0]
			if isRelative {
				lastX += s1
			} else {
				lastX = s1
			}
		case "v":
			s1 := s.Params[0]
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
			s1 := s.Params[len(s.Params)-2]
			s2 := s.Params[len(s.Params)-1]
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

	newSegments := []*Segment{}

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

// Converts *Segments from relative to absolute
//
func (sp *SvgPath) Abs() {

	sp.iterate(func(s *Segment, index int, x float64, y float64) []*Segment {
		name := s.Command
		nameUC := strings.ToUpper(name)

		// Skip absolute commands
		if name == nameUC {
			return nil
		}

		s.Command = nameUC

		switch name {
		case "v":
			// v has shifted coords parity
			s.Params[0] = s.Params[0] + y

		case "a":
			// ARC is: ['A', rx, ry, x-axis-rotation, large-arc-flag, sweep-flag, x, y]
			// touch x, y only
			s.Params[5] = s.Params[5] + x
			s.Params[6] = s.Params[6] + y

		default:
			for i := 0; i < len(s.Params); i++ {
				// odd values are Y, even - X
				if i%2 == 0 {
					s.Params[i] = s.Params[i] + x
				} else {
					s.Params[i] = s.Params[i] + y
				}
			}
		}
		return nil
	}, true)
}

// Converts *Segments from absolute to relative
//
func (sp *SvgPath) Rel() {

	sp.iterate(func(s *Segment, index int, x float64, y float64) []*Segment {
		name := s.Command
		nameLC := strings.ToLower(name)

		// Skip relative commands
		if name == nameLC {
			return nil
		}

		// Don't touch the first M to avoid potential confusions.
		if index == 0 && name == "M" {
			return nil
		}

		s.Command = nameLC

		switch name {
		case "V":
			// V has shifted coords parity
			s.Params[0] = s.Params[0] - y

		case "A":
			// ARC is: ['A', rx, ry, x-axis-rotation, large-arc-flag, sweep-flag, x, y]
			// touch x, y only
			s.Params[5] = s.Params[5] - x
			s.Params[6] = s.Params[6] - y

		default:
			for i := 0; i < len(s.Params); i++ {
				// odd values are Y, even - X
				if i%2 == 0 {
					s.Params[i] = s.Params[i] - x
				} else {
					s.Params[i] = s.Params[i] - y
				}
			}
		}
		return nil
	}, true)
}

// Converts arcs to cubic bézier curves
//
func (sp *SvgPath) Unarc() {
	sp.iterate(func(s *Segment, index int, x float64, y float64) []*Segment {
		result := []*Segment{}
		name := s.Command
		// Skip anything except arcs
		if name != "A" && name != "a" {
			return nil
		}

		var nextX, nextY float64
		if name == "a" {
			// convert relative arc coordinates to absolute
			nextX = x + s.Params[5]
			nextY = y + s.Params[6]
		} else {
			nextX = s.Params[5]
			nextY = s.Params[6]
		}

		new_segments := a2c(x, y, nextX, nextY, s.Params[3], s.Params[4], s.Params[0], s.Params[1], s.Params[2])

		// Degenerated arcs can be ignored by renderer, but should not be dropped
		// to avoid collisions with `S A S` and so on. Replace with empty line.
		if len(new_segments) == 0 {
			if s.Command == "a" {
				return []*Segment{&Segment{Command: "l", Params: []float64{s.Params[5], s.Params[6]}}}
			} else {
				return []*Segment{&Segment{Command: "L", Params: []float64{s.Params[5], s.Params[6]}}}
			}
		}

		for _, s := range new_segments {
			result = append(result, &Segment{Command: "C", Params: []float64{s[2], s[3], s[4], s[5], s[6], s[7]}})
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

	sp.iterate(func(s *Segment, index int, x float64, y float64) []*Segment {
		name := s.Command
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

			if prevSegment.Command == "Q" {
				prevControlX = prevSegment.Params[0] - x
				prevControlY = prevSegment.Params[1] - y
			} else if prevSegment.Command == "q" {
				prevControlX = prevSegment.Params[0] - prevSegment.Params[2]
				prevControlY = prevSegment.Params[1] - prevSegment.Params[3]
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
				segments[index] = &Segment{
					Command: "q",
					Params: []float64{
						curControlX, curControlY,
						s.Params[0], s.Params[1]},
				}
			} else {
				segments[index] = &Segment{
					Command: "Q",
					Params: []float64{curControlX, curControlY,
						s.Params[0], s.Params[1]},
				}
			}

		} else if nameUC == "S" { // cubic curve
			isRelative := name == "s"

			prevSegment := segments[index-1]

			if prevSegment.Command == "C" {
				prevControlX = prevSegment.Params[2] - x
				prevControlY = prevSegment.Params[3] - y
			} else if prevSegment.Command == "c" {
				prevControlX = prevSegment.Params[2] - prevSegment.Params[4]
				prevControlY = prevSegment.Params[3] - prevSegment.Params[5]
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
				segments[index] = &Segment{
					Command: "c",
					Params: []float64{curControlX, curControlY,
						s.Params[0], s.Params[1], s.Params[2], s.Params[3]},
				}
			} else {
				segments[index] = &Segment{
					Command: "C",
					Params: []float64{curControlX, curControlY,
						s.Params[0], s.Params[1], s.Params[2], s.Params[3]},
				}
			}
		}
		return nil
	}, false)
}

// Converts cubic bézier curves to quadratic bézier curves
//  NOTE: does not process "short" cubic bézier curves
//
func (sp *SvgPath) Uncubic() {
	sp.iterate(func(s *Segment, index int, x float64, y float64) []*Segment {
		result := []*Segment{}
		name := s.Command

		// Skip anything except cubics
		if name != "C" && name != "c" {
			return nil
		}

		var x1, y1, x2, y2, ex, ey float64
		if name == "c" {
			// convert relative cubic coordinates to absolute
			x1 = x + s.Params[0]
			y1 = y + s.Params[1]
			x2 = x + s.Params[2]
			y2 = y + s.Params[3]
			ex = x + s.Params[4]
			ey = y + s.Params[5]
		} else {
			x1 = s.Params[0]
			y1 = s.Params[1]
			x2 = s.Params[2]
			y2 = s.Params[3]
			ex = s.Params[4]
			ey = s.Params[5]
		}

		quad := cubic2quad.CubicToQuad(x, y, x1, y1, x2, y2, ex, ey, 0.0001)
		// Degenerated cubics can be ignored by renderer, but should not be dropped
		// to avoid collisions with `S A S` and so on. Replace with empty line.
		if len(quad) == 0 {
			if s.Command == "c" {
				return []*Segment{{Command: "l", Params: []float64{s.Params[4], s.Params[5]}}}
			}
			return []*Segment{{Command: "L", Params: []float64{s.Params[4], s.Params[5]}}}

		}

		if name == "c" {
			lastX := x
			lastY := y
			for i := 2; i < len(quad); i += 4 {
				result = append(result, &Segment{
					Command: "q",
					Params: []float64{quad[i+0] - lastX, quad[i+1] - lastY,
						quad[i+2] - lastX, quad[i+3] - lastY},
				})
				lastX = quad[i+2]
				lastY = quad[i+3]
			}
		} else {
			for i := 2; i < len(quad); i += 4 {
				result = append(result, &Segment{
					Command: "Q",
					Params: []float64{
						quad[i+0], quad[i+1],
						quad[i+2], quad[i+3]},
				})
			}
		}

		return result
	}, false)
}
