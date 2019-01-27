package qrss

import (
	"log"
)

type Tone float64

var W = ".--"
var SIX = "-...."
var R = ".-."
var E = "."
var K = "-.-"

var SLASH = "-..-."
var FOUR = "....-"
var A = ".-"
var T = "-"
var L = ".-.."
var TWO = "..---"
var H = "...."

var W6REK = []string{W, SIX, R, E, K}
var W6REK_4 = []string{W, SIX, R, E, K, SLASH, FOUR}
var W6REK_4_ATL = []string{W, SIX, R, E, K, SLASH, FOUR, SLASH, A, T, L}

var W2H = []string{W, TWO, H}

// ExpandLetter outputs a list of tones: 1 for a dit, 2 for a dah, or 0 for a gap.
func ExpandLetter(s string, final bool) []Tone {
	var z []Tone
	for _, c := range s {
		switch c {
		case '.':
			z = append(z, 1)
		case '-':
			z = append(z, 2)
		default:
			log.Fatalf("Should just be dots and dashes: %q", s)
		}
	}
	if !final {
		z = append(z, 0)
	}
	return z
}

// ExpandWord outputs tones for the word, with 0 gaps between letters but not initially or finally.
func ExpandWord(w []string) []Tone {
	var z []Tone
	n := len(w)
	for i, s := range w {
		z = append(z, ExpandLetter(s, (i+1 == n))...)
	}
	return z
}

// This was experimental and is not used now.
// It made a fractal outer ID using tones (1, 2) for lower inner ID, and (4, 5) for higher inner ID, and 0 for gaps.
// It leaves no gap between repeated outer elements, which is a problem reading the inner IDs when they run together.
func ExpandNested(w []string) []Tone {
	var z []Tone
	vec := ExpandWord(w)
	for _, a := range vec {
		if a > 0 {
			for _, b := range vec {
				if a > 0 && b > 0 {
					z = append(z, 3*a+b)
				} else {
					z = append(z, 0)
				}
			}
		} else {
			for i, _ := range vec {
				if (i % 2) == 0 {
					z = append(z, 0)
				}
			}
		}
	}
	return z
}
