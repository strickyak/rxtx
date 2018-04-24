// +build main

/*
   Usage:    go run demo.go "The quick brown fox" | less
*/
package main

import (
	"github.com/strickyak/rxtx/font5x7"

	"fmt"
	"os"
)

var printf = fmt.Printf

func render(bitmap []byte) {
	for _, row := range bitmap {
		var mask byte = 0x10
		for j := 0; j < 5; j++ {
			if 0 != (mask & row) {
				printf(" #")
			} else {
				printf(" _")
			}

			mask >>= 1
		}
		printf("\n")
	}
}

func main() {
	s := os.Args[1] // String to print.

	bitmap := font5x7.VerticalStringFiveBitsWide(s)
	render(bitmap)
}
