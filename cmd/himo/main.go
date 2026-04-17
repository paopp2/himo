package main

import (
	"fmt"
	"os"

	"github.com/npaolopepito/himo/internal/cli"
	"github.com/npaolopepito/himo/internal/config"
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
	case "new":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "usage: himo new <project>")
			os.Exit(1)
		}
		cfg, err := loadConfigOrExit()
		if err != nil {
			os.Exit(1)
		}
		if err := cli.NewProject(cfg.BaseDir, os.Args[2]); err != nil {
			fmt.Fprintln(os.Stderr, "himo new:", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintln(os.Stderr, "himo: unknown command", os.Args[1])
		os.Exit(1)
	}
}

func loadConfigOrExit() (*config.Config, error) {
	path, err := config.DefaultPath()
	if err != nil {
		fmt.Fprintln(os.Stderr, "himo: resolve config path:", err)
		return nil, err
	}
	cfg, err := config.Load(path)
	if os.IsNotExist(err) {
		fmt.Fprintln(os.Stderr, "himo: no config found. Run `himo init` first.")
		return nil, err
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "himo: load config:", err)
		return nil, err
	}
	return cfg, nil
}
