package views

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/dphaener/physa/client"
	"github.com/dphaener/rexuikit/ui"

	"github.com/dphaener/physa-rex/internal/look"
)

type fake struct {
	ideas    []client.IdeaOut
	pursuits map[string]client.PursuitOut
	err      error
	added    string
}

func (f *fake) Ideas(context.Context) ([]client.IdeaOut, error) { return f.ideas, f.err }
func (f *fake) Idea(_ context.Context, slug string) (client.IdeaOut, error) {
	for _, i := range f.ideas {
		if i.Slug == slug {
			return i, nil
		}
	}
	return client.IdeaOut{}, errors.New("no such idea")
}
func (f *fake) Pursuit(_ context.Context, slug string) (client.PursuitOut, error) {
	return f.pursuits[slug], nil
}
func (f *fake) Add(_ context.Context, text string) (client.IdeaOut, error) {
	f.added = text
	return client.IdeaOut{Slug: "try-it", Text: text}, f.err
}

var ansi = regexp.MustCompile(`\x1b\[[0-9;:]*[A-Za-z]`)

func plain(v tea.View) string { return ansi.ReplaceAllString(v.Content, "") }

// drive runs cmd and feeds every message it yields back into m, as the
// program would, skipping terminal queries the test has no terminal for.
func drive(t *testing.T, m tea.Model, cmd tea.Cmd) {
	t.Helper()
	queue := []tea.Cmd{cmd}
	for n := 0; len(queue) > 0; n++ {
		if n > 200 {
			t.Fatal("the view never settled")
		}
		c := queue[0]
		queue = queue[1:]
		if c == nil {
			continue
		}
		switch msg := c().(type) {
		case nil:
		case tea.BatchMsg:
			queue = append(queue, msg...)
		case tea.QuitMsg:
			return
		default:
			if query(msg) {
				continue
			}
			_, next := m.Update(msg)
			queue = append(queue, next)
		}
	}
}

// query reports a message that asks the terminal something; a test has no
// terminal to answer.
func query(msg tea.Msg) bool {
	name := strings.ToLower(fmt.Sprintf("%T", msg))
	return strings.HasPrefix(name, "tea.") && strings.Contains(name, "request")
}

func key(s string) tea.KeyPressMsg {
	switch s {
	case "enter":
		return tea.KeyPressMsg{Code: tea.KeyEnter}
	case "esc":
		return tea.KeyPressMsg{Code: tea.KeyEscape}
	}
	return tea.KeyPressMsg{Code: rune(s[0]), Text: s}
}

func ideas() *fake {
	return &fake{
		ideas: []client.IdeaOut{
			{Slug: "bulk-invite", Text: "Invite a whole team at once", Person: "ada", CreatedAt: "2026-11-20T10:00:00Z",
				Pursuit: "bulk-invite", PursuitState: "active"},
			{Slug: "dark-mode", Text: "A dark look for the phone web", Person: "grace", CreatedAt: "2026-11-19T10:00:00Z"},
		},
		pursuits: map[string]client.PursuitOut{"bulk-invite": {
			Slug: "bulk-invite", Person: "ada", State: "active", Spent: 120_000, Appetite: 200_000,
			StartedAt: "2026-11-21T10:00:00Z",
			Signal: &client.SignalOut{Claim: "Teams invite more people", For: "invites per team rise",
				Against: "they don't", Window: "2026-12-15", Resolver: "ada"},
			Progress: &client.ProgressOut{Planned: []string{"form", "send"}, Landed: []client.LandedOut{{Change: "form", Commit: "abc12345"}},
				InWork: []string{"send"}},
		}},
	}
}

func TestIdeasListThenPageThenBack(t *testing.T) {
	m := NewIdeas(ideas(), colorprofile.TrueColor)
	drive(t, m, m.Init())
	drive(t, m, func() tea.Msg { return tea.WindowSizeMsg{Width: 90, Height: 24} })
	got := plain(m.View())
	for _, want := range []string{"Invite a whole team at once", "ada · pursuit active", "grace · not pursued"} {
		if !strings.Contains(got, want) {
			t.Errorf("the list lacks %q:\n%s", want, got)
		}
	}

	drive(t, m, func() tea.Msg { return key("enter") })
	got = plain(m.View())
	for _, want := range []string{"ada's idea", "PURSUIT active", "120k of 200k tokens", "Teams invite more people", "1 of 2 landed · 1 in work", "✓ form", "◐ send"} {
		if !strings.Contains(got, want) {
			t.Errorf("the page lacks %q:\n%s", want, got)
		}
	}

	drive(t, m, func() tea.Msg { return key("esc") })
	if got = plain(m.View()); !strings.Contains(got, "Ideas, newest first") {
		t.Errorf("esc on the page should come back to the list:\n%s", got)
	}
}

func TestIdeasEmptyAndError(t *testing.T) {
	for name, f := range map[string]*fake{
		"empty": {},
		"error": {err: errors.New("connection refused")},
	} {
		m := NewIdeas(f, colorprofile.TrueColor)
		drive(t, m, m.Init())
		got := plain(m.View())
		want := map[string]string{"empty": "No ideas in this repo yet.", "error": "connection refused"}[name]
		if !strings.Contains(got, want) {
			t.Errorf("%s: want %q in:\n%s", name, want, got)
		}
	}
}

func TestPageNeverPursued(t *testing.T) {
	lines := strings.Join(pageLines(look.Dark(), 60, ideas().ideas[1], nil), "\n")
	if !strings.Contains(ansi.ReplaceAllString(lines, ""), "Not pursued yet.") {
		t.Errorf("an idea never pursued should say so:\n%s", lines)
	}
}

func TestAddSendsTheIdea(t *testing.T) {
	f := &fake{}
	m := NewAdd(f, colorprofile.TrueColor)
	drive(t, m, func() tea.Msg { return ui.FormSubmittedMsg{Values: ui.Values{"idea": "  Try a shorter signup  "}} })
	if f.added != "Try a shorter signup" {
		t.Fatalf("the engine got %q", f.added)
	}
	if got := plain(m.View()); !strings.Contains(got, "Added as try-it. It's yours.") {
		t.Errorf("after adding:\n%s", got)
	}
}

func TestTokens(t *testing.T) {
	for n, want := range map[int64]string{0: "0", 950: "950", 1000: "1k", 120_500: "120.5k", 1_500_000: "1.5m"} {
		if got := tokens(n); got != want {
			t.Errorf("tokens(%d) = %q, want %q", n, got, want)
		}
	}
}
