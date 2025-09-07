package main

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

func main() {
	fmt.Println("Testing logrus import...")
	logger := logrus.New()
	logger.Info("Logrus works!")
	fmt.Println("Logrus import works!")
}
