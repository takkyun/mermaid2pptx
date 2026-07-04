# Contributing to mermaid2pptx

*日本語版: [CONTRIBUTING.ja.md](CONTRIBUTING.ja.md).*

Thanks for your interest in improving mermaid2pptx. Issues and pull requests are
welcome.

## Ground rules

- **Standard library only.** This tool has no production dependencies and aims
  to keep it that way. Do not add a module `require` without discussing it in an
  issue first. (`python-pptx` and `mermaid-cli` are external dev/verify tools,
  not Go dependencies.)
- **Code comments in English.** Keep them at the density of the surrounding code
  and explain *why*, not *what*.
- **Commit messages in English.**
- See [AGENTS.md](AGENTS.md) for the architecture and deeper conventions.

## Development setup

Requires Go (see the version in [go.mod](go.mod)). `mermaid-cli` (`mmdc`) is
only needed if you convert `.mmd` input or regenerate the sample SVGs.

```sh
go build -o mermaid2pptx ./cmd/mermaid2pptx
```

## Before you open a pull request

Run all of the following and make sure they are clean:

```sh
gofmt -l .                    # must print nothing
go vet ./...
go test ./...
GOOS=windows GOARCH=amd64 go build -o /dev/null ./cmd/mermaid2pptx  # cross-compile check
```

The same checks run in CI on every push and pull request.

### If your change affects the generated .pptx

A passing well-formed-XML test does not prove the slide looks right. When you
change the generator:

1. Regenerate the samples and keep them in sync:
   ```sh
   ./mermaid2pptx -f sample/graph*.svg
   ```
2. Inspect the output (e.g. with `python-pptx`) and open at least one `.pptx` in
   PowerPoint to confirm shapes and connectors are placed and editable.
3. Keep `TestConnectorEndpointsMatchSVG` green — it is the guard for connector
   geometry. `fitTransform` in `slide.go` is the single source of truth for the
   px→EMU transform shared by the generator and the test.

Adding a new diagram type? See the "Adding a new diagram type" section in
[AGENTS.md](AGENTS.md).

## Pull request guidelines

- Keep each PR focused on one change; keep commits small and self-contained.
- Describe what changed and why, and note whether generated output changed.
- Do not commit the built binary or anything under `_sandbox/` (both are
  git-ignored).
- Do not include customer names, account numbers, or other PII in code, samples,
  or generated artifacts.

## License of contributions

By contributing, you agree that your contributions are licensed under the
[Apache License, Version 2.0](LICENSE), the same license as the project.
