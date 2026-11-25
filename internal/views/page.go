package views

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/dphaener/physa/client"
	"github.com/dphaener/rexuikit/ui/theme"
)

// evidenceShown is how many facts an idea's page lists.
const evidenceShown = 8

// pageLines draws an idea's page at width: the idea, then its latest
// pursuit's state, appetite, signal, progress, Needs you and evidence, then
// earlier pursuits. p is nil when the idea has never been pursued.
func pageLines(t theme.Theme, width int, idea client.IdeaOut, p *client.PursuitOut) []string {
	if width < 20 {
		width = 20
	}
	body, muted := t.Body(), t.On(t.Muted, t.Surface)
	var out []string
	add := func(s string) { out = append(out, strings.Split(s, "\n")...) }
	blank := func() { out = append(out, "") }
	field := func(label, value string) {
		const w = 10
		pad := strings.Repeat(" ", w)
		for i, line := range strings.Split(lipgloss.Wrap(value, width-w, " "), "\n") {
			lead := muted.Render(fmt.Sprintf("%-*s", w, label))
			if i > 0 {
				lead = body.Render(pad)
			}
			out = append(out, lead+body.Render(line))
		}
	}

	add(t.On(t.Text, t.Surface).Bold(true).Render(lipgloss.Wrap(idea.Text, width, " ")))
	add(muted.Render(fmt.Sprintf("%s's idea · added %s", idea.Person, day(idea.CreatedAt))))
	blank()

	if p == nil {
		add(muted.Render(lipgloss.Wrap("Not pursued yet. Its strength grows as attention, then evidence, flows into it.", width, " ")))
		return out
	}

	add(chip(t, "PURSUIT") + " " + stateWord(t, p.State) + muted.Render(fmt.Sprintf(" · %s, started %s", p.Person, day(firstOf(p.StartedAt, p.CreatedAt)))))
	if p.StopReason != "" {
		field("Stopped", p.StopReason)
	}
	field("Appetite", appetite(p.Spent, p.Appetite))
	if s := p.Signal; s != nil {
		field("Claim", s.Claim)
		field("For", s.For)
		field("Against", s.Against)
		field("Read", s.Window+", by "+s.Resolver)
	}

	if pr := p.Progress; pr != nil && len(pr.Planned) > 0 {
		blank()
		add(chip(t, "CHANGES") + muted.Render(" "+progressLine(pr)))
		landed := map[string]string{}
		for _, l := range pr.Landed {
			landed[l.Change] = l.Commit
		}
		inWork := map[string]bool{}
		for _, c := range pr.InWork {
			inWork[c] = true
		}
		for _, c := range pr.Planned {
			switch {
			case landed[c] != "":
				add(t.On(t.Done, t.Surface).Render("  ✓ ") + body.Render(c) + muted.Render(" landed "+short(landed[c])))
			case inWork[c]:
				add(t.On(t.Work, t.Surface).Render("  ◐ ") + body.Render(c) + muted.Render(" in work"))
			default:
				add(muted.Render("  · " + c))
			}
		}
	}

	if len(p.NeedsYou) > 0 {
		blank()
		add(chip(t, "NEEDS YOU"))
		for _, n := range p.NeedsYou {
			add(t.On(t.Need, t.Surface).Render("  ! ") + body.Render(lipgloss.Wrap(n.Text, width-4, " ")))
		}
	}

	if len(p.Evidence) > 0 {
		blank()
		add(chip(t, "EVIDENCE"))
		for i, e := range p.Evidence {
			if i == evidenceShown {
				add(muted.Render(fmt.Sprintf("  and %d more", len(p.Evidence)-evidenceShown)))
				break
			}
			add(muted.Render("  "+day(e.At)+"  ") + body.Render(evidenceLine(e)))
		}
	}

	if len(idea.Pursuits) > 1 {
		blank()
		add(chip(t, "EARLIER PURSUITS"))
		for _, ep := range idea.Pursuits[1:] {
			line := fmt.Sprintf("  %s, %s, %s", ep.Slug, ep.State, day(firstOf(ep.StartedAt, ep.CreatedAt)))
			if ep.StopReason != "" {
				line += ": " + ep.StopReason
			}
			add(muted.Render(lipgloss.Wrap(line, width, " ")))
		}
	}
	return out
}

// labelWidth is the most of an idea's text a list row shows, leaving room
// for its facts in the ideas layer.
const labelWidth = 56

// clip is s's first line, cut to n runes with an ellipsis.
func clip(s string, n int) string {
	s, _, _ = strings.Cut(strings.TrimSpace(s), "\n")
	if r := []rune(s); len(r) > n {
		return strings.TrimRight(string(r[:n-1]), " ,.;:") + "…"
	}
	return s
}

// listDetail is the facts beside an idea in the list.
func listDetail(i client.IdeaOut) string {
	if i.Pursuit == "" {
		return i.Person + " · not pursued"
	}
	return i.Person + " · pursuit " + i.PursuitState
}

func chip(t theme.Theme, s string) string {
	return t.On(t.Accent, t.Surface).Bold(true).Render(s)
}

func stateWord(t theme.Theme, state string) string {
	switch state {
	case "active":
		return t.On(t.Work, t.Surface).Render("active")
	case "done":
		return t.On(t.Done, t.Surface).Render("done")
	case "stopped":
		return t.On(t.Muted, t.Surface).Render("stopped")
	}
	return t.Body().Render(state)
}

func progressLine(pr *client.ProgressOut) string {
	s := fmt.Sprintf("%d of %d landed", len(pr.Landed), len(pr.Planned))
	if n := len(pr.InWork); n > 0 {
		s += fmt.Sprintf(" · %d in work", n)
	}
	if pr.AsksOpen > 0 {
		s += fmt.Sprintf(" · %d %s open", pr.AsksOpen, plural(pr.AsksOpen, "ask", "asks"))
	}
	if pr.Replans > 0 {
		s += fmt.Sprintf(" · replanned %d %s", pr.Replans, plural(pr.Replans, "time", "times"))
	}
	return s
}

func evidenceLine(e client.EvidenceOut) string {
	var s string
	switch e.Kind {
	case "landed":
		s = "landed"
	case "reverted":
		s = "reverted by " + e.ByCommit
	case "lead_time":
		s = "lead time"
		if e.Seconds != nil {
			s += " " + (time.Duration(*e.Seconds) * time.Second).Round(time.Minute).String()
		}
	case "ci_passed":
		s = "CI passed"
	case "ci_failed":
		s = "CI failed"
	case "ci_passed_after_failing":
		s = "CI passed after failing"
	default:
		s = e.Kind
	}
	if e.Change != "" {
		s += " · " + e.Change
	}
	if e.PR != 0 {
		s += " · #" + strconv.Itoa(e.PR)
	} else if e.Commit != "" {
		s += " · " + e.Commit
	}
	return s
}

// appetite says the spend against the appetite in tokens, as people say
// them: 120k of 200k.
func appetite(spent, appetite int64) string {
	if appetite <= 0 {
		return tokens(spent) + " tokens spent, no appetite set"
	}
	return fmt.Sprintf("%s of %s tokens", tokens(spent), tokens(appetite))
}

func tokens(n int64) string {
	switch {
	case n >= 1_000_000:
		return strings.TrimSuffix(strconv.FormatFloat(float64(n)/1e6, 'f', 1, 64), ".0") + "m"
	case n >= 1_000:
		return strings.TrimSuffix(strconv.FormatFloat(float64(n)/1e3, 'f', 1, 64), ".0") + "k"
	}
	return strconv.FormatInt(n, 10)
}

// day is an RFC 3339 time as a local date, or the text as it came.
func day(s string) string {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return s
	}
	return t.Local().Format("2 Jan 2006")
}

// short is a commit as people read it: its first eight characters.
func short(commit string) string {
	if len(commit) > 8 {
		return commit[:8]
	}
	return commit
}

func firstOf(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

func plural(n int, one, many string) string {
	if n == 1 {
		return one
	}
	return many
}
