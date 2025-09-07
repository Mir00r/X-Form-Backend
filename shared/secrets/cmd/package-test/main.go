package main

import (
	"fmt"

	"github.com/kamkaiz/x-form-backend/shared/secrets"
)

func main() {
	fmt.Println("Testing secrets package import...")

	// Just try to reference a type without creating anything
	var _ secrets.ProviderType

	fmt.Println("Secrets package imported successfully!")
}
