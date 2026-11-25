// Package rexlua puts Physa's views into Rex: it writes physa.lua beside Rex's init.lua, with an action per view and a key mode to open them, keeps the person's settings block across reinstalls, and requires the file from init.lua once.
package rexlua

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/dphaener/rexuikit/surface"
)

//go:embed physa.lua.tmpl
var source string

var tmpl = template.Must(template.New("physa.lua").Parse(source))

// requireLine is what Install adds to init.lua.
const requireLine = `require("physa") -- added by physa-rex rex install`

// settingsBlock matches the settings block, markers included.
var settingsBlock = regexp.MustCompile(`(?s)-- settings begin\n.*?-- settings end\n`)

// Sizes are the usable cells each view's layer opens with.
type Sizes struct {
	IdeasCols, IdeasRows int
	AddCols, AddRows     int
}

// Render returns physa.lua for sizes, with rexuikit's layer chrome.
func Render(s Sizes) (string, error) {
	var b bytes.Buffer
	err := tmpl.Execute(&b, map[string]int{
		"FramePxCols": surface.FramePxCols, "FramePxRows": surface.FramePxRows,
		"BorderPxCols": surface.BorderPxCols, "BorderPxRows": surface.BorderPxRows,
		"MarginPxCols": surface.InverseMarginPxCols, "MarginPxRows": surface.InverseMarginPxRows,
		"IdeasCols": s.IdeasCols, "IdeasRows": s.IdeasRows,
		"AddCols": s.AddCols, "AddRows": s.AddRows,
	})
	return b.String(), err
}

// ConfigDir is the directory holding Rex's init.lua: the directory of
// REX_CONFIG when it is set, else rex under the user's config directory.
func ConfigDir() (string, error) {
	if c := os.Getenv("REX_CONFIG"); c != "" {
		return filepath.Dir(c), nil
	}
	d, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, "rex"), nil
}

// Install writes physa.lua into dir, keeping the settings block of the one
// already there, and adds the require line to init.lua when it is missing.
// It returns the paths it wrote.
func Install(dir string, s Sizes) ([]string, error) {
	lua, err := Render(s)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	path := filepath.Join(dir, "physa.lua")
	old, err := os.ReadFile(path)
	switch {
	case err == nil:
		if kept := settingsBlock.Find(old); kept != nil {
			lua = settingsBlock.ReplaceAllLiteralString(lua, string(kept))
		}
	case !errors.Is(err, fs.ErrNotExist):
		return nil, err
	}
	if err := os.WriteFile(path, []byte(lua), 0o644); err != nil {
		return nil, err
	}
	wrote := []string{path}

	initPath := filepath.Join(dir, "init.lua")
	init, err := os.ReadFile(initPath)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return wrote, err
	}
	if requires(string(init)) {
		return wrote, nil
	}
	text := string(init)
	if text != "" && !strings.HasSuffix(text, "\n") {
		text += "\n"
	}
	if err := os.WriteFile(initPath, []byte(text+requireLine+"\n"), 0o644); err != nil {
		return wrote, fmt.Errorf("add the require to %s: %w", initPath, err)
	}
	return append(wrote, initPath), nil
}

var requireRe = regexp.MustCompile(`(?m)^\s*require\s*\(?\s*["']physa["']`)

func requires(init string) bool { return requireRe.MatchString(init) }
