package qrss

import (
	"log"
)

type TonePair struct {
	A, B Tone
}

// ExpandLetter returns pairs of tones.
func SlantedExpandLetter(s string, final bool) []TonePair {
	var z []TonePair
	for _, c := range s {
		switch c {
		case '.':
			z = append(z, TonePair{1, 2})
		case '-':
			z = append(z, TonePair{2, 1})
		default:
			log.Fatalf("Should just be dots and dashes: %q", s)
		}
	}
	if !final {
		z = append(z, TonePair{1.5, 1.5})
	}
	return z
}

func SlantedExpandWord(w []string) []TonePair {
	var z []TonePair
	n := len(w)
	for i, s := range w {
		z = append(z, SlantedExpandLetter(s, (i+1 == n))...)
	}
	return z
}
