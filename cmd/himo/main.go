package main

import (
	"fmt"
	"os"

	"github.com/npaolopepito/himo/internal/cli"
)

const version = "0.0.1-dev"

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: himo <init|new|add|ls> [args...]")
		os.Exit(1)
	}
	switch os.Args[1] {
	case "--version", "-v":
		fmt.Println(version)
	case "init":
		if err := cli.Init(os.Stdin, os.Stdout); err != nil {
			fmt.Fprintln(os.Stderr, "himo init:", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintln(os.Stderr, "himo: unknown command", os.Args[1])
		os.Exit(1)
	}
}
