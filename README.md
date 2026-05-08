# Rashed Alothman — Portfolio

A personal portfolio website built with Go's standard library. No frameworks,
no build pipeline, no JavaScript dependencies. Just `net/http`, `html/template`,
hand-crafted CSS, and a small amount of vanilla JavaScript.

---

## What This Is

A server-rendered portfolio site that demonstrates how Go can power a clean,
fast web page without reaching for external tools. The Go server reads a single
HTML template, fills in the data from typed structs, and serves the result.
Static files (CSS, JS) are served directly from the `./static` directory.

The goal was to write the minimum amount of code needed to get a great-looking,
fully functional portfolio online — and to understand every line of it.

---

## Tech Stack

| Layer       | Technology                        |
|-------------|-----------------------------------|
| Language    | Go                                |
| HTTP server | `net/http` (standard library)     |
| Templating  | `html/template` (standard library)|
| Styling     | Vanilla CSS (custom properties)   |
| JavaScript  | Vanilla JS (no framework)         |
| Fonts       | Inter, JetBrains Mono (Google Fonts) |

Zero third-party Go dependencies.

---

## Project Structure

```
mywebsite/
├── main.go                   Server, routes, and all portfolio data
├── go.mod                    Module definition
├── templates/
│   └── MainPage.html         HTML template with Go {{.Field}} syntax
└── static/
    ├── StyleForPages.css     All styles
    └── script.js             Typewriter effect, scroll reveal, nav behavior
```

For a detailed explanation of how every file connects, see `Structure.md`.

---

## Getting Started

**Prerequisites:** Go 1.21 or later. Verify with `go version`.

**Clone and run:**

```bash
git clone https://github.com/Rashed-alothman/Website.git
cd Website
go run main.go
```

Then open `http://localhost:8080` in your browser.

To build a standalone binary instead of running from source:

```bash
go build -o portfolio .
./portfolio
```

---

## Customizing the Content

All portfolio content — your name, bio, projects, skills, and contact links —
lives in the `data` variable inside `main.go`. You do not need to touch the
HTML template to change text.

Find the `Portfolio{}` struct literal in `main.go` and update the fields:

```go
data := Portfolio{
    Hero: Hero{
        Name:    "Your Name",
        Roles:   []string{"Your Role", "Another Role"},
        Tagline: "One sentence about what you do.",
    },
    // ... rest of the fields
}
```

The template automatically reflects every change on the next request.

---

## How the Template System Works

Go's `html/template` package reads `MainPage.html` and replaces every
`{{.FieldName}}` marker with the corresponding value from the Go struct.
This happens on the server — the browser receives plain HTML with no
template syntax in it.

Example: `{{.Hero.Name}}` in the template becomes `Rashed Alothman` in
the rendered page because the Go code passes `Hero.Name = "Rashed Alothman"`.

The package also escapes values based on context (HTML body, attributes,
script tags) so injecting user data cannot cause cross-site scripting.

See `Structure.md` for the complete type definitions and a request lifecycle
diagram.

---

## Learning Resources

`Tips.md` in this repository covers:

- Go language fundamentals (error handling, goroutines, interfaces)
- The `net/http` and `html/template` packages in depth
- HTTP concepts, REST API design, status codes
- Database access with `database/sql` and GORM
- Authentication patterns (sessions, JWT, password hashing)
- Docker and deployment
- Security fundamentals (SQL injection, XSS, CSRF, rate limiting)
- A recommended learning path from beginner to production-ready

---

## Deployment

The application is a single binary that serves HTTP on the configured port.
It can be deployed to any Linux server or container platform.

**Environment variables:**

| Variable | Default | Description                     |
|----------|---------|---------------------------------|
| `PORT`   | `8080`  | Port the server listens on      |

**Docker:**

```bash
docker build -t portfolio .
docker run -p 8080:8080 portfolio
```

A sample `Dockerfile` using a multi-stage build (Go builder + Alpine runtime)
is described in `Tips.md` under "Docker and Deployment".

For HTTPS in production, place the server behind a reverse proxy (Caddy or
Nginx) that handles TLS termination, or deploy to a platform that provides
it automatically (Fly.io, Railway, Render).

---

## License

MIT. Use it however you want.
