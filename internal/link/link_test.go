package link

import (
	"path/filepath"
	"testing"
)

func TestRepo(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("PHYSA_REX_REPO", dir)
	if got, err := Repo(); err != nil || got != filepath.Clean(dir) {
		t.Fatalf("Repo() = %q, %v; want %q", got, err, dir)
	}
}
