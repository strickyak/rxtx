package qrss

import (
	"log"
)

type Tone byte

var W = ".--"
var SIX = "-...."
var R = ".-."
var E = "."
var K = "-.-"

var W6REK = []string{W, SIX, R, E, K}

func ExpandLetter(s string) []Tone {
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
	z = append(z, 0)
	return z
}

func ExpandWord(w []string) []Tone {
	var z []Tone
	for _, s := range w {
		z = append(z, ExpandLetter(s)...)
	}
	return z
}

func ExpandNested(w []string) []Tone {
	var z []Tone
	vec := ExpandWord(w)
	var skip bool
	for _, a := range vec {
		if skip {
			skip = false
			continue
		}
		if a > 0 {
			for _, b := range vec {
				if a > 0 && b > 0 {
					z = append(z, 3*a+b)
				} else {
					z = append(z, 0)
				}
			}
			skip = true
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
