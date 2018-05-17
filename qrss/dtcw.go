package qrss

/* My personal plan for Dual Tone Continuous Wave:
	Dit and Dah are both 2 time units.
	One unit gap between repeated dits or dahs.
	Three unit gap between letters.
*/

var W = ".--"
var SIX = "-...."
var R = ".-."
var E = "."
var K = "-.-"

var W6REK = []string { W, SIX, R, E, K }

func ExpandLetter(s string) []byte {
	var z []byte
	prev := 'x'
	for _, c := range s {
		if c == prev {
		  z = append(z, 0)
		}
		switch c {
		case '.': z = append(z, 1, 1)
		case '-': z = append(z, 2, 2)
		default: panic(s)
		}
		prev = c
	}
	z = append(z, 0, 0, 0)
	return z
}

func ExpandWord(w []string) []byte {
	var z []byte
	for _, s := range w {
		z = append(z, ExpandLetter(s)...)
	}
	return z
}

func ExpandNested(w []string) []byte {
	var z []byte
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
					z = append(z, 3*a + b)
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
