// Package link connects a view to the machine's Physa engine through Physa's public client, starting the engine with the physa CLI when it is not running, and names the repo a view shows.
package link

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/dphaener/physa/client"
)

// startTimeout bounds `physa engine start`, which waits for the engine to
// answer before it returns.
const startTimeout = 15 * time.Second

// ErrNoPhysa is returned when no engine answers and the physa CLI is not on
// PATH to start one.
var ErrNoPhysa = errors.New("Physa's engine isn't running, and the physa command isn't on PATH to start it")

// Connect returns a client connected to the engine. When none answers it
// runs `physa engine start` once and tries again.
func Connect(ctx context.Context) (*client.Client, error) {
	socket, err := client.Socket()
	if err != nil {
		return nil, err
	}
	if c, err := client.Dial(ctx, socket); err == nil {
		return c, nil
	}
	bin, err := exec.LookPath("physa")
	if err != nil {
		return nil, ErrNoPhysa
	}
	sctx, cancel := context.WithTimeout(ctx, startTimeout)
	defer cancel()
	if out, err := exec.CommandContext(sctx, bin, "engine", "start").CombinedOutput(); err != nil {
		return nil, fmt.Errorf("physa engine start: %v: %s", err, strings.TrimSpace(string(out)))
	}
	return client.Dial(ctx, socket)
}

// Repo is the repo a view shows: the directory it was opened in. The engine
// finds the repo from any directory inside it.
func Repo() (string, error) {
	if d := os.Getenv("PHYSA_REX_REPO"); d != "" {
		return filepath.Abs(d)
	}
	return os.Getwd()
}
