package views

import (
	"image/color"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/dphaener/physa/client"
	"github.com/dphaener/rexuikit/ui"
	"github.com/dphaener/rexuikit/ui/theme"

	"github.com/dphaener/physa-rex/internal/look"
)

// Ideas is the ideas view: the repo's ideas, newest first, filterable by
// typing; enter opens an idea's page and esc on the page comes back.
type Ideas struct {
	src     Source
	profile colorprofile.Profile
	theme   theme.Theme
	size    tea.WindowSizeMsg

	ideas []client.IdeaOut
	list  *ui.List
	page  *ui.InfoPanel
	shown pageMsg    // what the page shows, to redraw it in a new theme
	note  *ui.Dialog // loading, empty or an error; any key but esc closes it when done
	done  bool       // the note is final: a key closes the view
}

type ideasMsg struct {
	ideas []client.IdeaOut
	err   error
}

type pageMsg struct {
	idea    client.IdeaOut
	pursuit *client.PursuitOut
	err     error
}

// NewIdeas returns the ideas view over src.
func NewIdeas(src Source, profile colorprofile.Profile) *Ideas {
	m := &Ideas{src: src, profile: profile, theme: look.Dark()}
	m.note = m.dialog("Ideas", []string{"Loading ideas…"}, "esc close")
	return m
}

func (m *Ideas) Init() tea.Cmd {
	return tea.Batch(tea.RequestBackgroundColor, m.load)
}

func (m *Ideas) load() tea.Msg {
	ctx, cancel := bounded()
	defer cancel()
	ideas, err := m.src.Ideas(ctx)
	return ideasMsg{ideas, err}
}

func (m *Ideas) open(slug string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := bounded()
		defer cancel()
		idea, err := m.src.Idea(ctx, slug)
		if err != nil || idea.Pursuit == "" {
			return pageMsg{idea: idea, err: err}
		}
		p, err := m.src.Pursuit(ctx, idea.Pursuit)
		return pageMsg{idea: idea, pursuit: &p, err: err}
	}
}

func (m *Ideas) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.BackgroundColorMsg:
		m.theme = look.FromBackground(msg.Color)
		if m.page != nil {
			m.showPage(m.shown) // InfoPanel takes no theme func, so it is redrawn
		}
	case tea.WindowSizeMsg:
		m.size = msg
	case ideasMsg:
		return m, m.showIdeas(msg)
	case pageMsg:
		return m, m.showPage(msg)
	case ui.ListCancelledMsg:
		return m, tea.Quit
	case ui.ListSelectedMsg:
		return m, m.open(m.ideas[msg.Index].Slug)
	case ui.InfoClosedMsg:
		m.page = nil
		return m, nil
	case tea.KeyPressMsg:
		if m.page == nil && m.list == nil && (m.done || msg.String() == "esc") {
			return m, tea.Quit
		}
	}
	return m, m.forward(msg)
}

// forward hands msg to whatever is on screen, and size and theme to every
// child so a view coming back is already right.
func (m *Ideas) forward(msg tea.Msg) tea.Cmd {
	_, broadcast := msg.(tea.WindowSizeMsg)
	var cmds []tea.Cmd
	if m.page != nil {
		if _, theming := msg.(tea.BackgroundColorMsg); theming {
			return m.forwardBehind(msg)
		}
		_, c := m.page.Update(msg)
		cmds = append(cmds, c)
		if !broadcast {
			return tea.Batch(cmds...)
		}
	}
	return tea.Batch(append(cmds, m.forwardBehind(msg))...)
}

// forwardBehind hands msg to the list, or to the note when there is no list.
func (m *Ideas) forwardBehind(msg tea.Msg) tea.Cmd {
	if m.list != nil {
		_, c := m.list.Update(msg)
		return c
	}
	if m.note != nil {
		return m.note.Update(msg)
	}
	return nil
}

func (m *Ideas) showIdeas(msg ideasMsg) tea.Cmd {
	switch {
	case msg.err != nil:
		m.done = true
		m.note = m.dialog("Ideas", []string{"The engine didn't answer:", msg.err.Error()}, "any key close")
		return m.sized(m.note.Update)
	case len(msg.ideas) == 0:
		m.done = true
		m.note = m.dialog("Ideas", []string{"No ideas in this repo yet.", "Add one from Physa's add-an-idea action."}, "any key close")
		return m.sized(m.note.Update)
	}
	m.ideas = msg.ideas
	rows := make([]ui.Row, len(msg.ideas))
	for i, idea := range msg.ideas {
		rows[i] = ui.Row{Label: clip(idea.Text, labelWidth), Detail: listDetail(idea)}
	}
	m.list = ui.NewList(m.theme, m.profile, rows).WithTitle("Ideas, newest first").WithThemeFunc(look.FromBackground)
	m.note = nil
	return m.sized(func(msg tea.Msg) tea.Cmd { _, c := m.list.Update(msg); return c })
}

func (m *Ideas) showPage(msg pageMsg) tea.Cmd {
	m.shown = msg
	if msg.err != nil {
		m.page = ui.NewInfoPanel(m.theme, m.profile, "", func(int) []string {
			return []string{"The engine didn't answer:", msg.err.Error()}
		})
	} else {
		t := m.theme // the panel is rebuilt, not re-themed, when the theme changes
		m.page = ui.NewInfoPanel(t, m.profile, "", func(w int) []string {
			return pageLines(t, w, msg.idea, msg.pursuit)
		})
	}
	m.page.WithHint("esc back")
	if m.size.Width > 0 {
		_, c := m.page.Update(m.size)
		return c
	}
	return nil
}

// sized gives a new child the window's size and the theme now in force,
// since both arrived before it existed.
func (m *Ideas) sized(update func(tea.Msg) tea.Cmd) tea.Cmd {
	var cmds []tea.Cmd
	if m.size.Width > 0 {
		cmds = append(cmds, update(m.size))
	}
	cmds = append(cmds, update(tea.BackgroundColorMsg{Color: background(m.theme)}))
	return tea.Batch(cmds...)
}

func (m *Ideas) View() tea.View {
	switch {
	case m.page != nil:
		return m.page.View()
	case m.list != nil:
		return m.list.View()
	}
	return tea.NewView(m.note.View())
}

func (m *Ideas) dialog(title string, body []string, hint string) *ui.Dialog {
	return ui.NewDialog(m.theme, m.profile, title, body).WithHint(hint).WithThemeFunc(look.FromBackground)
}

// background is a colour that look.FromBackground maps back to t, for
// re-theming a child that was made after the terminal answered.
func background(t theme.Theme) color.Color {
	if t.Surface == look.Light().Surface {
		return color.White
	}
	return color.Black
}
