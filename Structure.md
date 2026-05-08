# Project Structure

Everything you need to know about how this project is wired together.

---

## Directory Tree

```
mywebsite/
├── main.go                   ← You write this. The entire server lives here.
├── go.mod                    ← Module name + Go version. Already created.
├── templates/
│   └── MainPage.html         ← The portfolio page. Uses Go template syntax.
└── static/
    ├── StyleForPages.css     ← All CSS. Zero frameworks, handwritten.
    └── script.js             ← Typewriter, scroll reveal, nav behavior.
```

---

## What Each File Does

### `go.mod`
Declares the module name and minimum Go version. You do not edit this often.
When you add a third-party package (e.g. `go get github.com/go-chi/chi`),
Go automatically adds it here.

### `main.go` — You write this
This is your HTTP server. When you run `go run main.go`, Go compiles and
starts a web server listening on a port (typically `:8080`). See the
"Go Data Structures" section below for everything you need to paste in.

### `templates/MainPage.html`
A standard HTML file extended with Go's template syntax `{{.Field}}`.
Your server reads this file, fills in the data from your Go structs, and
sends the resulting HTML to the browser. The browser never sees the
`{{ }}` markers — they are resolved on the server.

### `static/StyleForPages.css`
All visual styles. Go serves this as a static file — the browser
downloads it directly. You can edit this freely without restarting
the server.

### `static/script.js`
Three small, independent features: the typewriter effect in the hero,
the scroll-reveal animation, and the nav highlight on scroll. Each is
wrapped in an immediately-invoked function so they do not pollute the
global scope (except for `window.ROLES` which the HTML template sets).

---

## The Go Data Structures

Paste this into your `main.go`. Every field maps directly to a
`{{.Field}}` in `MainPage.html`.

```go
package main

import (
    "html/template"
    "log"
    "net/http"
)

// --- Data types ---------------------------------------------------

type Portfolio struct {
    Meta            Meta
    Hero            Hero
    About           About
    FeaturedProject Project
    Projects        []Project   // the cards in the project grid
    Skills          Skills
    Contact         Contact
}

type Meta struct {
    Title       string  // browser tab title
    Description string  // <meta name="description">
    Author      string  // your name, shown in the footer
}

type Hero struct {
    Name    string   // your full name — the big heading
    Roles   []string // cycled by the typewriter: ["Engineer", "Builder", ...]
    Tagline string   // one sentence under the typewriter
}

type About struct {
    Bio   string // the paragraph about you
    Stats []Stat // the four stat cards
}

type Stat struct {
    Value string // e.g. "3+"
    Label string // e.g. "Years of Experience"
}

type Project struct {
    Title       string
    Description string
    Tags        []string // technology names shown as badges
    GitHub      string   // URL — empty string means the link is hidden
    Live        string   // URL — empty string means the link is hidden
}

type Skills struct {
    Languages  []string
    Frameworks []string
    Tools      []string
}

type Contact struct {
    Email    string // used for the mailto: link
    GitHub   string // full URL — empty = hidden
    LinkedIn string // full URL — empty = hidden
}

// --- Server -------------------------------------------------------

func main() {
    // Serve ./static/ at the /static/ URL path.
    // Browser requests /static/StyleForPages.css → reads ./static/StyleForPages.css
    http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

    // Parse the template once at startup.
    // template.Must panics if the template file has a syntax error — good,
    // because you want to catch that immediately, not on the first request.
    tmpl := template.Must(template.ParseFiles("templates/MainPage.html"))

    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        // Reject any path that is not exactly "/"
        // Without this, /anything/random would also serve the homepage.
        if r.URL.Path != "/" {
            http.NotFound(w, r)
            return
        }

        data := Portfolio{
            Meta: Meta{
                Title:       "Rashed Alothman — Software Engineer",
                Description: "Software engineer focused on backend systems and clean code.",
                Author:      "Rashed Alothman",
            },
            Hero: Hero{
                Name:    "Rashed Alothman",
                Roles:   []string{"Software Engineer", "Backend Developer", "Systems Builder"},
                Tagline: "I build reliable systems and tools that stay out of the way.",
            },
            About: About{
                Bio: "I am a software engineer focused on backend systems and developer tooling. " +
                     "I care about writing code that is clear, efficient, and easy to reason about.",
                Stats: []Stat{
                    {Value: "3+",   Label: "Years of Experience"},
                    {Value: "20+",  Label: "Projects Shipped"},
                    {Value: "Go",   Label: "Primary Language"},
                    {Value: "100%", Label: "Remote Ready"},
                },
            },
            FeaturedProject: Project{
                Title: "This Portfolio",
                Description: "A fast, zero-dependency portfolio built entirely with Go's " +
                             "standard library. No frameworks, no build steps.",
                Tags:   []string{"Go", "net/http", "html/template", "CSS"},
                GitHub: "https://github.com/Rashed-alothman",
            },
            Projects: []Project{
                {
                    Title:       "High-Throughput Pipeline",
                    Description: "Concurrent data processing pipeline using Go channels and goroutines.",
                    Tags:        []string{"Go", "Concurrency", "CLI"},
                    GitHub:      "#",
                    Live:        "#",
                },
                {
                    Title:       "REST API Service",
                    Description: "Production REST API with PostgreSQL, JWT auth, and Docker.",
                    Tags:        []string{"Go", "PostgreSQL", "Docker"},
                    GitHub:      "#",
                },
                {
                    Title:       "Dev Workflow CLI",
                    Description: "Automates repetitive dev workflows with a YAML config file.",
                    Tags:        []string{"Go", "YAML", "CLI"},
                    GitHub:      "#",
                    Live:        "#",
                },
            },
            Skills: Skills{
                Languages:  []string{"Go", "Python", "TypeScript", "SQL", "Bash"},
                Frameworks: []string{"net/http", "Chi", "Gin", "GORM", "React"},
                Tools:      []string{"Docker", "PostgreSQL", "Redis", "Git", "Linux"},
            },
            Contact: Contact{
                Email:    "rashed.m.alothman@gmail.com",
                GitHub:   "https://github.com/Rashed-alothman",
                LinkedIn: "#",
            },
        }

        w.Header().Set("Content-Type", "text/html; charset=utf-8")
        if err := tmpl.Execute(w, data); err != nil {
            log.Printf("template error: %v", err)
            http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        }
    })

    log.Println("Portfolio running on http://localhost:8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

---

## How the Template Connection Works

This is the most important thing to understand.

```
Go struct field            Template variable
──────────────────         ─────────────────
data.Hero.Name        →    {{.Hero.Name}}
data.About.Stats      →    {{range .About.Stats}} ... {{end}}
data.Contact.GitHub   →    {{if .Contact.GitHub}} ... {{end}}
```

When you call `tmpl.Execute(w, data)`, Go's template engine walks
through `MainPage.html`, finds every `{{...}}` expression, replaces it
with the value from your struct, and writes the final HTML to `w`
(the response writer). The browser receives plain HTML — it never sees
the Go template syntax.

**Context-aware escaping** — `html/template` (not `text/template`)
automatically escapes values based on where they appear:
- Inside HTML: `<` becomes `&lt;`  → prevents HTML injection
- Inside a `href=`: URL-encodes suspicious characters
- Inside a `<script>` block: JavaScript-escapes strings → prevents XSS

You cannot accidentally create a security hole by putting user data
into the template. The package handles it.

---

## The Request Lifecycle

```
Browser                   Go Server
───────                   ─────────
GET / HTTP/1.1       →    http.HandleFunc("/", handler)
                               ↓
                          Build Portfolio{} struct with your data
                               ↓
                          tmpl.Execute(w, data)
                               ↓
                          Template engine fills in {{.Fields}}
                               ↓
200 OK + HTML        ←    Write complete HTML to response

GET /static/style.css →   http.FileServer reads ./static/StyleForPages.css
200 OK + CSS          ←    Sends file bytes directly
```

---

## How to Run

```bash
# From the project root (mywebsite/)
go run main.go

# Then open in your browser:
# http://localhost:8080
```

To stop the server: `Ctrl+C`

---

## How to Add a New Page

1. Create `templates/AboutPage.html`
2. In `main.go`, add a new route:
   ```go
   http.HandleFunc("/about", func(w http.ResponseWriter, r *http.Request) {
       t := template.Must(template.ParseFiles("templates/AboutPage.html"))
       t.Execute(w, yourData)
   })
   ```

## How to Add a Contact Form (POST handler)

```go
http.HandleFunc("/contact", func(w http.ResponseWriter, r *http.Request) {
    if r.Method == http.MethodPost {
        name    := r.FormValue("name")
        email   := r.FormValue("email")
        message := r.FormValue("message")
        // do something: log it, email it, save to DB
        http.Redirect(w, r, "/#contact", http.StatusSeeOther)
        return
    }
    http.NotFound(w, r)
})
```
