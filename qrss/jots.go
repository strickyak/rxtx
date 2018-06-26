package qrss

import (
	"log"
)

/*
|             _ z
|  c \  / b   _ y
|  d  \/  a   _ x
*/

var jW = "cdadab"
var jSIX = "abxyza"
var jR = "abyzdb"
var jE = "abxyzxz"
var jK = "abdb"

var J_W6REK = []string{jW, jSIX, jR, jE, jK}

// ExpandLetter outputs a list of tones: 1 for a dit, 2 for a dah, or 0 for a gap.
func JotsExpandLetter(s string, final bool) []TonePair {
	var z []TonePair
	for _, c := range s {
		switch c {
		case 'a':
			z = append(z, TonePair{1, 2, Both})
		case 'b':
			z = append(z, TonePair{2, 3, Both})
		case 'c':
			z = append(z, TonePair{3, 2, Both})
		case 'd':
			z = append(z, TonePair{2, 1, Both})

		case 'x':
			z = append(z, TonePair{1, 1, Both})
		case 'y':
			z = append(z, TonePair{2, 2, Both})
		case 'z':
			z = append(z, TonePair{3, 3, Both})

		default:
			log.Fatalf("Should just be dots and dashes: %q", s)
		}
	}
	if !final {
		z = append(z, TonePair{0, 0, Silent})
	}
	return z
}

// ExpandWord outputs tones for the word, with 0 gaps between letters but not initially or finally.
func JotsExpandWord(w []string) []TonePair {
	var z []TonePair
	n := len(w)
	for i, s := range w {
		z = append(z, JotsExpandLetter(s, (i+1 == n))...)
	}
	return z
}

func PrintJots(w []string) {
	var p, q, r string
	for _, s := range w {
		for _, c := range s {
			switch c {
			case 'a':
				p += " "
				q += " "
				r += "/"
			case 'b':
				p += " "
				q += "/"
				r += " "
			case 'c':
				p += " "
				q += "\\"
				r += " "
			case 'd':
				p += " "
				q += " "
				r += "\\"

			case 'x':
				p += " "
				q += " "
				r += "_"
			case 'y':
				p += " "
				q += "_"
				r += " "
			case 'z':
				p += "_"
				q += " "
				r += " "

			default:
				log.Fatalf("Should just be dots and dashes: %q", s)
			}
		}
		p += " "
		q += " "
		r += " "
		p += " "
		q += " "
		r += " "
	}

	log.Println("###")
	log.Println("###", p)
	log.Println("###", q)
	log.Println("###", r)
	log.Println("###")
}
