package rexlua

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var sizes = Sizes{IdeasCols: 90, IdeasRows: 24, AddCols: 60, AddRows: 9}

func TestRenderFillsEveryValue(t *testing.T) {
	lua, err := Render(sizes)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(lua, "{{") || strings.Contains(lua, "<no value>") {
		t.Fatalf("a template value is missing:\n%s", lua)
	}
	for _, want := range []string{`"physa_ideas"`, `"physa_add_idea"`, `90, 24, "app", "ideas"`, `60, 9, "app", "add"`} {
		if !strings.Contains(lua, want) {
			t.Errorf("physa.lua lacks %s", want)
		}
	}
}

func TestInstallKeepsSettingsAndRequiresOnce(t *testing.T) {
	dir := t.TempDir()
	initLua := filepath.Join(dir, "init.lua")
	if err := os.WriteFile(initLua, []byte(`require("other")`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Install(dir, sizes); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "physa.lua")
	b, _ := os.ReadFile(path)
	edited := strings.Replace(string(b), `launch_key = "ctrl+a>p"`, `launch_key = "ctrl+a>y"`, 1)
	if err := os.WriteFile(path, []byte(edited), 0o644); err != nil {
		t.Fatal(err)
	}

	wrote, err := Install(dir, sizes)
	if err != nil {
		t.Fatal(err)
	}
	if len(wrote) != 1 {
		t.Errorf("a second install wrote %v; init.lua already requires physa", wrote)
	}
	b, _ = os.ReadFile(path)
	if !strings.Contains(string(b), `launch_key = "ctrl+a>y"`) {
		t.Error("the person's settings were not kept")
	}
	init, _ := os.ReadFile(initLua)
	if got := strings.Count(string(init), `require("physa")`); got != 1 {
		t.Errorf("init.lua requires physa %d times:\n%s", got, init)
	}
	if !strings.HasPrefix(string(init), "require(\"other\")\n") {
		t.Errorf("init.lua's own lines changed:\n%s", init)
	}
}
