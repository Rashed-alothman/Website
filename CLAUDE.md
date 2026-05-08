# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
go run main.go                          # run locally on :8080
go build -o portfolio .                 # build binary
go test ./...                           # run all tests
go test -run TestHome_ReturnsOK ./...   # run a single test
PPROF_ADDR=localhost:6060 go run main.go  # run with pprof debug server

docker build -t portfolio .
docker run -p 8080:8080 portfolio

fly deploy                              # deploy to Fly.io (fly launch on first run)
```

Tests must be run from the project root because `newServer()` resolves `templates/` relative to the working directory.

## Architecture

Everything lives in `main.go`. The `server` struct owns the two parsed templates and the immutable `Portfolio` data. There are no package-level variables. `newServer()` returns `(*server, error)` and is the only place templates are parsed and data is assembled.

**Two-mux isolation for pprof:** The public server uses `http.NewServeMux()`. The blank import `_ "net/http/pprof"` registers profiling handlers on `http.DefaultServeMux`. The debug server (started only when `PPROF_ADDR` env var is set) uses `nil` as its handler, which routes to `DefaultServeMux`. This keeps profiling endpoints unreachable from the public server.

**Buffered rendering:** `render()` executes templates into a `bytes.Buffer` before writing to `http.ResponseWriter`. This prevents partial HTML from being flushed to the browser if a template errors mid-execution.

**Content vs. presentation:** All portfolio content (name, bio, projects, skills) lives in `buildPortfolioData()` in `main.go`. The templates contain no hardcoded content — only `{{.Field}}` references.

## Template Rules (Critical)

`html/template` **executes `{{ }}` actions inside HTML `<!-- -->` comments** — they are not skipped. A `{{range}}` or `{{if}}` inside an HTML comment opens a block that must be closed, and failing to do so causes an "unexpected EOF" panic at startup. Use `{{/* */}}` Go template comments for documentation blocks that should be ignored by the parser.

**JS context:** `html/template`'s context-aware parser loses track of named range variables (e.g. `$r`) inside `<script>` blocks. Pass data into scripts via the `json` template function: `{{.Hero.Roles | json}}`. The `jsJSON` helper encodes to JSON and returns `template.JS`, which bypasses JS context escaping.

**Template naming:** `template.New("MainPage.html").ParseFiles("templates/MainPage.html")` — the name passed to `New` must match the file's basename exactly, or `Execute` runs the empty root template instead.
