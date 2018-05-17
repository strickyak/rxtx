// +build main

package main

import "bufio"
import "fmt"
import "os"
import "github.com/strickyak/rxtx/qrss"

func main() {

/*
	fmt.Printf("%d\n", len(qrss.ExpandNested(qrss.W6REK)))
	fmt.Printf("%d\n", len(qrss.ExpandWord(qrss.W6REK)))
*/
	fmt.Fprintf(os.Stderr, "%v\n", qrss.ExpandNested(qrss.W6REK))
	fmt.Fprintf(os.Stderr, "%v\n", qrss.ExpandWord(qrss.W6REK))

/*
	for _, c := range qrss.Play(qrss.ExpandNested(qrss.W6REK)) {
		fmt.Printf("%c", c)
		os.Stdout.W
	}
*/

	bb := qrss.Play(qrss.ExpandNested(qrss.W6REK))
	w := bufio.NewWriter(os.Stdout)
	w.Write(bb)
	w.Flush()
}
