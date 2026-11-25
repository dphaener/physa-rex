// Package look is Physa's look in the terminal: Physa's colour tokens, light and dark, mapped onto rexuikit's theme, so every view drawn with rexuikit's components reads as Physa.
//
// The values are copied from Physa's design system (internal/design/tokens.css
// in Physa), which stays the only source of the look. A change there is
// copied here by hand.
package look

import (
	"image/color"

	"github.com/dphaener/rexuikit/ui/theme"
)

// Dark is Physa's dark look, and the first-frame default.
func Dark() theme.Theme {
	return theme.Theme{
		Surface:       hex(0x0a1322), // surface
		Well:          hex(0x060d19), // surface-sunk
		Panel:         hex(0x101c31), // surface-raised
		Text:          hex(0xe9f0ff), // ink
		Muted:         hex(0x9fb2d4), // ink-muted
		Accent:        hex(0x6d9bff), // accent
		OnAccent:      hex(0x06112a), // on-accent
		Danger:        hex(0xff8a5c), // danger
		Border:        hex(0x6680a6), // line-strong
		Faint:         hex(0x22324d), // line
		Strip:         hex(0x16295a), // accent-soft
		StripText:     hex(0xe9f0ff), // ink
		Selection:     hex(0x16295a), // accent-soft
		SelectionText: hex(0xe9f0ff), // ink
		Need:          hex(0xff8a5c), // danger: what needs a person travels with a word
		Work:          hex(0x8fe3ff), // tube-idea: growing
		Done:          hex(0x3fd0d6), // steady
		Meta:          hex(0x9fb2d4), // ink-muted
	}
}

// Light is Physa's light look.
func Light() theme.Theme {
	return theme.Theme{
		Surface:       hex(0xf4f8ff),
		Well:          hex(0xe7eefb),
		Panel:         hex(0xffffff),
		Text:          hex(0x0b1a33),
		Muted:         hex(0x47597a),
		Accent:        hex(0x1757f5),
		OnAccent:      hex(0xffffff),
		Danger:        hex(0xad380b),
		Border:        hex(0x7487aa),
		Faint:         hex(0xd2ddf0),
		Strip:         hex(0xdce7ff),
		StripText:     hex(0x0b1a33),
		Selection:     hex(0xdce7ff),
		SelectionText: hex(0x0b1a33),
		Need:          hex(0xad380b),
		Work:          hex(0x0f45d1), // accent-hover: tube-idea is too light for text here
		Done:          hex(0x086a75),
		Meta:          hex(0x47597a),
	}
}

// Lesson is the only warm colour in Physa: the knot a withered idea leaves.
// It is not a theme token, so views ask for it by theme.
func Lesson(t theme.Theme) color.Color {
	if isLight(t) {
		return hex(0x8a5200) // lesson-ink
	}
	return hex(0xf6c46d)
}

// FromBackground picks Light for a light terminal background and Dark
// otherwise, the way rexuikit's theme.FromBackground does.
func FromBackground(bg color.Color) theme.Theme {
	if theme.FromBackground(bg).Surface == theme.Light().Surface {
		return Light()
	}
	return Dark()
}

func isLight(t theme.Theme) bool { return t.Surface == Light().Surface }

func hex(v uint32) color.Color {
	return color.RGBA{R: uint8(v >> 16), G: uint8(v >> 8), B: uint8(v), A: 0xff}
}
