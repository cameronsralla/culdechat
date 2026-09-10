package routes

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/cameronsralla/culdechat/internal/testutil"
	"github.com/cameronsralla/culdechat/models"
	"github.com/cameronsralla/culdechat/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}

func TestHealth(t *testing.T) {
	testutil.Setup(t)
	rec := doJSON(t, NewRouter(), http.MethodGet, "/api/health", "", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("health status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestAuthInviteAndComplete(t *testing.T) {
	testutil.Setup(t)
	r := NewRouter()
	_, adminToken := seedAdmin(t)

	rec := doJSON(t, r, http.MethodPost, "/api/auth/register", "", map[string]string{
		"email": "resident@test.local", "unit_number": "101",
	})
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("register without auth should 401, got %d", rec.Code)
	}

	rec = doJSON(t, r, http.MethodPost, "/api/auth/register", adminToken, map[string]string{
		"email": "resident@test.local", "unit_number": "101",
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("register status=%d body=%s", rec.Code, rec.Body.String())
	}
	var invited struct {
		Token     string `json:"registration_token"`
		Passcode  string `json:"passcode"`
		EmailSent bool   `json:"email_sent"`
	}
	decode(t, rec, &invited)
	if invited.EmailSent {
		t.Fatal("tests run without SMTP; email_sent should be false")
	}

	rec = doJSON(t, r, http.MethodPost, "/api/auth/login", "", map[string]string{
		"email": "resident@test.local", "password": "secret-pass",
	})
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("pending login should fail, got %d %s", rec.Code, rec.Body.String())
	}

	rec = doJSON(t, r, http.MethodPost, "/api/auth/complete-registration", "", map[string]string{
		"token": invited.Token, "passcode": invited.Passcode, "password": "secret-pass", "name": "Alex Rivera",
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("complete status=%d body=%s", rec.Code, rec.Body.String())
	}
	var auth struct {
		Token string `json:"token"`
		User  struct {
			Name       string `json:"name"`
			UnitNumber string `json:"unit_number"`
		} `json:"user"`
	}
	decode(t, rec, &auth)
	if auth.Token == "" || auth.User.Name != "Alex Rivera" || auth.User.UnitNumber != "101" {
		t.Fatalf("unexpected auth payload: %+v", auth)
	}

	rec = doJSON(t, r, http.MethodGet, "/api/auth/me", auth.Token, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("me status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestBoardsPostsCommentsReactionsFeed(t *testing.T) {
	testutil.Setup(t)
	r := NewRouter()
	_, adminToken := seedAdmin(t)
	residentToken := inviteAndComplete(t, r, adminToken, "resident@test.local", "101", "Alex Rivera")

	rec := doJSON(t, r, http.MethodPost, "/api/boards", residentToken, map[string]string{
		"name": "Dog Lovers", "description": "All things canine",
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("create board status=%d body=%s", rec.Code, rec.Body.String())
	}
	var board struct {
		ID           string `json:"id"`
		IsSubscribed bool   `json:"is_subscribed"`
		Count        int64  `json:"subscriber_count"`
	}
	decode(t, rec, &board)
	if !board.IsSubscribed || board.Count != 1 {
		t.Fatalf("creator should be subscribed: %+v", board)
	}

	rec = doJSON(t, r, http.MethodPost, "/api/boards/"+board.ID+"/subscribe", residentToken, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("unsubscribe status=%d body=%s", rec.Code, rec.Body.String())
	}
	var sub struct {
		Subscribed bool `json:"subscribed"`
	}
	decode(t, rec, &sub)
	if sub.Subscribed {
		t.Fatal("expected unsubscribe toggle")
	}
	doJSON(t, r, http.MethodPost, "/api/boards/"+board.ID+"/subscribe", residentToken, nil)

	rec = doJSON(t, r, http.MethodPost, "/api/boards/"+board.ID+"/posts", residentToken, map[string]any{
		"title": "Anyone have a ladder?", "content": "Need it for an hour.",
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("create post status=%d body=%s", rec.Code, rec.Body.String())
	}
	var post struct{ ID string `json:"id"` }
	decode(t, rec, &post)

	rec = doJSON(t, r, http.MethodPost, "/api/boards/"+board.ID+"/posts", residentToken, map[string]any{
		"title": "Official", "content": "Nope", "post_type": "bulletin",
	})
	if rec.Code != http.StatusForbidden {
		t.Fatalf("resident bulletin should 403, got %d %s", rec.Code, rec.Body.String())
	}

	rec = doJSON(t, r, http.MethodPost, "/api/boards/"+board.ID+"/posts", adminToken, map[string]any{
		"title": "Pool closed Friday", "content": "Maintenance.", "post_type": "bulletin",
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("admin bulletin status=%d body=%s", rec.Code, rec.Body.String())
	}
	var bulletin struct{ ID string `json:"id"` }
	decode(t, rec, &bulletin)

	rec = doJSON(t, r, http.MethodGet, "/api/posts", residentToken, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("feed status=%d body=%s", rec.Code, rec.Body.String())
	}
	var feed struct {
		Posts []struct {
			ID       string `json:"id"`
			IsPinned bool   `json:"is_pinned"`
			Title    string `json:"title"`
		} `json:"posts"`
	}
	decode(t, rec, &feed)
	if len(feed.Posts) != 2 {
		t.Fatalf("expected 2 feed posts, got %d", len(feed.Posts))
	}
	if !feed.Posts[0].IsPinned || feed.Posts[0].ID != bulletin.ID {
		t.Fatalf("pinned bulletin should be first: %+v", feed.Posts)
	}

	rec = doJSON(t, r, http.MethodPost, "/api/posts/"+bulletin.ID+"/comments", residentToken, map[string]string{
		"content": "Should not work",
	})
	if rec.Code != http.StatusForbidden {
		t.Fatalf("bulletin comment should 403, got %d %s", rec.Code, rec.Body.String())
	}

	rec = doJSON(t, r, http.MethodPost, "/api/posts/"+post.ID+"/comments", residentToken, map[string]string{
		"content": "I have one you can use!",
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("comment status=%d body=%s", rec.Code, rec.Body.String())
	}

	rec = doJSON(t, r, http.MethodPut, "/api/posts/"+post.ID+"/reactions", residentToken, map[string]string{
		"type": "like",
	})
	if rec.Code != http.StatusNoContent {
		t.Fatalf("react status=%d body=%s", rec.Code, rec.Body.String())
	}

	rec = doJSON(t, r, http.MethodGet, "/api/posts/"+post.ID, residentToken, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("get post status=%d body=%s", rec.Code, rec.Body.String())
	}
	var detail struct {
		Comments  []struct{ Content string `json:"content"` } `json:"comments"`
		Reactions []struct {
			Type  string `json:"type"`
			Count int64  `json:"count"`
		} `json:"reactions"`
	}
	decode(t, rec, &detail)
	if len(detail.Comments) != 1 || detail.Comments[0].Content != "I have one you can use!" {
		t.Fatalf("unexpected comments: %+v", detail.Comments)
	}
	if len(detail.Reactions) != 1 || detail.Reactions[0].Type != "like" || detail.Reactions[0].Count != 1 {
		t.Fatalf("unexpected reactions: %+v", detail.Reactions)
	}
}

func TestDirectoryOptInAndOffboard(t *testing.T) {
	testutil.Setup(t)
	r := NewRouter()
	adminID, adminToken := seedAdmin(t)
	residentToken := inviteAndComplete(t, r, adminToken, "resident@test.local", "101", "Alex Rivera")

	rec := doJSON(t, r, http.MethodGet, "/api/directory", residentToken, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("directory status=%d body=%s", rec.Code, rec.Body.String())
	}
	var dir []map[string]any
	decode(t, rec, &dir)
	if len(dir) != 0 {
		t.Fatalf("expected empty directory, got %+v", dir)
	}

	optIn := true
	rec = doJSON(t, r, http.MethodPatch, "/api/profile/me", residentToken, map[string]any{
		"directory_opt_in": optIn,
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("profile patch status=%d body=%s", rec.Code, rec.Body.String())
	}

	rec = doJSON(t, r, http.MethodGet, "/api/directory", residentToken, nil)
	decode(t, rec, &dir)
	if len(dir) != 1 || dir[0]["name"] != "Alex Rivera" || dir[0]["unit_number"] != "101" {
		t.Fatalf("expected opted-in resident in directory: %+v", dir)
	}

	me := doJSON(t, r, http.MethodGet, "/api/auth/me", residentToken, nil)
	var meBody struct{ ID string `json:"id"` }
	decode(t, me, &meBody)

	rec = doJSON(t, r, http.MethodPost, "/api/admin/users/"+adminID+"/offboard", adminToken, nil)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("offboard admin should 403, got %d %s", rec.Code, rec.Body.String())
	}

	rec = doJSON(t, r, http.MethodPost, "/api/admin/users/"+meBody.ID+"/offboard", adminToken, nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("offboard status=%d body=%s", rec.Code, rec.Body.String())
	}

	rec = doJSON(t, r, http.MethodPost, "/api/auth/login", "", map[string]string{
		"email": "resident@test.local", "password": "secret-pass",
	})
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("offboarded login should fail, got %d %s", rec.Code, rec.Body.String())
	}
}

func TestFeedPagination(t *testing.T) {
	testutil.Setup(t)
	r := NewRouter()
	_, token := seedAdmin(t)

	rec := doJSON(t, r, http.MethodPost, "/api/boards", token, map[string]string{"name": "General"})
	var board struct{ ID string `json:"id"` }
	decode(t, rec, &board)

	for i := 0; i < 3; i++ {
		rec = doJSON(t, r, http.MethodPost, "/api/boards/"+board.ID+"/posts", token, map[string]string{
			"title":   "Post",
			"content": "Body",
		})
		if rec.Code != http.StatusCreated {
			t.Fatalf("seed post %d status=%d body=%s", i, rec.Code, rec.Body.String())
		}
	}

	rec = doJSON(t, r, http.MethodGet, "/api/posts?limit=2", token, nil)
	var page1 struct {
		Posts          []struct{ ID string `json:"id"` } `json:"posts"`
		NextPageCursor *string                          `json:"next_page_cursor"`
	}
	decode(t, rec, &page1)
	if len(page1.Posts) != 2 || page1.NextPageCursor == nil {
		t.Fatalf("expected 2 posts and a cursor, got %+v", page1)
	}

	rec = doJSON(t, r, http.MethodGet, "/api/posts?limit=2&cursor="+*page1.NextPageCursor, token, nil)
	var page2 struct {
		Posts          []struct{ ID string `json:"id"` } `json:"posts"`
		NextPageCursor *string                          `json:"next_page_cursor"`
	}
	decode(t, rec, &page2)
	if len(page2.Posts) != 1 {
		t.Fatalf("expected remaining 1 post, got %+v", page2)
	}
	if page1.Posts[0].ID == page2.Posts[0].ID || page1.Posts[1].ID == page2.Posts[0].ID {
		t.Fatalf("cursor page overlapped: %v then %v", page1.Posts, page2.Posts)
	}
}

func TestFeedRequiresAuth(t *testing.T) {
	testutil.Setup(t)
	rec := doJSON(t, NewRouter(), http.MethodGet, "/api/posts", "", nil)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestOffboardRevokesAccess(t *testing.T) {
	testutil.Setup(t)
	r := NewRouter()
	_, adminToken := seedAdmin(t)
	residentToken := inviteAndComplete(t, r, adminToken, "gone@test.local", "202", "Leaver")
	me := doJSON(t, r, http.MethodGet, "/api/auth/me", residentToken, nil)
	var meBody struct{ ID string `json:"id"` }
	decode(t, me, &meBody)
	if rec := doJSON(t, r, http.MethodPost, "/api/admin/users/"+meBody.ID+"/offboard", adminToken, nil); rec.Code != http.StatusNoContent {
		t.Fatalf("offboard %d %s", rec.Code, rec.Body.String())
	}
	if rec := doJSON(t, r, http.MethodGet, "/api/posts", residentToken, nil); rec.Code != http.StatusUnauthorized {
		t.Fatalf("offboarded token should 401, got %d %s", rec.Code, rec.Body.String())
	}
}

func TestMustSubscribeToPost(t *testing.T) {
	testutil.Setup(t)
	r := NewRouter()
	_, adminToken := seedAdmin(t)
	residentToken := inviteAndComplete(t, r, adminToken, "poster@test.local", "303", "Poster")
	rec := doJSON(t, r, http.MethodPost, "/api/boards", adminToken, map[string]string{"name": "Closed"})
	var board struct{ ID string `json:"id"` }
	decode(t, rec, &board)
	rec = doJSON(t, r, http.MethodPost, "/api/boards/"+board.ID+"/posts", residentToken, map[string]string{
		"title": "Nope", "content": "Should fail",
	})
	if rec.Code != http.StatusForbidden {
		t.Fatalf("unsubscribed post should 403, got %d %s", rec.Code, rec.Body.String())
	}
}

func TestRefreshLogoutAndPasswordRules(t *testing.T) {
	testutil.Setup(t)
	r := NewRouter()
	_, adminToken := seedAdmin(t)
	rec := doJSON(t, r, http.MethodPost, "/api/auth/register", adminToken, map[string]string{
		"email": "ref@test.local", "unit_number": "404",
	})
	var invited struct {
		Token    string `json:"registration_token"`
		Passcode string `json:"passcode"`
	}
	decode(t, rec, &invited)
	rec = doJSON(t, r, http.MethodPost, "/api/auth/complete-registration", "", map[string]string{
		"token": invited.Token, "passcode": invited.Passcode, "password": "short", "name": "Ref",
	})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("short password should 400, got %d %s", rec.Code, rec.Body.String())
	}
	rec = doJSON(t, r, http.MethodPost, "/api/auth/complete-registration", "", map[string]string{
		"token": invited.Token, "passcode": invited.Passcode, "password": "secret-pass", "name": "Ref",
	})
	var auth struct {
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}
	decode(t, rec, &auth)
	if auth.RefreshToken == "" {
		t.Fatal("expected refresh token")
	}
	rec = doJSON(t, r, http.MethodPost, "/api/auth/refresh", "", map[string]string{"refresh_token": auth.RefreshToken})
	if rec.Code != http.StatusOK {
		t.Fatalf("refresh %d %s", rec.Code, rec.Body.String())
	}
	var rotated struct {
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}
	decode(t, rec, &rotated)
	rec = doJSON(t, r, http.MethodPost, "/api/auth/logout", "", map[string]string{"refresh_token": rotated.RefreshToken})
	if rec.Code != http.StatusNoContent {
		t.Fatalf("logout %d %s", rec.Code, rec.Body.String())
	}
	rec = doJSON(t, r, http.MethodPost, "/api/auth/refresh", "", map[string]string{"refresh_token": rotated.RefreshToken})
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("revoked refresh should 401, got %d %s", rec.Code, rec.Body.String())
	}
}

func TestBoardGetAdminRosterMyReactionAndUrlGuard(t *testing.T) {
	testutil.Setup(t)
	r := NewRouter()
	_, adminToken := seedAdmin(t)
	residentToken := inviteAndComplete(t, r, adminToken, "react@test.local", "505", "Reactor")
	rec := doJSON(t, r, http.MethodPost, "/api/boards", residentToken, map[string]string{"name": "Yard"})
	var board struct{ ID string `json:"id"` }
	decode(t, rec, &board)
	rec = doJSON(t, r, http.MethodGet, "/api/boards/"+board.ID, residentToken, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("get board %d %s", rec.Code, rec.Body.String())
	}
	rec = doJSON(t, r, http.MethodPost, "/api/boards/"+board.ID+"/posts", residentToken, map[string]string{
		"title": "Hi", "content": "There",
	})
	var post struct{ ID string `json:"id"` }
	decode(t, rec, &post)
	doJSON(t, r, http.MethodPut, "/api/posts/"+post.ID+"/reactions", residentToken, map[string]string{"type": "like"})
	rec = doJSON(t, r, http.MethodGet, "/api/posts/"+post.ID, residentToken, nil)
	var detail struct {
		MyReaction string `json:"my_reaction"`
	}
	decode(t, rec, &detail)
	if detail.MyReaction != "like" {
		t.Fatalf("expected my_reaction like, got %+v", detail)
	}
	rec = doJSON(t, r, http.MethodGet, "/api/admin/users", adminToken, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("admin users %d %s", rec.Code, rec.Body.String())
	}
	rec = doJSON(t, r, http.MethodPatch, "/api/profile/me", residentToken, map[string]string{
		"profile_picture_url": "javascript:alert(1)",
	})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("javascript url should 400, got %d %s", rec.Code, rec.Body.String())
	}
}

func seedAdmin(t *testing.T) (id string, token string) {
	t.Helper()
	hashed, err := utils.HashPassword("admin-pass")
	if err != nil {
		t.Fatal(err)
	}
	u := &models.User{
		ID:             uuid.New(),
		Email:          "admin@test.local",
		Name:           "Business Admin",
		UnitNumber:     "Admin",
		HashedPassword: hashed,
		IsAdmin:        true,
		Status:         models.UserStatusActive,
	}
	if err := models.InsertUser(context.Background(), u); err != nil {
		t.Fatal(err)
	}
	tok, err := utils.GenerateAccessToken(u.ID.String(), u.UnitNumber, true)
	if err != nil {
		t.Fatal(err)
	}
	return u.ID.String(), tok
}

func inviteAndComplete(t *testing.T, r http.Handler, adminToken, email, unit, name string) string {
	t.Helper()
	rec := doJSON(t, r, http.MethodPost, "/api/auth/register", adminToken, map[string]string{
		"email": email, "unit_number": unit,
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("invite status=%d body=%s", rec.Code, rec.Body.String())
	}
	var invited struct {
		Token    string `json:"registration_token"`
		Passcode string `json:"passcode"`
	}
	decode(t, rec, &invited)
	rec = doJSON(t, r, http.MethodPost, "/api/auth/complete-registration", "", map[string]string{
		"token": invited.Token, "passcode": invited.Passcode, "password": "secret-pass", "name": name,
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("complete status=%d body=%s", rec.Code, rec.Body.String())
	}
	var auth struct {
		Token string `json:"token"`
	}
	decode(t, rec, &auth)
	return auth.Token
}

func doJSON(t *testing.T, h http.Handler, method, path, token string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatal(err)
		}
	}
	req := httptest.NewRequest(method, path, &buf)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func decode(t *testing.T, rec *httptest.ResponseRecorder, dest any) {
	t.Helper()
	if err := json.Unmarshal(rec.Body.Bytes(), dest); err != nil {
		t.Fatalf("decode %s: %v", rec.Body.String(), err)
	}
}
