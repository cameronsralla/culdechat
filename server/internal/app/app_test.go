package app_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/cameronsralla/culdechat/server/internal/app"
	"github.com/cameronsralla/culdechat/server/internal/config"
)

// Tests run against a real Postgres (make up starts one). Set TEST_DATABASE_URL
// to point elsewhere; the database named in the URL is dropped and recreated.
const defaultTestURL = "postgres://postgres:postgres@localhost:5432/culdechat_test?sslmode=disable"

const (
	adminEmail = "admin@test.local"
	adminPass  = "admin-password-123"
)

var (
	srv    *httptest.Server
	appRef *app.App
)

func TestMain(m *testing.M) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		url = defaultTestURL
	}
	if err := recreateDB(url); err != nil {
		fmt.Fprintln(os.Stderr, "test db:", err, "(is the dev postgres running? make up)")
		os.Exit(1)
	}
	cfg := config.Config{
		Env: config.EnvDev, PublicURL: "http://web.test", DatabaseURL: url, DBMaxConns: 4,
		JWTSecret: []byte(strings.Repeat("t", 32)), AccessTokenTTL: 15 * time.Minute, RefreshTokenTTL: 24 * time.Hour,
		CORSOrigins: []string{"http://web.test"}, RequestTimeout: 10 * time.Second, MaxJSONBody: 1 << 20,
		RateLimitEnabled: false,
		BootstrapAdmin:   config.BootstrapAdmin{Email: adminEmail, Password: adminPass, Name: "Admin", Unit: "1"},
	}
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	a, err := app.New(context.Background(), cfg, log, "test")
	if err != nil {
		fmt.Fprintln(os.Stderr, "app:", err)
		os.Exit(1)
	}
	appRef = a
	srv = httptest.NewServer(a.Handler)
	code := m.Run()
	srv.Close()
	a.Close()
	os.Exit(code)
}

func recreateDB(url string) error {
	cfg, err := pgx.ParseConfig(url)
	if err != nil {
		return err
	}
	name := cfg.Database
	cfg.Database = "postgres"
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	conn, err := pgx.ConnectConfig(ctx, cfg)
	if err != nil {
		return err
	}
	defer conn.Close(ctx)
	if _, err := conn.Exec(ctx, fmt.Sprintf(`DROP DATABASE IF EXISTS %q WITH (FORCE)`, name)); err != nil {
		return err
	}
	_, err = conn.Exec(ctx, fmt.Sprintf(`CREATE DATABASE %q`, name))
	return err
}

// ---- helpers ----

type resp struct {
	Status int
	Body   map[string]any
	Raw    []byte
	Header http.Header
}

func call(t *testing.T, method, path, token string, body any) resp {
	t.Helper()
	var rd io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		rd = bytes.NewReader(b)
	}
	req, _ := http.NewRequest(method, srv.URL+path, rd)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(res.Body)
	out := resp{Status: res.StatusCode, Raw: raw, Header: res.Header}
	if len(raw) > 0 && raw[0] == '{' {
		_ = json.Unmarshal(raw, &out.Body)
	}
	return out
}

func want(t *testing.T, r resp, status int) {
	t.Helper()
	if r.Status != status {
		t.Fatalf("want %d got %d: %s", status, r.Status, r.Raw)
	}
}

func login(t *testing.T, email, pass string) (access, refresh string) {
	t.Helper()
	r := call(t, "POST", "/api/auth/login", "", map[string]string{"email": email, "password": pass})
	want(t, r, 200)
	return r.Body["access_token"].(string), r.Body["refresh_token"].(string)
}

// ---- tests ----

func TestHealth(t *testing.T) {
	want(t, call(t, "GET", "/healthz", "", nil), 200)
	want(t, call(t, "GET", "/readyz", "", nil), 200)
}

func TestSecurityAndCORS(t *testing.T) {
	r := call(t, "GET", "/healthz", "", nil)
	if r.Header.Get("X-Content-Type-Options") != "nosniff" || r.Header.Get("X-Request-Id") == "" {
		t.Fatalf("missing hardening headers: %v", r.Header)
	}

	req, _ := http.NewRequest("OPTIONS", srv.URL+"/api/me", nil)
	req.Header.Set("Origin", "http://web.test")
	req.Header.Set("Access-Control-Request-Method", "GET")
	res, _ := http.DefaultClient.Do(req)
	if res.StatusCode != 204 || res.Header.Get("Access-Control-Allow-Origin") != "http://web.test" {
		t.Fatalf("preflight allowed origin failed: %d %v", res.StatusCode, res.Header)
	}

	req, _ = http.NewRequest("OPTIONS", srv.URL+"/api/me", nil)
	req.Header.Set("Origin", "http://evil.test")
	req.Header.Set("Access-Control-Request-Method", "GET")
	res, _ = http.DefaultClient.Do(req)
	if res.StatusCode != 403 || res.Header.Get("Access-Control-Allow-Origin") != "" {
		t.Fatalf("preflight disallowed origin should 403: %d", res.StatusCode)
	}
}

func TestLoginAndMe(t *testing.T) {
	want(t, call(t, "GET", "/api/me", "", nil), 401)
	want(t, call(t, "GET", "/api/me", "garbage", nil), 401)

	r := call(t, "POST", "/api/auth/login", "", map[string]string{"email": adminEmail, "password": "wrong"})
	want(t, r, 401)
	r = call(t, "POST", "/api/auth/login", "", map[string]string{"email": "nobody@test.local", "password": "wrong"})
	want(t, r, 401)

	access, _ := login(t, adminEmail, adminPass)
	r = call(t, "GET", "/api/me", access, nil)
	want(t, r, 200)
	if r.Body["email"] != adminEmail || r.Body["is_admin"] != true {
		t.Fatalf("unexpected me: %s", r.Raw)
	}
	if _, leaked := r.Body["password_hash"]; leaked {
		t.Fatal("password hash leaked")
	}
}

func TestBindRejectsUnknownFields(t *testing.T) {
	r := call(t, "POST", "/api/auth/login", "", map[string]string{"email": adminEmail, "password": adminPass, "extra": "x"})
	want(t, r, 400)
}

func TestRefreshRotationAndReuse(t *testing.T) {
	_, refresh1 := login(t, adminEmail, adminPass)

	r := call(t, "POST", "/api/auth/refresh", "", map[string]string{"refresh_token": refresh1})
	want(t, r, 200)
	refresh2 := r.Body["refresh_token"].(string)
	if refresh2 == refresh1 {
		t.Fatal("refresh token not rotated")
	}

	// Reusing the old token must revoke the whole family.
	want(t, call(t, "POST", "/api/auth/refresh", "", map[string]string{"refresh_token": refresh1}), 401)
	want(t, call(t, "POST", "/api/auth/refresh", "", map[string]string{"refresh_token": refresh2}), 401)
}

func TestLogout(t *testing.T) {
	_, refresh := login(t, adminEmail, adminPass)
	want(t, call(t, "POST", "/api/auth/logout", "", map[string]string{"refresh_token": refresh}), 204)
	want(t, call(t, "POST", "/api/auth/refresh", "", map[string]string{"refresh_token": refresh}), 401)
}

func TestInviteFlow(t *testing.T) {
	admin, _ := login(t, adminEmail, adminPass)

	// Non-admin cannot invite (checked after we create one).
	r := call(t, "POST", "/api/admin/users/invite", admin, map[string]any{
		"email": "Resident@Test.local", "unit_number": "12B", "display_name": "",
	})
	want(t, r, 201)
	passcode := r.Body["passcode"].(string)
	inviteURL := r.Body["invite_url"].(string)
	token := inviteURL[strings.LastIndex(inviteURL, "token=")+6:]

	// Duplicate email conflicts (case-insensitive).
	r = call(t, "POST", "/api/admin/users/invite", admin, map[string]any{"email": "resident@test.local", "unit_number": "12B"})
	want(t, r, 409)

	// Peek shows who it's for.
	r = call(t, "GET", "/api/auth/invite?token="+token, "", nil)
	want(t, r, 200)
	if r.Body["email"] != "resident@test.local" {
		t.Fatalf("peek: %s", r.Raw)
	}

	// Cannot log in before completing.
	r = call(t, "POST", "/api/auth/login", "", map[string]string{"email": "resident@test.local", "password": "whatever-1234"})
	want(t, r, 401)

	// Wrong passcode.
	r = call(t, "POST", "/api/auth/complete-invite", "", map[string]string{
		"token": token, "passcode": "000000", "password": "resident-pass-1", "display_name": "Res",
	})
	if r.Status != 401 {
		// 1-in-a-million the random passcode is 000000; tolerate.
		if passcode != "000000" {
			t.Fatalf("wrong passcode should 401: %d %s", r.Status, r.Raw)
		}
	}

	// Complete.
	r = call(t, "POST", "/api/auth/complete-invite", "", map[string]string{
		"token": token, "passcode": passcode, "password": "resident-pass-1", "display_name": "Res Ident",
	})
	want(t, r, 201)
	resAccess := r.Body["access_token"].(string)

	// Token is single-use.
	r = call(t, "POST", "/api/auth/complete-invite", "", map[string]string{
		"token": token, "passcode": passcode, "password": "resident-pass-1", "display_name": "Res Ident",
	})
	want(t, r, 401)

	// Resident can see directory and self, not admin routes.
	want(t, call(t, "GET", "/api/me", resAccess, nil), 200)
	want(t, call(t, "GET", "/api/directory", resAccess, nil), 200)
	want(t, call(t, "GET", "/api/admin/users", resAccess, nil), 403)
	want(t, call(t, "POST", "/api/admin/users/invite", resAccess, map[string]any{"email": "x@test.local", "unit_number": "1"}), 403)

	// Profile update.
	r = call(t, "PATCH", "/api/me", resAccess, map[string]any{"display_name": "Resi", "directory_opt_in": false})
	want(t, r, 200)
	if r.Body["display_name"] != "Resi" {
		t.Fatalf("profile: %s", r.Raw)
	}

	// Admin roster lists both.
	r = call(t, "GET", "/api/admin/users", admin, nil)
	want(t, r, 200)
	var list []map[string]any
	_ = json.Unmarshal(r.Raw, &list)
	if len(list) != 2 {
		t.Fatalf("roster: %s", r.Raw)
	}

	// Deactivate resident: their session dies.
	var resID string
	for _, u := range list {
		if u["email"] == "resident@test.local" {
			resID = u["id"].(string)
		}
	}
	_, resRefresh := login(t, "resident@test.local", "resident-pass-1")
	want(t, call(t, "PUT", "/api/admin/users/"+resID+"/status", admin, map[string]bool{"active": false}), 200)
	want(t, call(t, "POST", "/api/auth/refresh", "", map[string]string{"refresh_token": resRefresh}), 401)
	r = call(t, "POST", "/api/auth/login", "", map[string]string{"email": "resident@test.local", "password": "resident-pass-1"})
	want(t, r, 403)

	// Admin cannot deactivate or demote self.
	r = call(t, "GET", "/api/me", admin, nil)
	adminID := r.Body["id"].(string)
	want(t, call(t, "PUT", "/api/admin/users/"+adminID+"/status", admin, map[string]bool{"active": false}), 409)
	want(t, call(t, "PUT", "/api/admin/users/"+adminID+"/admin", admin, map[string]bool{"is_admin": false}), 409)
}

func TestChangePassword(t *testing.T) {
	admin, _ := login(t, adminEmail, adminPass)
	r := call(t, "POST", "/api/auth/change-password", admin, map[string]string{"current_password": "nope", "new_password": "new-admin-pass-1"})
	want(t, r, 401)
	r = call(t, "POST", "/api/auth/change-password", admin, map[string]string{"current_password": adminPass, "new_password": "short"})
	want(t, r, 400)
	r = call(t, "POST", "/api/auth/change-password", admin, map[string]string{"current_password": adminPass, "new_password": "new-admin-pass-1"})
	want(t, r, 200)
	// Restore for other tests.
	newAccess := r.Body["access_token"].(string)
	want(t, call(t, "POST", "/api/auth/change-password", newAccess, map[string]string{"current_password": "new-admin-pass-1", "new_password": adminPass}), 200)
}

func TestSettings(t *testing.T) {
	admin, _ := login(t, adminEmail, adminPass)
	r := call(t, "GET", "/api/settings", admin, nil)
	want(t, r, 200)
	if r.Body["community_name"] != "Cul-de-Chat" {
		t.Fatalf("defaults: %s", r.Raw)
	}
	want(t, call(t, "PUT", "/api/admin/settings/community_name", admin, map[string]any{"value": "Maple Court"}), 200)
	want(t, call(t, "PUT", "/api/admin/settings/capability_tier", admin, map[string]any{"value": "ULTRA"}), 400)
	want(t, call(t, "PUT", "/api/admin/settings/bogus", admin, map[string]any{"value": 1}), 400)
	r = call(t, "GET", "/api/settings", admin, nil)
	if r.Body["community_name"] != "Maple Court" {
		t.Fatalf("cache not invalidated: %s", r.Raw)
	}
}

func TestNotFoundIsJSON(t *testing.T) {
	r := call(t, "GET", "/api/nope", "", nil)
	want(t, r, 404)
	if r.Body["error"] == nil {
		t.Fatalf("expected json error body: %s", r.Raw)
	}
}
