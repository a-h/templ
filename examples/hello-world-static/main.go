package main

import (
	"context"
	"log"
	"os"
)

func main() {
	component := hello("John")
	if err := component.Render(context.Background(), os.Stdout); err != nil {
		log.Fatalf("failed to render: %v", err)
	}
}
