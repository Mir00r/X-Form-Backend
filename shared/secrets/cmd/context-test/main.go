package main

import (
	"context"
	"fmt"
)

func main() {
	fmt.Println("Testing context import...")
	ctx := context.Background()
	_ = ctx
	fmt.Println("Context import works!")
}
