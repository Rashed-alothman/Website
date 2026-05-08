// Package main runs the portfolio web server. It serves the portfolio
// landing page at "/" and a contact form at "/contact", rendering all
// dynamic content through Go's html/template package.
//
// # Environment variables
//
//	PORT        Listen address for the public server. Defaults to 8080.
//	            Cloud platforms (Railway, Fly.io, Render) inject this automatically.
//	PPROF_ADDR  When set (e.g. "localhost:6060"), starts a second HTTP server
//	            on that address exposing /debug/pprof/ profiling endpoints.
//	            Always bind to localhost — never expose pprof to the public internet.
//
// # Running locally
//
//	go run main.go
//	# With profiling enabled:
//	PPROF_ADDR=localhost:6060 go run main.go
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	_ "net/http/pprof" // side-effect: registers /debug/pprof/* on http.DefaultServeMux
	"os"
)

// ── Data types ────────────────────────────────────────────────────────────────

// Portfolio is the complete data set rendered into templates/MainPage.html.
// Every exported field maps to a {{.Field}} expression in the template.
// Pass a populated Portfolio to tmpl.Execute to render the landing page.
type Portfolio struct {
	Meta            Meta
	Hero            Hero
	About           About
	FeaturedProject Project
	Projects        []Project
	Skills          Skills
	Contact         Contact
}

// Meta holds page-level metadata written into <head> and the footer.
type Meta struct {
	Title       string // browser tab title
	Description string // <meta name="description">
	Author      string // shown in the footer
}

// Hero holds the data for the introduction section at the top of the page.
// Roles is cycled character-by-character by the typewriter in script.js.
type Hero struct {
	Name    string   // the large heading
	Roles   []string // e.g. ["Software Engineer", "Backend Developer"]
	Tagline string   // one sentence rendered below the typewriter
}

// About contains the biography paragraph and the stat cards beside it.
type About struct {
	Bio   string
	Stats []Stat
}

// Stat is one value/label card in the About section.
type Stat struct {
	Value string // short value, e.g. "3+"
	Label string // descriptive label, e.g. "Years of Experience"
}

// Project describes a single project card in the projects grid,
// or the featured project shown above the grid.
// An empty GitHub or Live string hides the corresponding link in the template.
type Project struct {
	Title       string
	Description string
	Tags        []string // technology badge labels
	GitHub      string   // repository URL — empty string hides the link
	Live        string   // live demo URL — empty string hides the link
}

// Skills groups the three technology tag lists shown in the Skills section.
type Skills struct {
	Languages  []string
	Frameworks []string
	Tools      []string
}

// Contact holds the owner's email address and social profile URLs.
// An empty GitHub or LinkedIn string hides that button in the template.
type Contact struct {
	Email    string // used for the mailto: link
	GitHub   string // full profile URL — empty = hidden
	LinkedIn string // full profile URL — empty = hidden
}

// ContactPage is the data passed to templates/ContactPage.html for /contact.
// It is a separate type from Portfolio because /contact is its own route
// and only needs a subset of the site-wide data.
// Flash="sent" renders the thank-you state; any other value renders the form.
type ContactPage struct {
	Meta    Meta
	Contact Contact
	Flash   string // "sent" → thank-you message; "" → form
}

// ── Template helper ───────────────────────────────────────────────────────────

// jsJSON encodes v as JSON and wraps it in template.JS, marking it safe
// for direct insertion into a <script> block without further escaping.
//
// html/template is context-aware and loses track of range-variable names
// (e.g. $r) when parsing inside a <script> JS context. Returning a typed
// template.JS bypasses that escaping entirely and produces correct output:
//
//	{{.Hero.Roles | json}}  →  ["Software Engineer","Backend Developer"]
func jsJSON(v any) template.JS {
	b, _ := json.Marshal(v) // encoding a []string or primitive can never fail
	return template.JS(b)
}

// ── Server ────────────────────────────────────────────────────────────────────

// server owns the parsed templates and the immutable portfolio data.
// Keeping state here instead of in package-level variables makes the
// dependency graph explicit and the handlers easy to test in isolation.
type server struct {
	mainTmpl    *template.Template
	contactTmpl *template.Template
	data        Portfolio
}

// newServer parses the HTML templates and assembles the portfolio data.
// It returns an error rather than panicking, so main can log.Fatal cleanly.
func newServer() (*server, error) {
	funcs := template.FuncMap{"json": jsJSON}

	// template.New("MainPage.html") — the name must match the file's basename.
	// If it doesn't, Execute runs the empty root template instead of the file.
	mainTmpl, err := template.New("MainPage.html").
		Funcs(funcs).
		ParseFiles("templates/MainPage.html")
	if err != nil {
		return nil, fmt.Errorf("parse MainPage.html: %w", err)
	}

	contactTmpl, err := template.ParseFiles("templates/ContactPage.html")
	if err != nil {
		return nil, fmt.Errorf("parse ContactPage.html: %w", err)
	}

	return &server{
		mainTmpl:    mainTmpl,
		contactTmpl: contactTmpl,
		data:        buildPortfolioData(),
	}, nil
}

// routes registers all URL handlers on a fresh ServeMux and returns it.
// Using an explicit mux (not http.DefaultServeMux) keeps the public server
// isolated from pprof, which is registered on http.DefaultServeMux.
func (s *server) routes() http.Handler {
	mux := http.NewServeMux()

	// /static/ maps to ./static/ on disk.
	// StripPrefix removes the URL prefix before FileServer looks up the file,
	// so GET /static/StyleForPages.css reads ./static/StyleForPages.css.
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	mux.HandleFunc("/", s.home)
	mux.HandleFunc("/contact", s.contact)
	return mux
}

// render executes tmpl with data, writing the result to w only if the
// template executes without error. This prevents partial HTML from being
// sent to the browser when an error occurs mid-template.
func render(w http.ResponseWriter, tmpl *template.Template, data any) error {
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return err
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, err := buf.WriteTo(w)
	return err
}

// home serves GET /. Any path other than exactly "/" returns 404.
func (s *server) home(w http.ResponseWriter, r *http.Request) {
	// The "/" pattern in ServeMux matches everything not handled elsewhere.
	// Without this guard, GET /anything would serve the homepage.
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if err := render(w, s.mainTmpl, s.data); err != nil {
		log.Printf("home: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// contact handles GET and POST /contact.
//
// GET  — renders the form (or the thank-you message if ?sent=1 is present).
// POST — reads the form fields, logs them, and redirects with the
// Post/Redirect/Get pattern to prevent duplicate submissions on refresh.
func (s *server) contact(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		name := r.FormValue("name")
		email := r.FormValue("email")
		message := r.FormValue("message")
		log.Printf("contact: from=%s <%s>: %s", name, email, message)
		http.Redirect(w, r, "/contact?sent=1", http.StatusSeeOther)
		return
	}

	// ?sent=1 is set by the redirect above. The template shows the
	// thank-you state when Flash equals "sent".
	flash := ""
	if r.URL.Query().Get("sent") == "1" {
		flash = "sent"
	}

	page := ContactPage{
		Meta: Meta{
			Title:       "Contact — " + s.data.Meta.Author,
			Description: "Get in touch with " + s.data.Meta.Author + ".",
			Author:      s.data.Meta.Author,
		},
		Contact: s.data.Contact,
		Flash:   flash,
	}

	if err := render(w, s.contactTmpl, page); err != nil {
		log.Printf("contact: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// ── Portfolio data ────────────────────────────────────────────────────────────

// buildPortfolioData returns the content shown on the site.
// Edit the literal values here to update the site without touching templates.
func buildPortfolioData() Portfolio {
	return Portfolio{
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
				{Value: "3+", Label: "Years of Experience"},
				{Value: "5+", Label: "Projects Shipped"},
				{Value: "Go", Label: "Primary Language"},
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
				Title:       "Student Performance Prediction",
				Description: "A machine learning-based educational technology system that predicts student academic outcomes through three specialized models: final exam mark prediction, dropout risk assessment, and pass/fail forecasting. Built with Python, Flask, and scikit-learn to help educational institutions identify at-risk students and implement timely interventions.",
				Tags:        []string{"Python", "flask", "sklearn"},
				GitHub:      "https://github.com/Rashed-alothman/Student-Performance-Prediction",
			},
			{
				Title:       "TMS",
				Description: "Task Management System A flexible, lightweight task management system built with Python and Flask. Designed for personal productivity with a vision for collaborative team environments and seamless calendar integration.",
				Tags:        []string{"Python", "PostgreSQL", "Docker"},
				GitHub:      "#",
			},
			{
				Title:       "Snatch",
				Description: "A powerful terminal-based Python downloader for YouTube, Twitter, TikTok & more with high-quality video/audio extraction capabilities.",
				Tags:        []string{"Python", "Json", "CLI"},
				GitHub:      "https://github.com/Rashed-alothman/Shell-in-go",
			},
			{
				Title:       "Shell in go",
				Description: "A simple shell implemented in Go, supporting basic command execution, piping, and redirection.",
				Tags:        []string{"Go", "Json", "CLI"},
				GitHub:      "https://github.com/Rashed-alothman/Shell-in-go",
			},
		},
		Skills: Skills{
			Languages:  []string{"Go", "Python", "C++", "SQL", "Bash"},
			Frameworks: []string{"net/http", "Chi", "Gin", "GORM", "React", "flask", "Django", "sklearn"},
			Tools:      []string{"Docker", "PostgreSQL", "Redis", "Git", "Linux"},
		},
		Contact: Contact{
			Email:    "rashed.m.alothman@gmail.com",
			GitHub:   "https://github.com/Rashed-alothman",
			LinkedIn: "linkedin.com/in/rashed-alothman-09386a24a",
		},
	}
}

// ── Profiling ─────────────────────────────────────────────────────────────────

// startDebugServer starts a pprof profiling HTTP server on addr in a background
// goroutine. The blank import of net/http/pprof above registers all profiling
// handlers on http.DefaultServeMux; this server uses that mux (nil handler).
//
// The public server uses its own explicit ServeMux, so pprof endpoints are
// never reachable from the outside — only through this debug server.
//
// Useful commands once the server is running:
//
//	# Heap snapshot
//	go tool pprof http://localhost:6060/debug/pprof/heap
//
//	# 30-second CPU profile
//	go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30
//
//	# Goroutine dump
//	go tool pprof http://localhost:6060/debug/pprof/goroutine
//
//	# Live view in browser
//	open http://localhost:6060/debug/pprof/
func startDebugServer(addr string) {
	go func() {
		log.Printf("pprof listening on http://%s/debug/pprof/", addr)
		if err := http.ListenAndServe(addr, nil); err != nil {
			log.Printf("debug server stopped: %v", err)
		}
	}()
}

// ── Entry point ───────────────────────────────────────────────────────────────

func main() {
	srv, err := newServer()
	if err != nil {
		log.Fatal(err)
	}

	// Start the pprof debug server only when PPROF_ADDR is set.
	// Example: PPROF_ADDR=localhost:6060 go run main.go
	// Never set this in production unless you restrict access at the network level.
	if pprofAddr := os.Getenv("PPROF_ADDR"); pprofAddr != "" {
		startDebugServer(pprofAddr)
	}

	// PORT is injected by Railway, Fly.io, Render, and most cloud platforms.
	// Falls back to 8080 for local development.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Portfolio running on http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, srv.routes()))
}
