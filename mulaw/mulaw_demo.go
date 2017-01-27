// +build main

/*
  This will generate mulaw.expect,
  which you must eyeball to make sure it is right.
*/
package main

import "github.com/strickyak/rxtx/mulaw"
import "fmt"

func main() {
	var i int16
	for i = 0; i < 8160; i++ {
		x := mulaw.EncodeMulaw16(i)
		y := mulaw.DecodeMulaw16(x)
		fmt.Printf("%5d\t%3d\t%5d\n", i, x, y)
	}
	for i = 0; i < 8160; i++ {
		x := mulaw.EncodeMulaw16(-i)
		y := mulaw.DecodeMulaw16(x)
		fmt.Printf("%5d\t%3d\t%5d\n", -i, x, y)
	}
}
