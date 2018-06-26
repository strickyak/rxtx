package qrss

import (
	"log"
)

type FadeEnd byte

const (
	Both = FadeEnd(iota)
	Left
	Right
	Silent
)

type TonePair struct {
	A, B Tone
	Fade FadeEnd
}

// ExpandLetter returns pairs of tones.
func SlantedExpandLetter(s string, final bool) []TonePair {
	var z []TonePair
	for _, c := range s {
		switch c {
		case '.':
			z = append(z, TonePair{1, 2, Both})
		case '-':
			z = append(z, TonePair{2, 1, Both})
		default:
			log.Fatalf("Should just be dots and dashes: %q", s)
		}
	}
	if !final {
		z = append(z, TonePair{1.5, 1.5, Both})
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

func DuoExpandLetter(s string) []TonePair {
	var z []TonePair
	for _, c := range s {
		switch c {
		case '.':
			// z = append(z, TonePair{1.0, 1.0, Left}, TonePair{1.0, 0.1, Right})
			z = append(z, TonePair{1.0, 1.0, Left}, TonePair{1.0, 0.5, Right})
		case '-':
			// z = append(z, TonePair{2.0, 2.0, Left}, TonePair{2.0, 2.9, Right})
			z = append(z, TonePair{2.0, 2.0, Left}, TonePair{2.0, 2.5, Right})
		default:
			log.Fatalf("Should just be dots and dashes: %q", s)
		}
	}
	return z
}

func DuoExpandWord(w []string) []TonePair {
	var z []TonePair
	// z = append(z, TonePair{2.0, 0.75, Left}, TonePair{0.75, -0.5, Right})
	n := len(w)
	for i, s := range w {
		z = append(z, DuoExpandLetter(s)...)
		if i < n-1 {
			// z = append(z, TonePair{0.5, 1.5, Left}, TonePair{1.5, 2.5, Right})
			// z = append(z, TonePair{1.5, 1.5, Left}, TonePair{1.5, 1.5, Right})
			z = append(z, TonePair{}, TonePair{}, TonePair{})
		}
	}
	// z = append(z, TonePair{0.5, 2.5, Left}, TonePair{0.5, 2.5, Right})
	return z
}

func NeoExpandLetter(s string) []TonePair {
	var z []TonePair
	for _, c := range s {
		switch c {
		case '.':
			z = append(z, TonePair{1, 1, Both})
		case '-':
			z = append(z, TonePair{2, 2, Both})
		default:
			log.Fatalf("Should just be dots and dashes: %q", s)
		}
	}
	return z
}

func NeoExpandWord(w []string) []TonePair {
	var z []TonePair
	z = append(z, TonePair{2.0, -0.5, Both})
	n := len(w)
	for i, s := range w {
		z = append(z, NeoExpandLetter(s)...)
		if i < n-1 {
			// z = append(z, TonePair{0.1, 2.1, Both})
			// z = append(z, TonePair{-1.0, 0.1, Both})
			z = append(z, TonePair{-1.0, -1.0, Both})
		}
	}
	z = append(z, TonePair{-2.0, 1.0, Both})
	return z
}
