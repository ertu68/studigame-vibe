package handlers

import (
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"studigame/internal/auth"
	"studigame/internal/config"
)

// gameLoopServer wires the routes of the game loop against a temp database and
// a fixed clock, so a test can walk the same path a browser would.
type gameLoopServer struct {
	t      *testing.T
	deps   *Deps
	server *httptest.Server
	client *http.Client
	now    time.Time
}

func newGameLoopServer(t *testing.T) *gameLoopServer {
	t.Helper()
	conn := testDB(t)

	gs := &gameLoopServer{
		t:   t,
		now: time.Date(2026, 8, 26, 8, 0, 0, 0, time.UTC),
	}
	gs.deps = &Deps{
		DB:       conn,
		Sessions: auth.NewSessionStore(conn),
		Mailer:   auth.ConsoleMailer{},
		Config:   config.Config{SiteName: "Testspiel", SiteDomain: "example.test"},
		Now:      func() time.Time { return gs.now },
		Roll:     func() float64 { return 0.5 },
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /login", gs.deps.LoginForm)
	mux.HandleFunc("POST /login", RequireCSRF(gs.deps.Login))
	mux.HandleFunc("GET /character/new", gs.deps.RequireAuth(gs.deps.CharacterNewForm))
	mux.HandleFunc("POST /character/new", RequireCSRF(gs.deps.RequireAuth(gs.deps.CharacterNew)))
	mux.HandleFunc("GET /dashboard", gs.deps.RequireAuth(gs.deps.Dashboard))
	mux.HandleFunc("GET /actions", gs.deps.RequireAuth(gs.deps.Actions))
	mux.HandleFunc("POST /actions/collect", RequireCSRF(gs.deps.RequireAuth(gs.deps.Collect)))

	gs.server = httptest.NewServer(mux)
	t.Cleanup(gs.server.Close)

	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatalf("cookie jar: %v", err)
	}
	gs.client = &http.Client{Jar: jar}
	return gs
}

func (gs *gameLoopServer) get(path string) (*http.Response, string) {
	gs.t.Helper()
	resp, err := gs.client.Get(gs.server.URL + path)
	if err != nil {
		gs.t.Fatalf("GET %s: %v", path, err)
	}
	return resp, readBody(gs.t, resp)
}

func (gs *gameLoopServer) post(path string, form url.Values) (*http.Response, string) {
	gs.t.Helper()
	form.Set("csrf_token", gs.csrfToken())
	resp, err := gs.client.PostForm(gs.server.URL+path, form)
	if err != nil {
		gs.t.Fatalf("POST %s: %v", path, err)
	}
	return resp, readBody(gs.t, resp)
}

// csrfToken reads the double-submit cookie the server handed out.
func (gs *gameLoopServer) csrfToken() string {
	gs.t.Helper()
	u, _ := url.Parse(gs.server.URL)
	for _, c := range gs.client.Jar.Cookies(u) {
		if c.Name == "csrf_token" {
			return c.Value
		}
	}
	gs.t.Fatal("no csrf cookie issued")
	return ""
}

func readBody(t *testing.T, resp *http.Response) string {
	t.Helper()
	defer resp.Body.Close()
	var sb strings.Builder
	buf := make([]byte, 4096)
	for {
		n, err := resp.Body.Read(buf)
		sb.Write(buf[:n])
		if err != nil {
			break
		}
	}
	return sb.String()
}

// registerVerified creates a user that can log in straight away.
func (gs *gameLoopServer) registerVerified(name, email string) {
	gs.t.Helper()
	if _, err := auth.Register(gs.deps.DB, name, email, "hunter22", "Köln"); err != nil {
		gs.t.Fatalf("register: %v", err)
	}
	if _, err := gs.deps.DB.Exec(`UPDATE users SET verified = TRUE WHERE email = ?`, email); err != nil {
		gs.t.Fatalf("verify user: %v", err)
	}
}

// subjectID resolves a Studienfach name to its id for the form.
func (gs *gameLoopServer) subjectID(name string) string {
	gs.t.Helper()
	var id int64
	err := gs.deps.DB.QueryRow(
		`SELECT s.id FROM subjects s JOIN editions e ON e.id = s.edition_id
		 WHERE e.slug = 'studistadt' AND s.name = ?`, name).Scan(&id)
	if err != nil {
		gs.t.Fatalf("resolve subject %q: %v", name, err)
	}
	return strconv.FormatInt(id, 10)
}

// TestGameLoop walks login, character creation, a collect shift and its payout.
func TestGameLoop(t *testing.T) {
	gs := newGameLoopServer(t)
	gs.registerVerified("Tessa", "tessa@example.com")

	// The form issues the CSRF cookie the POST needs.
	gs.get("/login")
	resp, _ := gs.post("/login", url.Values{
		"email":    {"tessa@example.com"},
		"password": {"hunter22"},
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("login: status %d", resp.StatusCode)
	}

	// Without a character the dashboard sends the player to onboarding.
	resp, body := gs.get("/dashboard")
	if !strings.Contains(body, "Neu anfangen") {
		t.Fatalf("expected onboarding, got status %d body %q", resp.StatusCode, truncate(body))
	}

	resp, body = gs.post("/character/new", url.Values{
		"subject": {gs.subjectID("Informatik")},
		"name":    {"Tessa"},
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("create character: status %d", resp.StatusCode)
	}
	if !strings.Contains(body, "Informatik") {
		t.Fatalf("dashboard missing subject: %q", truncate(body))
	}
	if !strings.Contains(body, "10,00 €") {
		t.Fatalf("dashboard missing start capital: %q", truncate(body))
	}

	// Start a one hour shift.
	resp, body = gs.post("/actions/collect", url.Values{"duration": {"3600"}})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("start collect: status %d", resp.StatusCode)
	}
	if !strings.Contains(body, "Unterwegs") {
		t.Fatalf("expected a running shift: %q", truncate(body))
	}

	// Halfway through nothing has been paid out.
	gs.now = gs.now.Add(30 * time.Minute)
	_, body = gs.get("/actions")
	if strings.Contains(body, "Schicht vorbei") {
		t.Fatalf("shift resolved early: %q", truncate(body))
	}
	if !strings.Contains(body, "30 min") {
		t.Fatalf("expected remaining time: %q", truncate(body))
	}

	// The page schedules its own reload for the moment the shift is due.
	if !strings.Contains(body, `http-equiv="refresh" content="1801"`) {
		t.Fatalf("expected a meta refresh for the remaining 30 min: %q", truncate(body))
	}

	// After the shift the payout shows up once.
	gs.now = gs.now.Add(30 * time.Minute)
	_, body = gs.get("/actions")
	if !strings.Contains(body, "Schicht vorbei") {
		t.Fatalf("expected the payout: %q", truncate(body))
	}
	if !strings.Contains(body, "40 Flaschen") {
		t.Fatalf("expected 40 bottles: %q", truncate(body))
	}
	if strings.Contains(body, "http-equiv=\"refresh\"") {
		t.Fatalf("idle page still reloads itself: %q", truncate(body))
	}

	// Reloading does not pay a second time.
	_, body = gs.get("/actions")
	if strings.Contains(body, "Schicht vorbei") {
		t.Fatalf("payout shown twice: %q", truncate(body))
	}

	// 10 € start capital plus 10 € earned.
	_, body = gs.get("/dashboard")
	if !strings.Contains(body, "20,00 €") {
		t.Fatalf("dashboard missing new balance: %q", truncate(body))
	}
}

// TestGameLoopRequiresLogin checks the game routes are not open to anonymous visitors.
func TestGameLoopRequiresLogin(t *testing.T) {
	gs := newGameLoopServer(t)
	gs.client.CheckRedirect = func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}

	for _, path := range []string{"/dashboard", "/actions", "/character/new"} {
		resp, _ := gs.get(path)
		if resp.StatusCode != http.StatusSeeOther {
			t.Errorf("GET %s: status %d, want 303", path, resp.StatusCode)
		}
		if loc := resp.Header.Get("Location"); loc != "/login" {
			t.Errorf("GET %s: redirected to %q, want /login", path, loc)
		}
	}
}

func truncate(s string) string {
	if len(s) > 300 {
		return s[:300] + "..."
	}
	return s
}
