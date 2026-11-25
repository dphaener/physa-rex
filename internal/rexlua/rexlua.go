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
	"runtime"
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

// ConfigDir is the directory holding Rex's init.lua, where Rex looks for it:
// the directory of REX_CONFIG when it is set, else rex under XDG_CONFIG_HOME,
// else ~/.config/rex on macOS and Linux and rex under %AppData% on Windows.
// Go's os.UserConfigDir differs on macOS, so it is used only on Windows.
func ConfigDir() (string, error) {
	return configDir(runtime.GOOS, os.Getenv, os.UserHomeDir, os.UserConfigDir)
}

func configDir(goos string, getenv func(string) string, home, appData func() (string, error)) (string, error) {
	if c := getenv("REX_CONFIG"); c != "" {
		return filepath.Dir(c), nil
	}
	if x := getenv("XDG_CONFIG_HOME"); x != "" {
		return filepath.Join(x, "rex"), nil
	}
	if goos == "windows" {
		d, err := appData()
		if err != nil {
			return "", err
		}
		return filepath.Join(d, "rex"), nil
	}
	h, err := home()
	if err != nil {
		return "", err
	}
	return filepath.Join(h, ".config", "rex"), nil
}

// Bins are absolute paths for physa.lua's settings: Rex for bin, physa-rex
// itself, and Physa for physa, the CLI that starts the engine. An empty one
// leaves its setting alone.
type Bins struct {
	Rex, Physa string
}

// Install writes physa.lua into dir, keeping the settings block of the one
// already there, and adds the require line to init.lua when it is missing.
// A path in bins fills its setting only while the setting is empty, so a
// person's own value is kept. It returns the paths it wrote.
func Install(dir string, s Sizes, bins Bins) ([]string, error) {
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
	for _, f := range []struct{ name, value string }{{"bin", bins.Rex}, {"physa", bins.Physa}} {
		if lua, err = fill(lua, f.name, f.value); err != nil {
			return nil, err
		}
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

// fill sets the setting name to value in lua's settings block when it is
// empty, and adds it when an older block lacks it. An empty value changes
// nothing.
func fill(lua, name, value string) (string, error) {
	if value == "" {
		return lua, nil
	}
	quoted, err := luaString(value)
	if err != nil {
		return "", fmt.Errorf("%s: %w", name, err)
	}
	block := settingsBlock.FindString(lua)
	if block == "" {
		return lua, nil
	}
	line := regexp.MustCompile(`(?m)^(\s*)` + name + `\s*=\s*(.*?),?\s*$`)
	var out string
	if m := line.FindStringSubmatch(block); m != nil {
		if strings.TrimSpace(m[2]) != `""` {
			return lua, nil
		}
		out = line.ReplaceAllLiteralString(block, m[1]+name+" = "+quoted+",")
	} else {
		end := strings.LastIndex(block, "}")
		out = block[:end] + "  " + name + " = " + quoted + ",\n" + block[end:]
	}
	return strings.Replace(lua, block, out, 1), nil
}

// luaString quotes s as a Lua string literal. Only printable ASCII is
// written, so the quoting holds in every Lua version.
func luaString(s string) (string, error) {
	var b strings.Builder
	b.WriteByte('"')
	for _, r := range s {
		if r < 0x20 || r > 0x7e {
			return "", fmt.Errorf("%q has a character physa.lua can't hold; set it by hand", s)
		}
		if r == '"' || r == '\\' {
			b.WriteByte('\\')
		}
		b.WriteRune(r)
	}
	b.WriteByte('"')
	return b.String(), nil
}

var requireRe = regexp.MustCompile(`(?m)^\s*require\s*\(?\s*["']physa["']`)

func requires(init string) bool { return requireRe.MatchString(init) }
