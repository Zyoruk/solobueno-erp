package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("Solobueno ERP Migration Tool")

	if len(os.Args) < 2 {
		fmt.Println("Usage: migrate [up|down|status]")
		os.Exit(1)
	}

	command := os.Args[1]
	fmt.Printf("Command: %s\n", command)
	fmt.Println("TODO: Implement migration commands in future features")
}
