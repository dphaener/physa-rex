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

func TestPhysaBinPrefersTheSetting(t *testing.T) {
	t.Setenv(PhysaEnv, "/opt/physa/bin/physa")
	if got, err := physaBin(); err != nil || got != "/opt/physa/bin/physa" {
		t.Fatalf("physaBin() = %q, %v; want the %s setting", got, err, PhysaEnv)
	}
}
