package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cameronsralla/culdechat/internal/testutil"
	"github.com/cameronsralla/culdechat/models"
	"github.com/cameronsralla/culdechat/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestAdminRequiredDoesNotRunHandlerForResident(t *testing.T) {
	testutil.Setup(t)
	gin.SetMode(gin.TestMode)

	adminTok := tokenFor(t, true)
	residentTok := tokenFor(t, false)

	hits := 0
	r := gin.New()
	r.GET("/secret", AdminRequired(), func(c *gin.Context) {
		hits++
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/secret", nil)
	req.Header.Set("Authorization", "Bearer "+residentTok)
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("resident status=%d body=%s", rec.Code, rec.Body.String())
	}
	if hits != 0 {
		t.Fatalf("resident must not reach the handler, hits=%d", hits)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/secret", nil)
	req.Header.Set("Authorization", "Bearer "+adminTok)
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("admin status=%d body=%s", rec.Code, rec.Body.String())
	}
	if hits != 1 {
		t.Fatalf("admin should hit the handler once, hits=%d", hits)
	}
}

func tokenFor(t *testing.T, admin bool) string {
	t.Helper()
	hashed, err := utils.HashPassword("pass-word")
	if err != nil {
		t.Fatal(err)
	}
	email := "resident@auth.test"
	unit := "201"
	if admin {
		email = "admin@auth.test"
		unit = "A-1"
	}
	u := &models.User{
		ID:             uuid.New(),
		Email:          email,
		Name:           "Test",
		UnitNumber:     unit,
		HashedPassword: hashed,
		IsAdmin:        admin,
		Status:         models.UserStatusActive,
	}
	if err := models.InsertUser(context.Background(), u); err != nil {
		t.Fatal(err)
	}
	loaded, err := models.GetUserByID(context.Background(), u.ID)
	if err != nil || loaded == nil || loaded.Status != models.UserStatusActive {
		t.Fatalf("seed user missing id=%s err=%v loaded=%+v", u.ID, err, loaded)
	}
	tok, err := utils.GenerateAccessToken(u.ID.String(), u.UnitNumber, admin)
	if err != nil {
		t.Fatal(err)
	}
	return tok
}
