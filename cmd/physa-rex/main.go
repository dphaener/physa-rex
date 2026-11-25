// Command physa-rex is Physa's views in the Rex terminal: `app ideas` and `app add` run a view inside a Rex layer, and `rex install` adds the actions and keys that open them.
package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/colorprofile"

	"github.com/dphaener/physa-rex/internal/link"
	"github.com/dphaener/physa-rex/internal/rexlua"
	"github.com/dphaener/physa-rex/internal/views"
)

const usage = `usage: physa-rex app ideas   the repo's ideas, and each idea's page
       physa-rex app add     add an idea; it belongs to you
       physa-rex rex install add Physa's actions and keys to Rex

The views show the repo of the directory they open in, and talk to the
machine's Physa engine, starting it with the physa CLI when it isn't running.
`

// ideasCols and ideasRows are the ideas layer's size: room for an idea and
// its facts on one row, and a page without much scrolling.
const ideasCols, ideasRows = 90, 24

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "physa-rex:", err)
		os.Exit(1)
	}
}

func run(args []string, stdout io.Writer) error {
	switch {
	case len(args) == 2 && args[0] == "app":
		return app(args[1])
	case len(args) == 2 && args[0] == "rex" && args[1] == "install":
		return install(stdout)
	case len(args) == 1 && (args[0] == "-h" || args[0] == "--help" || args[0] == "help"):
		fmt.Fprint(stdout, usage)
		return nil
	}
	return fmt.Errorf("unknown command\n\n%s", usage)
}

func app(name string) error {
	profile := colorprofile.Detect(os.Stdout, os.Environ())
	if name != "ideas" && name != "add" {
		return fmt.Errorf("no view named %q\n\n%s", name, usage)
	}
	repo, err := link.Repo()
	if err != nil {
		return err
	}
	src := &views.Engine{Dial: link.Connect, Repo: repo}
	defer src.Close()
	var model tea.Model
	switch name {
	case "ideas":
		model = views.NewIdeas(src, profile)
	case "add":
		model = views.NewAdd(src, profile)
	}
	_, err = tea.NewProgram(model).Run()
	return err
}

func install(stdout io.Writer) error {
	dir, err := rexlua.ConfigDir()
	if err != nil {
		return err
	}
	addCols, addRows := views.NewAdd(nil, colorprofile.TrueColor).NaturalSize()
	bins, notes := loginBins()
	wrote, err := rexlua.Install(dir, rexlua.Sizes{IdeasCols: ideasCols, IdeasRows: ideasRows, AddCols: addCols, AddRows: addRows}, bins)
	for _, p := range wrote {
		fmt.Fprintln(stdout, "wrote", p)
	}
	if err != nil {
		return err
	}
	for _, n := range notes {
		fmt.Fprintln(stdout, n)
	}
	fmt.Fprintln(stdout, "Run rex config reload to load it.")
	fmt.Fprintln(stdout, "Physa's key mode is ctrl+a y, then i for the ideas or a to add one. Change the key in physa.lua's settings block.")
	return nil
}

// loginBins finds what a Rex layer can't: Rex starts the views through a
// login shell, which reads the login profile but not an interactive shell's
// configuration, where tools such as mise often add to PATH. For physa-rex
// and physa that such a shell can't find, it returns the absolute paths
// found now, with a note for each.
func loginBins() (rexlua.Bins, []string) {
	var bins rexlua.Bins
	var notes []string
	if !loginFinds("physa-rex") {
		exe, err := os.Executable()
		if err == nil {
			exe, err = filepath.EvalSymlinks(exe)
		}
		switch {
		case err != nil:
			notes = append(notes, "A login shell can't find physa-rex. Set bin in physa.lua's settings to its path.")
		case strings.Contains(exe, "go-build"):
			notes = append(notes, "A login shell can't find physa-rex, and this one is a go run build. Install it, then run physa-rex rex install again.")
		default:
			bins.Rex = exe
			notes = append(notes, "A login shell can't find physa-rex, so physa.lua runs "+exe+".")
		}
	}
	if !loginFinds("physa") {
		if p, err := exec.LookPath("physa"); err == nil {
			if abs, err := filepath.Abs(p); err == nil {
				bins.Physa = abs
				notes = append(notes, "A login shell can't find physa, so the views start the engine with "+abs+".")
			}
		} else {
			notes = append(notes, "physa isn't on PATH. Install Physa's CLI, then run physa-rex rex install again.")
		}
	}
	return bins, notes
}

// loginFinds reports whether a login shell, started the way Rex starts a
// layer, finds name on its PATH. It starts from a bare environment, so the
// PATH of the shell running the install doesn't leak in.
func loginFinds(name string) bool {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, shell, "-lc", `command -v "$0" >/dev/null 2>&1`, name)
	cmd.Env = []string{"HOME=" + os.Getenv("HOME"), "USER=" + os.Getenv("USER"), "LOGNAME=" + os.Getenv("LOGNAME"),
		"SHELL=" + shell, "PATH=/usr/bin:/bin:/usr/sbin:/sbin"}
	return cmd.Run() == nil
}
