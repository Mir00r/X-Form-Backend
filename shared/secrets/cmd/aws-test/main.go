package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
)

func main() {
	fmt.Println("Testing AWS SDK import...")
	ctx := context.Background()
	_, err := config.LoadDefaultConfig(ctx)
	fmt.Printf("AWS SDK test result: %v\n", err)
	fmt.Println("AWS SDK import test completed!")
}
