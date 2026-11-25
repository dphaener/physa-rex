// Package views is Physa's views in a Rex layer: the ideas in a repo, an idea's page with its pursuit, and adding an idea, each a bubbletea model drawn with rexuikit's components in Physa's look.
//
// The views read and act only through the engine's API (Physa's vision §10),
// behind Source so they can be tested without an engine. Their words are
// Physa's glossary words, and no screen shows an internal identifier.
package views

import (
	"context"
	"sync"
	"time"

	"github.com/dphaener/physa/client"
)

// callTimeout bounds every call a view makes, so a hung engine shows an
// error instead of a frozen layer.
const callTimeout = 10 * time.Second

// Source is what the views need from the engine.
type Source interface {
	Ideas(ctx context.Context) ([]client.IdeaOut, error)
	Idea(ctx context.Context, slug string) (client.IdeaOut, error)
	Pursuit(ctx context.Context, slug string) (client.PursuitOut, error)
	Add(ctx context.Context, text string) (client.IdeaOut, error)
}

// Engine is a Source over the machine's engine, for one repo. The engine
// names the person who adds an idea, as the CLI does. It dials on the first call, after the view has
// painted, so starting the engine never delays the first frame.
type Engine struct {
	Dial func(context.Context) (*client.Client, error)
	Repo string

	mu sync.Mutex
	c  *client.Client
}

func (e *Engine) call(ctx context.Context, method string, in, out any) error {
	e.mu.Lock()
	if e.c == nil {
		c, err := e.Dial(ctx)
		if err != nil {
			e.mu.Unlock()
			return err
		}
		e.c = c
	}
	c := e.c
	e.mu.Unlock()
	return c.Call(ctx, method, in, out)
}

// Close closes the connection, if there is one.
func (e *Engine) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.c == nil {
		return nil
	}
	return e.c.Close()
}

func (e *Engine) Ideas(ctx context.Context) ([]client.IdeaOut, error) {
	var out []client.IdeaOut
	err := e.call(ctx, "idea.list", client.RepoIn{Repo: e.Repo}, &out)
	return out, err
}

func (e *Engine) Idea(ctx context.Context, slug string) (client.IdeaOut, error) {
	var out client.IdeaOut
	err := e.call(ctx, "idea.show", client.SlugIn{Repo: e.Repo, Slug: slug}, &out)
	return out, err
}

func (e *Engine) Pursuit(ctx context.Context, slug string) (client.PursuitOut, error) {
	var out client.PursuitOut
	err := e.call(ctx, "pursuit.show", client.SlugIn{Repo: e.Repo, Slug: slug}, &out)
	return out, err
}

func (e *Engine) Add(ctx context.Context, text string) (client.IdeaOut, error) {
	var out client.IdeaOut
	err := e.call(ctx, "idea.add", client.AddIn{Repo: e.Repo, Text: text}, &out)
	return out, err
}

func bounded() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), callTimeout)
}
