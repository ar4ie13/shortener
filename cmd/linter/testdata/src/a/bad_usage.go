package a

import (
	"log"
	"os"
)

func BadFunc() {
	panic("this is forbidden")  // want `forbidden use of panic\(\)`
	log.Fatal("also forbidden") // want `forbidden call to log\.Fatal in non-main package`
	os.Exit(42)                 // want `forbidden call to os\.Exit in non-main package`
}
