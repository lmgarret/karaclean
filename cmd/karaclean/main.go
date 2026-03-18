package main

import (
	"fmt"
	"os"

	"github.com/lm/karaclean/internal/config"
)

func main() {
	path := config.ResolvePath("")
	_, err := config.Load(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("config loaded successfully")
}
