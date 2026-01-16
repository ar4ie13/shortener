package main

import (
	"log"
	"os"
)

func helper() {
	log.Fatal("not in main") // want `forbidden call to log\.Fatal outside main\(\) in package main`
	os.Exit(1)               // want `forbidden call to os\.Exit outside main\(\) in package main`
}

func main() {
	helper()
	panic("even here it's forbidden") // want `forbidden use of panic\(\)`
}
