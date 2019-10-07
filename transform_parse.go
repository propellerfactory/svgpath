// port of https://github.com/fontello/svgpath
package svgpath

import (
	"regexp"
	"strconv"
)

var CMD_SPLIT_RE = regexp.MustCompile(`\s*(matrix|translate|scale|rotate|skewX|skewY)\s*\(\s*(.+?)\s*\)[\s,]*`)
var PARAMS_SPLIT_RE = regexp.MustCompile(`[\s,]+`)

func RegSplit(reg *regexp.Regexp, text string) []string {
	matches := reg.FindAllStringSubmatch(text, -1)
	result := []string{}
	for _, match := range matches {
		for i := 1; i < len(match); i++ {
			result = append(result, match[i])
		}
	}
	return result
}

var operations = map[string]bool{
	"matrix":    true,
	"scale":     true,
	"rotate":    true,
	"translate": true,
	"skewX":     true,
	"skewY":     true,
}

func TransformParse(transformString string) *Matrix {
	matrix := NewMatrix()

	var cmd string
	// Split value into ['', 'translate', '10 50', '', 'scale', '2', '', 'rotate',  '-45', '']
	items := RegSplit(CMD_SPLIT_RE, transformString) // CMD_SPLIT_RE.Split(transformString, -1)
	for _, item := range items {
		// Skip empty elements
		if len(item) == 0 {
			continue
		}

		// remember operation
		if _, isOperation := operations[item]; isOperation {
			cmd = item
			continue
		}

		// extract params & att operation to matrix
		paramStrings := PARAMS_SPLIT_RE.Split(item, -1)
		params := make([]float64, len(paramStrings))
		for i, param := range paramStrings {
			value, err := strconv.ParseFloat(param, 64)
			if err == nil {
				params[i] = value
			} else {
				params[i] = 0.0
			}
		}

		// If params count is not correct - ignore command
		switch cmd {
		case "matrix":
			if len(params) == 6 {
				matrix.Matrix(params)
			}
		case "scale":
			if len(params) == 1 {
				matrix.Scale(params[0], params[0])
			} else if len(params) == 2 {
				matrix.Scale(params[0], params[1])
			}
		case "rotate":
			if len(params) == 1 {
				matrix.Rotate(params[0], 0, 0)
			} else if len(params) == 3 {
				matrix.Rotate(params[0], params[1], params[2])
			}
		case "translate":
			if len(params) == 1 {
				matrix.Translate(params[0], 0)
			} else if len(params) == 2 {
				matrix.Translate(params[0], params[1])
			}
		case "skewX":
			if len(params) == 1 {
				matrix.SkewX(params[0])
			}
		case "skewY":
			if len(params) == 1 {
				matrix.SkewY(params[0])
			}
		}
	}
	return matrix
}
