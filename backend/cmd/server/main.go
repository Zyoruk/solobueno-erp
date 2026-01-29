package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("Solobueno ERP Server")
	fmt.Println("Version: 0.0.1")
	fmt.Printf("Go version: %s\n", "1.22+")

	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Server would start on port %s\n", port)
	fmt.Println("TODO: Implement server startup in future features")
}
