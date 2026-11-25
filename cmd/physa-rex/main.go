// Command physa-rex is Physa's views in the Rex terminal: `app ideas` and `app add` run a view inside a Rex layer, and `rex install` adds the actions and keys that open them.
package main

import (
	"fmt"
	"io"
	"os"

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
		person, perr := link.Person()
		blocked := ""
		if perr != nil {
			blocked = perr.Error()
		}
		src.Person = person
		model = views.NewAdd(src, profile, blocked)
	}
	_, err = tea.NewProgram(model).Run()
	return err
}

func install(stdout io.Writer) error {
	dir, err := rexlua.ConfigDir()
	if err != nil {
		return err
	}
	addCols, addRows := views.NewAdd(nil, colorprofile.TrueColor, "").NaturalSize()
	wrote, err := rexlua.Install(dir, rexlua.Sizes{IdeasCols: ideasCols, IdeasRows: ideasRows, AddCols: addCols, AddRows: addRows})
	for _, p := range wrote {
		fmt.Fprintln(stdout, "wrote", p)
	}
	if err != nil {
		return err
	}
	fmt.Fprintln(stdout, "Physa's key mode is ctrl+a p, then i for the ideas or a to add one. Change the key in physa.lua's settings block.")
	return nil
}
