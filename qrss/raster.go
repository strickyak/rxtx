package qrss

import (
	font "github.com/strickyak/rxtx/font5x7"
)

func RasterExpandLetter(a rune) []TonePair {
	var z []TonePair
	for c := 0; c < 5; c++ {
		for r := 7; r >= 0; r-- {
			if font.Pixel(byte(a), r, c) {
				z = append(z, TonePair{Tone(8 - r), Tone(8 - r), Both})
			} else {
				z = append(z, TonePair{0, 0, Silent})
			}
		}
	}
	return z
}

func RasterExpandWord(s string) []TonePair {
	var z []TonePair
	for i, a := range s {
		if i != 0 {
			for j := 0; j < 15; j++ {
				z = append(z, TonePair{0, 0, Silent})
			}
		}
		z = append(z, RasterExpandLetter(a)...)
	}
	return z
}
