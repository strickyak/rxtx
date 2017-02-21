package rxtx

import (
	"fmt"
	"log"
)

func Show(a interface{}) string {
	return fmt.Sprintf("%#v ", a)
}

func Assert(ok bool, a ...interface{}) {
	if !ok {
		log.Fatalln("Assertion Failed", fmt.Sprintln(a...))
	}
}

func Check(err error, a ...interface{}) {
	if err != nil {
		log.Fatalln(fmt.Sprintf("Check Failed: %v", err), fmt.Sprintln(a...))
	}
}
