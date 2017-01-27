package mulaw

import (
	"fmt"
	"io/ioutil"
	"strings"
	"testing"
)

// Confirms the encode/decode still match the golden mulaw.expect file.
func TestMulaw16(t *testing.T) {
	guts, err := ioutil.ReadFile("mulaw.expect")
	if err != nil {
		panic(err)
	}
	lines := strings.Split(string(guts), "\n")
	lines = lines[0 : len(lines)-1]

	for _, line := range lines {
		// println("LINE <", line, ">")
		var a, b, c int16
		n, err := fmt.Sscanf(line, "%d %d %d", &a, &b, &c)
		if err != nil {
			panic(err)
		}
		if n != 3 {
			panic(n)
		}

		x := EncodeMulaw16(a)
		y := DecodeMulaw16(x)
		if int16(x) != b || y != c {
			t.Errorf("for %d got %d,%d expected %d,%d", a, x, y, b, c)
		}

	}
}
