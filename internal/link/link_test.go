package link

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPersonFromSetting(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("PHYSA_PERSON", "")
	if _, err := Person(); err == nil {
		t.Fatal("no setting should be an error, not a guess")
	}
	if err := os.MkdirAll(filepath.Join(dir, "physa"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "physa", "person"), []byte("ada\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if p, err := Person(); err != nil || p != "ada" {
		t.Fatalf("Person() = %q, %v; want ada", p, err)
	}
	t.Setenv("PHYSA_PERSON", "grace")
	if p, _ := Person(); p != "grace" {
		t.Fatalf("PHYSA_PERSON should win, got %q", p)
	}
}
