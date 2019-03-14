package textimage

import (
	"fmt"
	"image/color"
	"strings"
)

func Hex(col string) (color.RGBA, error) {
	format := "#%02x%02x%02x"
	factor := uint8(1)
	if len(col) == 4 {
		format = "#%1x%1x%1x"
		factor = uint8(17)
	}
	var r, g, b uint8
	n, err := fmt.Sscanf(col, format, &r, &g, &b)

	if err != nil {
		return color.RGBA{}, err
	}
	if n != 3 {
		return color.RGBA{}, fmt.Errorf("color: %v is not a hex-color", col)
	}
	return color.RGBA{r * factor, g * factor, b * factor, 255}, nil
}

func wordWrapUtil(t string, n int) string {
	s := strings.Split(t, "\n")
	var newString []string
	for _, v := range s {
		str := spliteString(v, n)
		newString = append(newString, str...)
	}

	return strings.Join(newString, "\n")
}

func spliteString(t string, n int) []string {
	var front string
	var end string
	var newS []string
	if len([]rune(t)) >= n {
		front = string([]rune(t)[:n])
		end = string([]rune(t)[n:])
		newS = append(newS, front)
		aa := spliteString(end, n)
		newS = append(newS, aa...)
	} else {
		newS = append(newS, t)
	}
	return newS
}
