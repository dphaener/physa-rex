# physa-rex

Physa's views in the Rex terminal: a Physa view plugin (Physa's vision §10). The rules of Physa's map apply here, and Physa's own repo holds them.

## Read first

1. Physa's `.physa/glossary.md`: the only words we use for Physa's concepts, on screens, in code, commits and docs. Never write "ADR".
2. Physa's `views` capability (`physa map slice views` in Physa): what each person sees, and the decisions behind each view.
3. rexuikit's README: its components, surfaces and fast-path rules.

## Rules

- The views read and act only through the engine's API, with Physa's public `client` package. They never import Physa's `internal/` packages or read its store.
- Every view draws with rexuikit's `ui` components in `internal/look`'s theme, not UI of its own. A component that ignores the theme (InfoPanel) is rebuilt when the theme changes.
- Physa's design system is the only source of the look. `internal/look` copies its tokens; when they change there, they change here.
- No internal identifiers on a screen. Slugs and short commits are names people use.
- Paint first: a view's first frame never waits on the engine. Calls run after it, each bounded.
- Plain Go, small packages, each with one job and a doc comment stating its promise, each under 800 non-blank lines of non-test Go.
- The owner decides; agents propose. Ask before anything hard to reverse or outward-facing.

Commit messages follow Conventional Commits (`type(scope): Subject`, then a short prose body), with a `Capability: views` trailer, since that is the Physa capability these views serve.
