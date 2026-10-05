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
	"net/url"
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
	unit12B := ensureUnit(t, admin, "12B")
	r := call(t, "POST", "/api/admin/users/invite", admin, map[string]any{
		"email": "Resident@Test.local", "unit_id": unit12B, "display_name": "",
	})
	want(t, r, 201)
	passcode := r.Body["passcode"].(string)
	inviteURL := r.Body["invite_url"].(string)
	token := inviteURL[strings.LastIndex(inviteURL, "token=")+6:]

	// Duplicate email conflicts (case-insensitive).
	r = call(t, "POST", "/api/admin/users/invite", admin, map[string]any{"email": "resident@test.local", "unit_id": unit12B})
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
	list := userItems(t, r)
	if len(list) < 2 {
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

	// Reactivate restores login.
	want(t, call(t, "PUT", "/api/admin/users/"+resID+"/status", admin, map[string]bool{"active": true}), 200)
	_, _ = login(t, "resident@test.local", "resident-pass-1")

	// Admin cannot deactivate or demote self.
	r = call(t, "GET", "/api/me", admin, nil)
	adminID := r.Body["id"].(string)
	want(t, call(t, "PUT", "/api/admin/users/"+adminID+"/status", admin, map[string]bool{"active": false}), 409)
	want(t, call(t, "PUT", "/api/admin/users/"+adminID+"/admin", admin, map[string]bool{"is_admin": false}), 409)
}

func TestAdminResetPassword(t *testing.T) {
	admin, _ := login(t, adminEmail, adminPass)

	// Ensure a registered resident exists (invite flow may have run in other tests on a shared DB...
	// This suite recreates the DB per TestMain, but tests share the same DB within a run.
	r := call(t, "GET", "/api/admin/users?page_size=100", admin, nil)
	want(t, r, 200)
	list := userItems(t, r)
	var resID string
	for _, u := range list {
		if u["email"] == "resident@test.local" && u["status"] == "active" {
			resID = u["id"].(string)
		}
	}
	if resID == "" {
		// Create one if InviteFlow didn't leave an active resident (order-independent).
		r = call(t, "POST", "/api/admin/users/invite", admin, map[string]any{"email": "resetme@test.local", "unit_id": ensureUnit(t, admin, "9")})
		want(t, r, 201)
		passcode := r.Body["passcode"].(string)
		inviteURL := r.Body["invite_url"].(string)
		token := inviteURL[strings.LastIndex(inviteURL, "token=")+6:]
		r = call(t, "POST", "/api/auth/complete-invite", "", map[string]string{
			"token": token, "passcode": passcode, "password": "old-password-1", "display_name": "Reset Me",
		})
		want(t, r, 201)
		resID = r.Body["user"].(map[string]any)["id"].(string)
		_, oldRefresh := login(t, "resetme@test.local", "old-password-1")

		r = call(t, "POST", "/api/admin/users/"+resID+"/reset-password", admin, nil)
		want(t, r, 200)
		temp := r.Body["temporary_password"].(string)
		if len(temp) < 10 {
			t.Fatalf("temp password too short: %q", temp)
		}
		want(t, call(t, "POST", "/api/auth/refresh", "", map[string]string{"refresh_token": oldRefresh}), 401)
		want(t, call(t, "POST", "/api/auth/login", "", map[string]string{"email": "resetme@test.local", "password": "old-password-1"}), 401)
		_, _ = login(t, "resetme@test.local", temp)

		r = call(t, "GET", "/api/me", admin, nil)
		adminID := r.Body["id"].(string)
		want(t, call(t, "POST", "/api/admin/users/"+adminID+"/reset-password", admin, nil), 409)
		return
	}

	_, oldRefresh := login(t, "resident@test.local", "resident-pass-1")
	r = call(t, "POST", "/api/admin/users/"+resID+"/reset-password", admin, nil)
	want(t, r, 200)
	temp := r.Body["temporary_password"].(string)
	want(t, call(t, "POST", "/api/auth/refresh", "", map[string]string{"refresh_token": oldRefresh}), 401)
	want(t, call(t, "POST", "/api/auth/login", "", map[string]string{"email": "resident@test.local", "password": "resident-pass-1"}), 401)
	_, _ = login(t, "resident@test.local", temp)

	// Restore a known password for any later tests that might need the resident.
	resAccess, _ := login(t, "resident@test.local", temp)
	want(t, call(t, "POST", "/api/auth/change-password", resAccess, map[string]string{"current_password": temp, "new_password": "resident-pass-1"}), 200)

	r = call(t, "GET", "/api/me", admin, nil)
	adminID := r.Body["id"].(string)
	want(t, call(t, "POST", "/api/admin/users/"+adminID+"/reset-password", admin, nil), 409)
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

func userItems(t *testing.T, r resp) []map[string]any {
	t.Helper()
	var page struct {
		Items []map[string]any `json:"items"`
		Total int              `json:"total"`
	}
	if err := json.Unmarshal(r.Raw, &page); err != nil {
		t.Fatalf("users page: %v %s", err, r.Raw)
	}
	return page.Items
}

func TestAdminUserList(t *testing.T) {
	admin, _ := login(t, adminEmail, adminPass)
	for i, email := range []string{"pager-a@test.local", "pager-b@test.local"} {
		r := call(t, "POST", "/api/admin/users/invite", admin, map[string]any{"email": email, "unit_id": ensureUnit(t, admin, fmt.Sprintf("P%d", i+1))})
		if r.Status == 409 {
			continue
		}
		want(t, r, 201)
	}
	r := call(t, "GET", "/api/admin/users?q=pager-&page_size=1&page=1&sort=email&dir=asc", admin, nil)
	want(t, r, 200)
	var page struct {
		Items    []map[string]any `json:"items"`
		Total    int              `json:"total"`
		Page     int              `json:"page"`
		PageSize int              `json:"page_size"`
	}
	if err := json.Unmarshal(r.Raw, &page); err != nil {
		t.Fatal(err)
	}
	if page.Total != 2 || len(page.Items) != 1 || page.PageSize != 1 {
		t.Fatalf("page: %s", r.Raw)
	}
	if page.Items[0]["email"] != "pager-a@test.local" {
		t.Fatalf("sort: %s", r.Raw)
	}
	r = call(t, "GET", "/api/admin/users?q=pager-&page_size=1&page=2&sort=email&dir=asc", admin, nil)
	want(t, r, 200)
	items := userItems(t, r)
	if len(items) != 1 || items[0]["email"] != "pager-b@test.local" {
		t.Fatalf("page 2: %s", r.Raw)
	}
	// Column filter is applied with search, not only to the current page.
	r = call(t, "GET", "/api/admin/users?q=pager-&status=active&page_size=1", admin, nil)
	want(t, r, 200)
	var filtered struct {
		Total int `json:"total"`
	}
	_ = json.Unmarshal(r.Raw, &filtered)
	if filtered.Total != 0 {
		t.Fatalf("active filter should exclude invited pager users: %s", r.Raw)
	}
}

func TestDirectoryList(t *testing.T) {
	admin, _ := login(t, adminEmail, adminPass)
	ada := registerResident(t, admin, "ada@dir.test", "A1", "Ada Stone")
	registerResident(t, admin, "bea@dir.test", "B2", "Bea Stone")
	r := call(t, "POST", "/api/admin/users/invite", admin, map[string]any{"email": "cy@dir.test", "unit_id": ensureUnit(t, admin, "C3")})
	want(t, r, 201)

	r = call(t, "GET", "/api/directory?q=@dir.test&page_size=1&page=1&sort=email&dir=asc", ada, nil)
	want(t, r, 200)
	var page struct {
		Items []map[string]any `json:"items"`
		Total int              `json:"total"`
	}
	if err := json.Unmarshal(r.Raw, &page); err != nil {
		t.Fatal(err)
	}
	if page.Total != 2 || len(page.Items) != 1 || page.Items[0]["email"] != "ada@dir.test" {
		t.Fatalf("page: %s", r.Raw)
	}
	r = call(t, "GET", "/api/directory?q=@dir.test&page_size=1&page=2&sort=email&dir=asc", ada, nil)
	want(t, r, 200)
	if err := json.Unmarshal(r.Raw, &page); err != nil {
		t.Fatal(err)
	}
	if len(page.Items) != 1 || page.Items[0]["email"] != "bea@dir.test" {
		t.Fatalf("page 2: %s", r.Raw)
	}
	r = call(t, "GET", "/api/directory?q=@dir.test&name=Ada&page_size=1", ada, nil)
	want(t, r, 200)
	if err := json.Unmarshal(r.Raw, &page); err != nil {
		t.Fatal(err)
	}
	if page.Total != 1 || page.Items[0]["display_name"] != "Ada Stone" {
		t.Fatalf("name filter: %s", r.Raw)
	}

	r = call(t, "PATCH", "/api/me", ada, map[string]any{"display_name": "Ada Stone", "directory_opt_in": false})
	want(t, r, 200)
	r = call(t, "GET", "/api/directory?q=@dir.test&page_size=10", ada, nil)
	want(t, r, 200)
	if err := json.Unmarshal(r.Raw, &page); err != nil {
		t.Fatal(err)
	}
	if page.Total != 1 || page.Items[0]["email"] != "bea@dir.test" {
		t.Fatalf("opt-out name search: %s", r.Raw)
	}
	r = call(t, "GET", "/api/directory?unit=A1&page_size=10", ada, nil)
	want(t, r, 200)
	page.Items = nil
	if err := json.Unmarshal(r.Raw, &page); err != nil {
		t.Fatal(err)
	}
	if page.Total != 1 || page.Items[0]["kind"] != "unit" || page.Items[0]["unit_number"] != "A1" {
		t.Fatalf("hidden unit: %s", r.Raw)
	}
	if _, ok := page.Items[0]["id"]; ok {
		t.Fatalf("hidden unit leaked id: %s", r.Raw)
	}
	if _, ok := page.Items[0]["display_name"]; ok {
		t.Fatalf("hidden unit leaked name: %s", r.Raw)
	}
	if _, ok := page.Items[0]["email"]; ok {
		t.Fatalf("hidden unit leaked email: %s", r.Raw)
	}
	if page.Items[0]["self"] != true {
		t.Fatalf("own unit should be marked: %s", r.Raw)
	}
	r = call(t, "GET", "/api/directory?name=Ada&page_size=10", ada, nil)
	want(t, r, 200)
	if err := json.Unmarshal(r.Raw, &page); err != nil {
		t.Fatal(err)
	}
	if page.Total != 0 {
		t.Fatalf("name filter should not find a hidden resident: %s", r.Raw)
	}
}

func TestHiddenUnitMessage(t *testing.T) {
	admin, _ := login(t, adminEmail, adminPass)
	ada := registerResident(t, admin, "ada.msg@test.local", "M1", "Ada Msg")
	bea := registerResident(t, admin, "bea.msg@test.local", "M2", "Bea Msg")
	beaID := call(t, "GET", "/api/me", bea, nil).Body["id"].(string)
	r := call(t, "PATCH", "/api/me", bea, map[string]any{"display_name": "Bea Msg", "directory_opt_in": false})
	want(t, r, 200)

	r = call(t, "POST", "/api/messages", ada, map[string]any{"unit_number": "M2", "content": "Your package is at my door"})
	want(t, r, 200)
	if r.Body["status"] != "pending" {
		t.Fatalf("request: %s", r.Raw)
	}
	convID := r.Body["id"].(string)
	peer := r.Body["peer"].(map[string]any)
	if _, ok := peer["id"]; ok || peer["display_name"] != nil || peer["unit_number"] != "M2" {
		t.Fatalf("peer: %s", r.Raw)
	}
	if strings.Contains(string(r.Raw), beaID) {
		t.Fatalf("hidden id leaked: %s", r.Raw)
	}

	r = call(t, "POST", "/api/messages", ada, map[string]any{"unit_number": "M2", "content": "again"})
	want(t, r, 409)
	r = call(t, "POST", "/api/messages", bea, map[string]any{"conversation_id": convID, "content": "thanks"})
	want(t, r, 409)

	r = call(t, "GET", "/api/messages/conversations/"+convID, bea, nil)
	want(t, r, 200)
	if r.Body["incoming"] != true || r.Body["status"] != "pending" {
		t.Fatalf("incoming: %s", r.Raw)
	}
	msgs := r.Body["messages"].([]any)
	if len(msgs) != 1 || msgs[0].(map[string]any)["body"] != "Your package is at my door" {
		t.Fatalf("body: %s", r.Raw)
	}

	want(t, call(t, "POST", "/api/messages/conversations/"+convID+"/decline", bea, nil), 200)
	r = call(t, "GET", "/api/messages/conversations", bea, nil)
	want(t, r, 200)
	if strings.Contains(string(r.Raw), convID) {
		t.Fatalf("declined request still in recipient inbox: %s", r.Raw)
	}
	r = call(t, "GET", "/api/messages/conversations", ada, nil)
	want(t, r, 200)
	if !strings.Contains(string(r.Raw), `"status":"declined"`) && !strings.Contains(string(r.Raw), `"status": "declined"`) {
		t.Fatalf("sender should see the decline: %s", r.Raw)
	}

	r = call(t, "POST", "/api/messages", ada, map[string]any{"unit_number": "M2", "content": "Still at my door"})
	want(t, r, 200)
	if r.Body["status"] != "pending" {
		t.Fatalf("resend: %s", r.Raw)
	}
	want(t, call(t, "POST", "/api/messages/conversations/"+convID+"/accept", bea, nil), 200)
	r = call(t, "POST", "/api/messages", bea, map[string]any{"conversation_id": convID, "content": "I'll come by"})
	want(t, r, 200)
	if r.Body["status"] != "open" {
		t.Fatalf("reply: %s", r.Raw)
	}
	r = call(t, "GET", "/api/messages/conversations/"+convID, ada, nil)
	want(t, r, 200)
	if strings.Contains(string(r.Raw), beaID) || strings.Contains(string(r.Raw), "Bea Msg") {
		t.Fatalf("hidden identity leaked after accept: %s", r.Raw)
	}
	msgs = r.Body["messages"].([]any)
	if len(msgs) != 3 || msgs[2].(map[string]any)["body"] != "I'll come by" {
		t.Fatalf("thread: %s", r.Raw)
	}
}

func TestUnits(t *testing.T) {
	admin, _ := login(t, adminEmail, adminPass)
	r := call(t, "POST", "/api/admin/units", admin, map[string]any{"number": "Z1"})
	want(t, r, 201)
	z1 := r.Body["id"].(string)
	if residents, ok := r.Body["residents"].([]any); !ok || len(residents) != 0 {
		t.Fatalf("new unit should be empty: %s", r.Raw)
	}
	want(t, call(t, "POST", "/api/admin/units", admin, map[string]any{"number": "Z1"}), 409)

	r = call(t, "POST", "/api/admin/users/invite", admin, map[string]any{"email": "z1@test.local", "unit_id": z1})
	want(t, r, 201)
	if r.Body["user"].(map[string]any)["is_primary"] != true || r.Body["user"].(map[string]any)["unit_number"] != "Z1" {
		t.Fatalf("invite: %s", r.Raw)
	}
	want(t, call(t, "POST", "/api/admin/users/invite", admin, map[string]any{"email": "z1b@test.local", "unit_id": z1}), 409)
	want(t, call(t, "DELETE", "/api/admin/units/"+z1, admin, nil), 409)

	z2 := ensureUnit(t, admin, "Z2")
	want(t, call(t, "DELETE", "/api/admin/units/"+z2, admin, nil), 204)
	r = call(t, "GET", "/api/admin/units?vacant=true&q=Z1&page_size=10", admin, nil)
	want(t, r, 200)
	var page struct {
		Items []map[string]any `json:"items"`
		Total int              `json:"total"`
	}
	if err := json.Unmarshal(r.Raw, &page); err != nil {
		t.Fatal(err)
	}
	if page.Total != 0 {
		t.Fatalf("occupied unit should not be vacant: %s", r.Raw)
	}
}

func ensureUnit(t *testing.T, admin, number string) string {
	t.Helper()
	r := call(t, "POST", "/api/admin/units", admin, map[string]any{"number": number})
	if r.Status == 409 {
		r = call(t, "GET", "/api/admin/units?q="+url.QueryEscape(number)+"&page_size=100", admin, nil)
		want(t, r, 200)
		var page struct {
			Items []map[string]any `json:"items"`
		}
		if err := json.Unmarshal(r.Raw, &page); err != nil {
			t.Fatal(err)
		}
		for _, item := range page.Items {
			if item["number"] == number {
				return item["id"].(string)
			}
		}
		t.Fatalf("unit %s not found: %s", number, r.Raw)
	}
	want(t, r, 201)
	return r.Body["id"].(string)
}

func registerResident(t *testing.T, admin, email, unit, name string) string {
	t.Helper()
	r := call(t, "POST", "/api/admin/users/invite", admin, map[string]any{
		"email": email, "unit_id": ensureUnit(t, admin, unit), "display_name": name,
	})
	want(t, r, 201)
	passcode := r.Body["passcode"].(string)
	inviteURL := r.Body["invite_url"].(string)
	token := inviteURL[strings.LastIndex(inviteURL, "token=")+6:]
	r = call(t, "POST", "/api/auth/complete-invite", "", map[string]string{
		"token": token, "passcode": passcode, "password": "resident-pass-1", "display_name": name,
	})
	want(t, r, 201)
	return r.Body["access_token"].(string)
}
