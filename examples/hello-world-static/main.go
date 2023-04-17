package main

import (
	"context"
	"os"
)

func main() {
	component := hello("John")
	component.Render(context.Background(), os.Stdout)
}
