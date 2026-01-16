package main

import (
	"log"
	"os"
)

func main() {
	log.Println("Starting...")
	os.Exit(0)              // OK: inside main()
	log.Fatal("Final exit") // OK: inside main()
}
