// port of https://github.com/fontello/svgpath
package svgpath

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/pkg/errors"
)

var paramCounts = map[string]int{"a": 7, "c": 6, "h": 1, "l": 2, "m": 2, "r": 4, "q": 4, "s": 4, "t": 2, "v": 1, "z": 0}

var SPECIAL_SPACES = []rune{
	0x1680, 0x180E, 0x2000, 0x2001, 0x2002, 0x2003, 0x2004, 0x2005, 0x2006,
	0x2007, 0x2008, 0x2009, 0x200A, 0x202F, 0x205F, 0x3000, 0xFEFF,
}

func isSpecialSpace(ch rune) bool {
	for _, cch := range SPECIAL_SPACES {
		if cch == ch {
			return true
		}
	}
	return false
}

func isSpace(ch rune) bool {
	return (ch == 0x0A) || (ch == 0x0D) || (ch == 0x2028) || (ch == 0x2029) || // Line terminators
		// White spaces
		(ch == 0x20) || (ch == 0x09) || (ch == 0x0B) || (ch == 0x0C) || (ch == 0xA0) ||
		(ch >= 0x1680 && isSpecialSpace(ch))
}

func isCommand(code rune) bool {
	/*eslint-disable no-bitwise*/
	switch code | 0x20 {
	case 0x6D /* m */ :
		return true
	case 0x7A /* z */ :
		return true
	case 0x6C /* l */ :
		return true
	case 0x68 /* h */ :
		return true
	case 0x76 /* v */ :
		return true
	case 0x63 /* c */ :
		return true
	case 0x73 /* s */ :
		return true
	case 0x71 /* q */ :
		return true
	case 0x74 /* t */ :
		return true
	case 0x61 /* a */ :
		return true
	case 0x72 /* r */ :
		return true
	}
	return false
}

func isDigit(code rune) bool {
	return (code >= 48 && code <= 57) // 0..9
}

func isDigitStart(code rune) bool {
	return (code >= 48 && code <= 57) || /* 0..9 */
		code == 0x2B || /* + */
		code == 0x2D || /* - */
		code == 0x2E /* . */
}

type State struct {
	index        int
	path         string
	max          int
	result       [][]interface{}
	param        float64
	err          error
	segmentStart int
	data         []float64
}

func NewState(path string) *State {
	return &State{
		index:        0,
		path:         path,
		max:          utf8.RuneCountInString(path),
		result:       [][]interface{}{},
		param:        0.0,
		err:          nil,
		segmentStart: 0,
		data:         []float64{},
	}
}

func (s *State) SkipSpaces() {
	for s.index < s.max && isSpace(s.getRuneAtIndex(s.index)) {
		s.index++
	}
}

func (s *State) getRuneAtIndex(index int) rune {
	i := 0
	for _, ch := range s.path {
		if i == index {
			return ch
		}
		i++
	}
	return 0
}

func (s *State) ScanParam() error {
	start := s.index
	index := start
	max := s.max
	zeroFirst := false
	hasCeiling := false
	hasDecimal := false
	hasDot := false

	if index >= max {
		s.err = errors.Errorf("SvgPath: missed param (at pos %d)", index)
		return s.err
	}
	ch := s.getRuneAtIndex(index)

	if ch == 0x2B /* + */ || ch == 0x2D /* - */ {
		index++
		ch = s.getRuneAtIndex(index)
	}

	// This logic is shamelessly borrowed from Esprima
	// https://github.com/ariya/esprimas
	//
	if !isDigit(ch) && ch != 0x2E /* . */ {
		s.err = errors.Errorf("SvgPath: param should start with 0..9 or `.` (at pos %d)", index)
		return s.err
	}

	if ch != 0x2E /* . */ {
		zeroFirst = (ch == 0x30 /* 0 */)
		index++

		ch = s.getRuneAtIndex(index)

		if zeroFirst && index < max {
			// decimal number starts with '0' such as '09' is illegal.
			if ch != 0 && isDigit(ch) {
				s.err = errors.Errorf("SvgPath: numbers started with `0` such as `09` are ilegal (at pos %d)", start)
				return s.err
			}
		}

		for {
			if index >= max {
				break
			} else if !isDigit(s.getRuneAtIndex(index)) {
				break
			}
			index++
			hasCeiling = true
		}
		ch = s.getRuneAtIndex(index)
	}

	if ch == 0x2E /* . */ {
		hasDot = true
		index++
		for {
			if !isDigit(s.getRuneAtIndex(index)) {
				break
			}
			index++
			hasDecimal = true

		}
		ch = s.getRuneAtIndex(index)
	}

	if ch == 0x65 /* e */ || ch == 0x45 /* E */ {
		if hasDot && !hasCeiling && !hasDecimal {
			s.err = errors.Errorf("SvgPath: invalid float exponent (at pos %d)", index)
			return s.err
		}

		index++

		ch = s.getRuneAtIndex(index)
		if ch == 0x2B /* + */ || ch == 0x2D /* - */ {
			index++
		}
		if index < max && isDigit(s.getRuneAtIndex(index)) {
			for {
				if index >= max {
					break
				} else if !isDigit(s.getRuneAtIndex(index)) {
					break
				}
				index++
			}
		} else {
			s.err = errors.Errorf("SvgPath: invalid float exponent (at pos %d)", index)
			return s.err
		}
	}

	value := ""
	s.index = index
	for i := start; i < index; i++ {
		value = fmt.Sprintf("%s%c", value, s.getRuneAtIndex(i))
	}
	param, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return errors.Wrap(err, "Failed to parse param")
	}
	s.param = param
	return nil
}

func (s *State) FinalizeSegment() {
	// Process duplicated commands (without comand name)

	// This logic is shamelessly borrowed from Raphael
	// https://github.com/DmitryBaranovskiy/raphael/
	//
	cmd := fmt.Sprintf("%c", s.getRuneAtIndex(s.segmentStart))
	cmdLC := strings.ToLower(cmd)

	params := s.data

	if cmdLC == "m" && len(params) > 2 {
		s.result = append(s.result, []interface{}{cmd, params[0], params[1]})
		params = params[2:]
		cmdLC = "l"
		if cmd == "m" {
			cmd = "l"
		} else {
			cmd = "L"
		}
	}

	if cmdLC == "r" {
		result := []interface{}{cmd}
		for _, param := range params {
			result = append(result, param)
		}
		s.result = append(s.result, result)
	} else {
		for len(params) >= paramCounts[cmdLC] {
			result := []interface{}{cmd}
			for i := 0; i < paramCounts[cmdLC]; i++ {
				result = append(result, params[i])
			}
			params = params[paramCounts[cmdLC]:]
			s.result = append(s.result, result)
			if paramCounts[cmdLC] == 0 {
				break
			}
		}
	}
}

func (s *State) ScanSegment() error {
	max := s.max

	s.segmentStart = s.index
	cmdCode := s.getRuneAtIndex(s.index)

	if !isCommand(cmdCode) {
		s.err = errors.Errorf("SvgPath: bad command %c (at pos %d)", s.getRuneAtIndex(s.index), s.index)
		return s.err
	}

	need_params := paramCounts[strings.ToLower(fmt.Sprintf("%c", s.getRuneAtIndex(s.index)))]

	s.index++
	s.SkipSpaces()

	s.data = []float64{}

	if need_params == 0 {
		// Z
		s.FinalizeSegment()
		return nil
	}

	comma_found := false

	for {
		for i := need_params; i > 0; i-- {
			s.ScanParam()
			if s.err != nil {
				return s.err
			}
			s.data = append(s.data, s.param)

			s.SkipSpaces()
			comma_found = false

			if s.index < max && s.getRuneAtIndex(s.index) == 0x2C /* , */ {
				s.index++
				s.SkipSpaces()
				comma_found = true
			}
		}

		// after ',' param is mandatory
		if comma_found {
			continue
		}

		if s.index >= s.max {
			break
		}

		// Stop on next segment
		if !isDigitStart(s.getRuneAtIndex(s.index)) {
			break
		}
	}

	s.FinalizeSegment()
	return nil
}

/* Returns array of segments:
 *
 * [
 *   [ command, coord1, coord2, ... ]
 * ]
 */
func PathParse(svgPath string) ([][]interface{}, error) {
	s := NewState(svgPath)
	max := s.max

	s.SkipSpaces()

	for s.index < max && s.err == nil {
		s.ScanSegment()
	}

	if s.err != nil {
		s.result = [][]interface{}{}

	} else if len(s.result) > 0 {

		if !(s.result[0][0] == "m" || s.result[0][0] == "M") {
			s.err = errors.Errorf("SvgPath: string should start with `M` or `m`")
			s.result = [][]interface{}{}
		} else {
			s.result[0][0] = "M"
		}
	}

	return s.result, s.err
}
