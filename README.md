# physa-rex

[Physa](https://github.com/dphaener/physa)'s views in the [Rex](https://www.superlogical.com/rex) terminal. Each view is a small layer that opens over your work, does one job and closes, drawn with [rexuikit](https://github.com/dphaener/rexuikit) in Physa's look.

## Views

- **Ideas**: the repo's ideas, newest first. Type to filter, then press enter for an idea's page: who it belongs to, its pursuit's state, its appetite, its signal, its planned changes and what landed, what needs you, and the evidence behind it. Esc goes back.
- **Add an idea**: one field. The idea belongs to you.

Needs you and watching a live session come when Physa's engine runs them. An idea's page shows its strength once Physa works strength out; until then it shows the facts behind it.

## Install

You need Physa's `physa` CLI on your PATH. The views start its engine when it isn't running.

```sh
go install github.com/dphaener/physa-rex/cmd/physa-rex@latest
physa-rex rex install
```

Rex starts the views through a login shell, which doesn't read an interactive shell's setup such as `.zshrc`, where mise and similar tools often add to PATH. `rex install` checks what a login shell finds, and when it can't find `physa-rex` or `physa`, it writes their absolute paths into `physa.lua`'s settings and says so. After moving either binary, run `rex install` again, or edit `bin` and `physa` in the settings block.

`rex install` writes `physa.lua` beside Rex's `init.lua` and requires it from there once. Press `ctrl+a y` to enter Physa's key mode, then `i` for the ideas or `a` to add one. Both are also in Rex's command palette. Change the key in the settings block at the top of `physa.lua`; a reinstall keeps it.

The views show the repo of the pane you open them from. The engine names you when you add an idea, the same way the `physa` CLI does.

## How it works

The views call the engine's local API only through Physa's public `client` package, as every Physa plugin does. They never read Physa's store. Physa's colours, light and dark, are mapped onto rexuikit's theme in `internal/look`, and every text colour clears WCAG AA on every ground.

## Developing

```sh
go vet ./... && go test ./...
```

## License

MIT. Physa's core is under Apache-2.0.
