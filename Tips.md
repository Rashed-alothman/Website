# Backend Development Tips — Go Edition

A reference for concepts, mythology, and ideas you will encounter while
learning backend development and Go. Read this alongside building things —
theory without practice fades fast.

---

## Part 1 — Go Language Fundamentals

### The things that make Go different from everything else

**Zero values** — Every variable in Go has a default "zero" value when
declared without initialization. `int` is `0`, `string` is `""`,
`bool` is `false`, pointers and interfaces are `nil`. This eliminates an
entire class of "uninitialized variable" bugs.

**Error handling by return value** — Go has no exceptions. Functions
return an `error` as the last return value. You handle it explicitly.
```go
data, err := os.ReadFile("file.txt")
if err != nil {
    // something went wrong — handle it here, not in some catch block far away
    log.Fatal(err)
}
```
This feels verbose at first. It becomes an advantage: you always know
exactly which operations can fail and exactly where you handle them.

**Goroutines — cheap concurrency** — A goroutine is a function running
concurrently, started with the `go` keyword. Go can run millions of
them because they are managed by the Go runtime, not the OS.
```go
go doSomethingSlowly() // starts immediately, does not block
```

**Channels — communicate between goroutines** — Do not share memory and
protect it with a lock. Instead, send data through channels.
```go
ch := make(chan string)
go func() { ch <- "done" }()
result := <-ch // blocks until something is sent
```

**Interfaces — structural typing** — In Go, a type satisfies an
interface automatically if it implements the required methods. No
`implements` keyword needed.
```go
// Any type with a Write([]byte)(int, error) method satisfies io.Writer.
// http.ResponseWriter, os.File, bytes.Buffer — all satisfy it.
```

**Defer** — Schedules a function call to run when the surrounding
function returns, no matter how it returns (normal or panic).
```go
f, _ := os.Open("file.txt")
defer f.Close() // runs when the function exits — you cannot forget to close
```

**No inheritance — composition instead** — Go has no class hierarchy.
You build complex types by embedding simpler ones.
```go
type Server struct {
    http.Server           // embed: Server now has all of http.Server's fields and methods
    db       *sql.DB
    logger   *log.Logger
}
```

---

## Part 2 — The net/http Package

This is your foundation. Learn it before adding any framework.

### How an HTTP server works in Go

```go
// A handler is anything with a ServeHTTP(w, r) method.
// The simplest way: use http.HandlerFunc, which turns a function into a handler.

http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    // w  = what you write to (the response)
    // r  = what came in (method, URL, headers, body)

    w.Header().Set("Content-Type", "text/plain")
    w.WriteHeader(http.StatusOK) // 200 — optional, 200 is the default
    w.Write([]byte("Hello, world"))
})

http.ListenAndServe(":8080", nil) // nil = use the default ServeMux
```

### Reading request data

```go
// Query string: /search?q=golang&page=2
q := r.URL.Query().Get("q") // "golang"

// Form body (POST, application/x-www-form-urlencoded)
r.ParseForm()
name := r.FormValue("name")

// JSON body
var body struct{ Name string }
json.NewDecoder(r.Body).Decode(&body)

// URL path parameter — with Chi router: /users/{id}
id := chi.URLParam(r, "id")

// Headers
token := r.Header.Get("Authorization")
```

### Writing responses

```go
// JSON response
w.Header().Set("Content-Type", "application/json")
json.NewEncoder(w).Encode(map[string]string{"status": "ok"})

// Redirect
http.Redirect(w, r, "/login", http.StatusFound) // 302

// Error
http.Error(w, "Not Found", http.StatusNotFound) // 404

// Render an HTML template (see Part 3)
tmpl.Execute(w, data)
```

### Middleware — functions that wrap handlers

Middleware is a function that takes a handler and returns a handler.
It runs code before and/or after the inner handler.
```go
func logging(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        next.ServeHTTP(w, r)   // call the real handler
        log.Printf("%s %s — %v", r.Method, r.URL.Path, time.Since(start))
    })
}

// Use it:
http.Handle("/", logging(myHandler))
```

Common middleware you will want to write or find a library for:
- Logging (log every request)
- Authentication (check JWT or session cookie)
- Rate limiting (reject too many requests from one IP)
- CORS (allow browser requests from a different domain)
- Recovery (catch panics and return 500 instead of crashing)

---

## Part 3 — html/template

### Why html/template, not text/template

`html/template` is identical to `text/template` in syntax but adds
**context-aware escaping**. It understands whether a value is going
into an HTML attribute, an HTML body, a CSS property, or a JavaScript
string, and escapes accordingly. Use `html/template` for any page
shown to a browser. Use `text/template` for emails, config files, etc.

### Template syntax cheat sheet

```
{{.}}                    output the current value
{{.Field}}               access a struct field or map key
{{.Method}}              call a method with no arguments
{{if .Field}} ... {{end}}             conditional
{{if .Field}} ... {{else}} ... {{end}} conditional with fallback
{{range .Slice}} ... {{end}}          iterate — {{.}} is the current element
{{range $i, $v := .Slice}} ...        iterate with index
{{$variable := .Field}}               assign to a variable
{{/* this is a comment */}}           comment — not sent to browser
```

### Passing multiple pieces of data

You cannot call `Execute` with two arguments. Instead, bundle
everything into a single struct (what this project does) or use a
`map[string]any`:
```go
tmpl.Execute(w, map[string]any{
    "Title": "My Page",
    "User":  currentUser,
})
// In template: {{.Title}}, {{.User.Name}}
```

### Template functions

You can add custom functions to a template:
```go
funcMap := template.FuncMap{
    "upper": strings.ToUpper,
    "add":   func(a, b int) int { return a + b },
}
tmpl := template.Must(template.New("page").Funcs(funcMap).ParseFiles("page.html"))
// In template: {{upper .Name}}, {{add 1 2}}
```

Built-in template functions: `and`, `or`, `not`, `eq`, `ne`, `lt`,
`le`, `gt`, `ge`, `len`, `index`, `print`, `printf`, `println`, `html`, `js`, `urlquery`

---

## Part 4 — HTTP Fundamentals

### Methods (verbs)

| Method | Purpose                          | Body? |
|--------|----------------------------------|-------|
| GET    | Read/fetch a resource            | No    |
| POST   | Create a resource or submit data | Yes   |
| PUT    | Replace a resource entirely      | Yes   |
| PATCH  | Partially update a resource      | Yes   |
| DELETE | Remove a resource                | No    |

### Status codes you will use constantly

| Code | Meaning               | When to use                              |
|------|-----------------------|------------------------------------------|
| 200  | OK                    | Successful GET, successful anything      |
| 201  | Created               | Successfully created a new resource      |
| 204  | No Content            | Successful DELETE, nothing to return     |
| 301  | Moved Permanently     | Permanent redirect (update bookmarks)    |
| 302  | Found (temp redirect) | Temporary redirect — after POST/redirect |
| 400  | Bad Request           | Client sent malformed or invalid data    |
| 401  | Unauthorized          | Not logged in                            |
| 403  | Forbidden             | Logged in but not allowed                |
| 404  | Not Found             | Resource does not exist                  |
| 409  | Conflict              | Duplicate resource (e.g. username taken) |
| 422  | Unprocessable Entity  | Validation failed                        |
| 429  | Too Many Requests     | Rate limit exceeded                      |
| 500  | Internal Server Error | Something blew up on the server          |

### Headers worth knowing

```
Content-Type: application/json          what format is the body
Authorization: Bearer <token>           authentication token
Set-Cookie: session=abc; HttpOnly       server sets a cookie
Cookie: session=abc                     browser sends cookies back
Cache-Control: no-cache, max-age=3600   caching instructions
X-Request-Id: uuid                      for tracing requests in logs
```

---

## Part 5 — REST API Design

REST is a set of conventions for building web APIs using HTTP the way
it was intended. It is not a specification — it is a style.

### Core idea: resources + methods

A REST API represents data as resources (nouns) accessed via URLs.
HTTP methods (verbs) express what you want to do to them.

```
GET    /users          → list all users
POST   /users          → create a new user
GET    /users/42       → get user with id 42
PUT    /users/42       → replace user 42 entirely
PATCH  /users/42       → update some fields of user 42
DELETE /users/42       → delete user 42
GET    /users/42/posts → get all posts by user 42
```

### Rules to follow

- URLs use nouns, not verbs. `/users`, not `/getUsers`.
- Plural nouns for collections: `/articles`, not `/article`.
- Return appropriate status codes — do not always return 200.
- Return JSON with a consistent shape. An error always looks the same:
  `{"error": "email already in use"}`, not something different each time.
- Version your API if others depend on it: `/api/v1/users`.

### What makes an API RESTful vs not

Stateless: the server stores no session between requests. Every request
carries all the information the server needs (token in header, etc.).
This is what makes REST APIs easy to scale horizontally.

---

## Part 6 — Databases with Go

### database/sql — the standard library interface

`database/sql` is an abstraction layer. You pair it with a driver for
your specific database (PostgreSQL, MySQL, SQLite).

```go
import (
    "database/sql"
    _ "github.com/lib/pq" // PostgreSQL driver — blank import registers it
)

db, err := sql.Open("postgres", "postgres://user:pass@localhost/dbname?sslmode=disable")

// Query multiple rows
rows, err := db.Query("SELECT id, name FROM users WHERE active = $1", true)
defer rows.Close()
for rows.Next() {
    var id   int
    var name string
    rows.Scan(&id, &name)
}

// Query one row
var name string
db.QueryRow("SELECT name FROM users WHERE id = $1", 42).Scan(&name)

// Execute (INSERT, UPDATE, DELETE)
result, err := db.Exec("INSERT INTO users (name, email) VALUES ($1, $2)", name, email)
id, _ := result.LastInsertId()
```

### GORM — ORM (optional, opinionated)

GORM maps Go structs to database tables automatically. Faster to
prototype with, harder to control exactly what SQL runs.

```go
import "gorm.io/gorm"

type User struct {
    gorm.Model             // adds ID, CreatedAt, UpdatedAt, DeletedAt
    Name  string
    Email string `gorm:"unique"`
}

db.Create(&User{Name: "Rashed", Email: "me@example.com"})
db.First(&user, 1)                            // SELECT * FROM users WHERE id = 1
db.Where("name = ?", "Rashed").Find(&users)   // SELECT * WHERE name = ?
db.Save(&user)                                // UPDATE
db.Delete(&user)                              // soft delete (sets DeletedAt)
```

### SQL concepts you must understand

- **Transactions** — group multiple queries so they all succeed or all
  fail together. Critical for anything involving money or consistency.
- **Indexes** — B-tree structures that make lookups fast. Add an index
  on any column you filter or join on. Without it, every query scans
  the entire table.
- **Foreign keys** — a column that references another table's primary
  key. Enforces referential integrity at the database level.
- **N+1 problem** — querying for a list of items, then querying for
  related data per item in a loop. Kills performance. Fix with JOINs or
  eager loading.
- **Migrations** — version-controlled SQL files that evolve your schema
  over time. Never alter a production table by hand. Use a migration
  tool (golang-migrate, Goose).

---

## Part 7 — Authentication Patterns

### Sessions (cookie-based)

1. User logs in → server creates a session record in the database,
   sets a `Set-Cookie` header with a random session ID.
2. Browser stores the cookie, sends it on every subsequent request.
3. Server looks up the session ID in the database to find the user.

Good for: server-rendered websites (like this portfolio with a CMS).

### JWT (JSON Web Tokens)

1. User logs in → server creates and signs a JWT containing user claims
   (user ID, role, expiry). Returns it to the client.
2. Client stores the JWT (usually in memory or localStorage for SPAs).
3. Client sends `Authorization: Bearer <token>` on every request.
4. Server verifies the signature — no database lookup needed.

Good for: APIs consumed by a frontend or mobile app.

**Never store JWTs in cookies without the `HttpOnly` and `Secure` flags.**
Never store sensitive data in the payload — the payload is base64-encoded,
not encrypted. Anyone can decode it.

### Password storage — the law

Never store passwords in plain text. Ever. Not even "temporarily".
Use a proper hashing algorithm designed for passwords:
- **bcrypt** — the standard choice, available in `golang.org/x/crypto/bcrypt`
- **argon2** — newer, more configurable, more memory-hard

```go
import "golang.org/x/crypto/bcrypt"

// When user registers:
hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
// Store hash in database

// When user logs in:
err := bcrypt.CompareHashAndPassword(hash, []byte(inputPassword))
// err == nil means password matches
```

---

## Part 8 — Performance and Concurrency in Go

### The key mental model

Go's goroutines are cheap. One goroutine per HTTP request is the
standard pattern. Go's scheduler multiplexes thousands of goroutines
onto a handful of OS threads.

```go
// The Go HTTP server already does this for you — each incoming request
// runs in its own goroutine automatically.
```

### Common concurrency patterns

**Worker pool** — limit concurrent operations (e.g. max 10 DB queries at once):
```go
sem := make(chan struct{}, 10) // buffered channel as semaphore
for _, item := range items {
    sem <- struct{}{} // blocks when 10 are already running
    go func(i Item) {
        defer func() { <-sem }()
        process(i)
    }(item)
}
```

**sync.WaitGroup** — wait for a group of goroutines to finish:
```go
var wg sync.WaitGroup
for _, url := range urls {
    wg.Add(1)
    go func(u string) {
        defer wg.Done()
        fetch(u)
    }(url)
}
wg.Wait() // block until all goroutines call Done()
```

**sync.Mutex** — protect shared data from concurrent writes:
```go
var mu sync.Mutex
var count int

mu.Lock()
count++
mu.Unlock()
```

Prefer channels for communication. Use mutexes for protecting shared state.

### Profiling

Go has profiling built in. Add this import to any server:
```go
import _ "net/http/pprof"
// Then visit: http://localhost:8080/debug/pprof/
```

---

## Part 9 — Docker and Deployment

### Why Docker

Docker packages your application and all its dependencies into a
container that runs identically on any machine. No more "it works on
my machine".

### A minimal Dockerfile for a Go binary

```dockerfile
# Stage 1: build the binary
FROM golang:1.22 AS builder
WORKDIR /app
COPY . .
RUN go build -o server .

# Stage 2: run it in a tiny image
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/server .
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/static    ./static
EXPOSE 8080
CMD ["./server"]
```

The two-stage build keeps the final image small (~10MB) by discarding
the Go toolchain after compilation.

### Environment variables — the 12-factor way

Never hardcode secrets (DB passwords, API keys) in source code.
Read them from environment variables instead.

```go
import "os"

dbURL := os.Getenv("DATABASE_URL")
port  := os.Getenv("PORT")
if port == "" {
    port = "8080" // default for local dev
}
```

Use a `.env` file locally (never commit it), and set real env vars
in your deployment environment.

---

## Part 10 — Security Fundamentals

These are the things that will bite you if you ignore them.

### SQL Injection

Never concatenate user input into SQL strings.
```go
// WRONG — attacker sends name = "'; DROP TABLE users; --"
query := "SELECT * FROM users WHERE name = '" + name + "'"

// RIGHT — use parameterized queries, always
db.Query("SELECT * FROM users WHERE name = $1", name)
```

### Cross-Site Scripting (XSS)

Never put unsanitized user input into HTML. `html/template` handles
this automatically — another reason to use it.

### Cross-Site Request Forgery (CSRF)

An attacker tricks a logged-in user into submitting a form to your
server. Defense: include a random CSRF token in every form and verify
it on POST. Libraries like `gorilla/csrf` handle this.

### Rate Limiting

Without rate limiting, anyone can hammer your server with thousands of
requests. Protect expensive endpoints (login, registration, search).
Use `golang.org/x/time/rate` for a token bucket limiter.

### HTTPS in production

Never run HTTP in production. Use a reverse proxy (Caddy, Nginx) that
handles TLS termination, or deploy to a platform that provides it
(Railway, Render, Fly.io).

---

## Part 11 — Go Packages Worth Knowing

| Package                    | What it does                              |
|----------------------------|-------------------------------------------|
| `net/http`                 | HTTP client and server — standard library |
| `html/template`            | Templating with XSS protection            |
| `database/sql`             | Database interface — standard library     |
| `encoding/json`            | JSON encode/decode — standard library     |
| `os`, `io`, `bufio`        | File and I/O — standard library           |
| `context`                  | Cancellation and timeouts across calls    |
| `log/slog`                 | Structured logging (Go 1.21+)             |
| `github.com/go-chi/chi`    | Lightweight router with middleware        |
| `github.com/lib/pq`        | PostgreSQL driver                         |
| `gorm.io/gorm`             | ORM for databases                         |
| `golang.org/x/crypto`      | bcrypt and other crypto not in stdlib     |
| `github.com/joho/godotenv` | Load .env files into os.Getenv            |
| `github.com/golang-migrate` | Database migrations                      |

---

## Part 12 — Recommended Learning Path

The order matters. Each step builds on the last.

1. **Write a CLI tool in Go** — read input, transform it, print output.
   Get comfortable with error handling and types before adding HTTP.

2. **Build a basic HTTP server** — study `net/http` directly, not a
   framework. Understand handlers, routing, middleware, and how requests
   flow before you add abstractions on top.

3. **Serve an HTML template** — what this project does. Connect Go data
   to HTML. Understand how `Execute` works.

4. **Add a database** — connect PostgreSQL, write basic CRUD queries,
   understand migrations. Use `database/sql` first, GORM later.

5. **Build a REST API** — JSON in, JSON out. Authentication with JWT.
   Proper status codes. Middleware for logging and auth.

6. **Learn Docker** — containerize what you built. Deploy it.

7. **Add observability** — structured logging, metrics, health check
   endpoint. You cannot improve what you cannot measure.

8. **Study system design** — caching, load balancing, message queues,
   database replication. These matter once your app has real traffic.

---

## Mental Models That Stick

**"Make it work, make it right, make it fast" — in that order.**
Do not optimize before you have something running. Profile before
you guess where the bottleneck is.

**Errors are values** — in Go, treat errors like any other return value.
An ignored error is a ticking time bomb. Wrap errors with context:
`fmt.Errorf("loading user %d: %w", id, err)`

**Fail loud at startup, not silently at runtime** — validate
configuration, parse templates, and open database connections when
the server starts. If anything is wrong, crash immediately with a
clear message rather than serving broken responses hours later.

**Read the standard library** — Go's standard library is one of the
best-designed in any language. Before reaching for a third-party
package, check if the stdlib already does what you need.
