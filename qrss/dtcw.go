package qrss

import (
	"flag"
	"log"
)

var WITH_TAILS = flag.Bool("tail", false, "with experimental tails between letters")

type Tone int16

var W = ".--"
var SIX = "-...."
var R = ".-."
var E = "."
var K = "-.-"

var W6REK = []string{W, SIX, R, E, K}

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
		if *WITH_TAILS {
			z = append(z, -1)
		} else {
			z = append(z, 0)
		}
	}
	return z
}

func ExpandWord(w []string) []Tone {
	var z []Tone
	n := len(w)
	for i, s := range w {
		z = append(z, ExpandLetter(s, (i+1 == n))...)
	}
	return z
}

func ExpandNested(w []string) []Tone {
	var z []Tone
	vec := ExpandWord(w)
	for _, a := range vec {
		if a > 0 || *WITH_TAILS {
			for _, b := range vec {
				if a > 0 && b > 0 || *WITH_TAILS {
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
