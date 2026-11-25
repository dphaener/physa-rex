package look

import (
	"image/color"
	"testing"

	"github.com/dphaener/rexuikit/ui/theme"
)

// Every text token must clear WCAG AA (4.5:1) on every ground it is drawn
// on, as rexuikit's own themes do.
func TestTextClearsAA(t *testing.T) {
	for name, th := range map[string]theme.Theme{"dark": Dark(), "light": Light()} {
		grounds := map[string]color.Color{"surface": th.Surface, "well": th.Well, "panel": th.Panel}
		texts := map[string]color.Color{"text": th.Text, "muted": th.Muted, "accent": th.Accent,
			"danger": th.Danger, "done": th.Done, "work": th.Work, "lesson": Lesson(th)}
		for gn, g := range grounds {
			for tn, c := range texts {
				if r := theme.Contrast(c, g); r < 4.5 {
					t.Errorf("%s: %s on %s is %.2f:1, under 4.5:1", name, tn, gn, r)
				}
			}
		}
		if r := theme.Contrast(th.SelectionText, th.Selection); r < 4.5 {
			t.Errorf("%s: selection text is %.2f:1", name, r)
		}
		if r := theme.Contrast(th.OnAccent, th.Accent); r < 4.5 {
			t.Errorf("%s: on-accent is %.2f:1", name, r)
		}
	}
}

func TestFromBackground(t *testing.T) {
	if FromBackground(color.White).Surface != Light().Surface {
		t.Error("a white background should pick Light")
	}
	if FromBackground(color.Black).Surface != Dark().Surface {
		t.Error("a black background should pick Dark")
	}
	if FromBackground(nil).Surface != Dark().Surface {
		t.Error("no background yet should pick Dark")
	}
}
