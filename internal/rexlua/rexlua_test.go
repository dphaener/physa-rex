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
	if _, err := Install(dir, sizes, Bins{}); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "physa.lua")
	b, _ := os.ReadFile(path)
	edited := strings.Replace(string(b), `launch_key = "ctrl+a>y"`, `launch_key = "ctrl+a>z"`, 1)
	if err := os.WriteFile(path, []byte(edited), 0o644); err != nil {
		t.Fatal(err)
	}

	wrote, err := Install(dir, sizes, Bins{})
	if err != nil {
		t.Fatal(err)
	}
	if len(wrote) != 1 {
		t.Errorf("a second install wrote %v; init.lua already requires physa", wrote)
	}
	b, _ = os.ReadFile(path)
	if !strings.Contains(string(b), `launch_key = "ctrl+a>z"`) {
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

func TestInstallFillsOnlyEmptyBins(t *testing.T) {
	dir := t.TempDir()
	if _, err := Install(dir, sizes, Bins{Rex: "/opt/go/bin/physa-rex", Physa: `/Users/a "b"/bin/physa`}); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "physa.lua")
	b, _ := os.ReadFile(path)
	for _, want := range []string{`bin = "/opt/go/bin/physa-rex",`, `physa = "/Users/a \"b\"/bin/physa",`} {
		if !strings.Contains(string(b), want) {
			t.Errorf("physa.lua lacks %s:\n%s", want, b)
		}
	}
	if _, err := Install(dir, sizes, Bins{Rex: "/elsewhere/physa-rex"}); err != nil {
		t.Fatal(err)
	}
	b, _ = os.ReadFile(path)
	if !strings.Contains(string(b), `bin = "/opt/go/bin/physa-rex",`) {
		t.Errorf("a set bin was overwritten:\n%s", b)
	}
}

func TestInstallAddsPhysaToAnOlderSettingsBlock(t *testing.T) {
	dir := t.TempDir()
	old := "-- settings begin\nlocal settings = {\n  launch_key = \"ctrl+a>p\",\n  bin = \"\",\n}\n-- settings end\n"
	if err := os.WriteFile(filepath.Join(dir, "physa.lua"), []byte(old), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Install(dir, sizes, Bins{Physa: "/u/bin/physa"}); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(filepath.Join(dir, "physa.lua"))
	for _, want := range []string{`launch_key = "ctrl+a>p",`, `bin = "",`, "  physa = \"/u/bin/physa\",\n}\n-- settings end"} {
		if !strings.Contains(string(b), want) {
			t.Errorf("physa.lua lacks %q:\n%s", want, b)
		}
	}
}

func TestLuaStringRefusesWhatItCannotHold(t *testing.T) {
	if _, err := luaString("/tmp/caf\u00e9/physa"); err == nil {
		t.Error("a non-ASCII path was quoted; it should be refused")
	}
}

func TestConfigDirIsWhereRexLooks(t *testing.T) {
	home := func() (string, error) { return "/Users/d", nil }
	appData := func() (string, error) { return "/Users/d/Library/Application Support", nil }
	env := func(m map[string]string) func(string) string { return func(k string) string { return m[k] } }
	for _, c := range []struct {
		goos string
		env  map[string]string
		want string
	}{
		{"darwin", nil, "/Users/d/.config/rex"},
		{"linux", nil, "/Users/d/.config/rex"},
		{"darwin", map[string]string{"XDG_CONFIG_HOME": "/x"}, "/x/rex"},
		{"darwin", map[string]string{"REX_CONFIG": "/r/init.lua", "XDG_CONFIG_HOME": "/x"}, "/r"},
		{"windows", nil, "/Users/d/Library/Application Support/rex"},
	} {
		got, err := configDir(c.goos, env(c.env), home, appData)
		if err != nil || got != c.want {
			t.Errorf("configDir(%s, %v) = %q, %v; want %q", c.goos, c.env, got, err, c.want)
		}
	}
}
