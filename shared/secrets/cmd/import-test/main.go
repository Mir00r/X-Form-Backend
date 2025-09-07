package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("Testing basic imports...")
	os.Setenv("TEST", "value")
	fmt.Println("Basic imports work!")
}
