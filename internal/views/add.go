package views

import (
	"errors"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/dphaener/physa/client"
	"github.com/dphaener/rexuikit/ui"

	"github.com/dphaener/physa-rex/internal/look"
)

// ideaLimit caps an idea's text: a few plain words, not a pitch.
const ideaLimit = 280

// Add is the add-an-idea view: one field for the idea, then what the engine
// made of it. The idea belongs to the person who adds it.
type Add struct {
	src     Source
	profile colorprofile.Profile
	form    *ui.Form
	note    *ui.Dialog
	size    tea.WindowSizeMsg
}

type addedMsg struct {
	idea client.IdeaOut
	err  error
}

// NewAdd returns the add-an-idea view over src. A non-empty blocked says why
// the view cannot add anything (Physa doesn't know the person yet) and is
// shown instead of the form.
func NewAdd(src Source, profile colorprofile.Profile, blocked string) *Add {
	m := &Add{src: src, profile: profile}
	if blocked != "" {
		m.note = m.dialog([]string{blocked})
		return m
	}
	idea := ui.NewTextArea("idea", "The idea").
		Placeholder("What you want to try, in a few plain words").
		CharLimit(ideaLimit).
		Validator(func(s string) error {
			if strings.TrimSpace(s) == "" {
				return errors.New("say what the idea is")
			}
			return nil
		})
	m.form = ui.NewForm(look.Dark(), profile, idea).WithThemeFunc(look.FromBackground)
	return m
}

// NaturalSize is the size a layer should open at.
func (m *Add) NaturalSize() (cols, rows int) {
	if m.form != nil {
		return m.form.NaturalSize()
	}
	return m.note.NaturalSize()
}

func (m *Add) Init() tea.Cmd {
	if m.form != nil {
		return m.form.Init()
	}
	return m.note.Init()
}

func (m *Add) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.size = msg
	case ui.FormCancelledMsg:
		return m, tea.Quit
	case ui.FormSubmittedMsg:
		text := strings.TrimSpace(msg.Values["idea"].(string))
		return m, func() tea.Msg {
			ctx, cancel := bounded()
			defer cancel()
			idea, err := m.src.Add(ctx, text)
			return addedMsg{idea, err}
		}
	case addedMsg:
		if msg.err != nil {
			m.note = m.dialog([]string{"The engine didn't take it:", msg.err.Error()})
		} else {
			m.note = m.dialog([]string{"Added as " + msg.idea.Slug + ". It's yours."})
		}
		t := m.form.Theme()
		m.form = nil
		cmds := []tea.Cmd{m.note.Update(tea.BackgroundColorMsg{Color: background(t)})}
		if m.size.Width > 0 {
			cmds = append(cmds, m.note.Update(m.size))
		}
		return m, tea.Batch(cmds...)
	case tea.KeyPressMsg:
		if m.form == nil {
			return m, tea.Quit
		}
	}
	if m.form != nil {
		_, c := m.form.Update(msg)
		return m, c
	}
	return m, m.note.Update(msg)
}

func (m *Add) View() tea.View {
	if m.form != nil {
		return m.form.View()
	}
	return tea.NewView(m.note.View())
}

func (m *Add) dialog(body []string) *ui.Dialog {
	return ui.NewDialog(look.Dark(), m.profile, "Add an idea", body).WithHint("any key close").WithThemeFunc(look.FromBackground)
}
