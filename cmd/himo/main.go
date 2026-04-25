package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/paopp2/himo/internal/cli"
	"github.com/paopp2/himo/internal/config"
	"github.com/paopp2/himo/internal/state"
	"github.com/paopp2/himo/internal/store"
	"github.com/paopp2/himo/internal/tui"
)

const version = "0.0.1-dev"

func main() {
	if len(os.Args) < 2 {
		openTUIWithConfig()
		return
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
	case "add":
		fs := flag.NewFlagSet("add", flag.ExitOnError)
		project := fs.String("p", "", "project name (default: default_project)")
		fs.Parse(os.Args[2:])
		if fs.NArg() < 1 {
			fmt.Fprintln(os.Stderr, "usage: himo add [-p project] \"<title>\"")
			os.Exit(1)
		}
		cfg, err := loadConfigOrExit()
		if err != nil {
			os.Exit(1)
		}
		name := *project
		if name == "" {
			name = cfg.DefaultProject
		}
		if name == "" {
			fmt.Fprintln(os.Stderr, "himo add: no project (-p) and no default_project set")
			os.Exit(1)
		}
		if err := cli.AddTask(cfg.BaseDir, name, fs.Arg(0)); err != nil {
			fmt.Fprintln(os.Stderr, "himo add:", err)
			os.Exit(1)
		}
	case "ls":
		fs := flag.NewFlagSet("ls", flag.ExitOnError)
		project := fs.String("p", "", "project name (default: all projects)")
		status := fs.String("s", "", "filter by status (pending, active, blocked, backlog, done, cancelled)")
		fs.Parse(os.Args[2:])
		cfg, err := loadConfigOrExit()
		if err != nil {
			os.Exit(1)
		}
		if err := cli.Ls(cfg.BaseDir, *project, *status, os.Stdout); err != nil {
			fmt.Fprintln(os.Stderr, "himo ls:", err)
			os.Exit(1)
		}
	default:
		if strings.HasPrefix(os.Args[1], "-") {
			fmt.Fprintln(os.Stderr, "himo: unknown flag", os.Args[1])
			os.Exit(1)
		}
		cfg, err := loadConfigOrExit()
		if err != nil {
			os.Exit(1)
		}
		openTUI(cfg, os.Args[1], loadState())
	}
}

func openTUIWithConfig() {
	cfg, err := loadConfigOrFirstRun()
	if err != nil {
		os.Exit(1)
	}
	st := loadState()
	name := st.LastProject
	if name == "" {
		name = cfg.DefaultProject
	}
	if name == "" {
		names, err := store.ListProjects(cfg.BaseDir)
		if err != nil {
			fmt.Fprintln(os.Stderr, "himo:", err)
			os.Exit(1)
		}
		if len(names) == 0 {
			fmt.Fprintln(os.Stderr, "himo: no projects. Run `himo new <name>`.")
			os.Exit(1)
		}
		name = names[0]
	}
	openTUI(cfg, name, st)
}

func openTUI(cfg *config.Config, project string, st *state.State) {
	m, err := tui.NewModelFromBase(cfg.BaseDir, project, tui.StyleOptions{
		AsciiGlyphs: cfg.AsciiGlyphs,
		NoColor:     cfg.NoColor,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "himo:", err)
		os.Exit(1)
	}
	if st.LastFilter != "" {
		m = m.WithFilter(tui.FilterFromName(st.LastFilter))
	}
	if st.LastSort != "" {
		m = m.WithSort(tui.SortFromName(st.LastSort))
	}
	if st.LastAllProjects {
		m = m.WithAllProjects()
	}
	final, err := tea.NewProgram(m, tea.WithAltScreen()).Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, "himo:", err)
		os.Exit(1)
	}
	if fm, ok := final.(tui.Model); ok {
		if p := fm.SessionProject(); p != "" {
			st.LastProject = p
		}
		if f := fm.SessionFilter(); f != "" {
			st.LastFilter = f
		}
		if s := fm.SessionSort(); s != "" {
			st.LastSort = s
		}
		st.LastAllProjects = fm.SessionAllProjects()
		if err := state.Save(st); err != nil {
			fmt.Fprintln(os.Stderr, "himo: save state:", err)
		}
	}
}

// loadState reads the on-disk session state, logging and defaulting to
// an empty State on any error so a bad state file never blocks launch.
func loadState() *state.State {
	st, err := state.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "himo: load state:", err)
		return &state.State{}
	}
	return st
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

func loadConfigOrFirstRun() (*config.Config, error) {
	path, err := config.DefaultPath()
	if err != nil {
		fmt.Fprintln(os.Stderr, "himo: resolve config path:", err)
		return nil, err
	}
	cfg, err := config.Load(path)
	if os.IsNotExist(err) {
		if err := cli.Init(os.Stdin, os.Stdout); err != nil {
			fmt.Fprintln(os.Stderr, "himo init:", err)
			return nil, err
		}
		cfg, err = config.Load(path)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "himo: load config:", err)
		return nil, err
	}
	return cfg, nil
}
