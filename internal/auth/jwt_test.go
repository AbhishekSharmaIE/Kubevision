package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestJWTIssueAndValidate(t *testing.T) {
	j := &JWT{
		secret:        []byte("test-secret-at-least-32-characters-long"),
		accessTTL:     time.Hour,
		refreshMaxAge: 24 * time.Hour,
	}
	uid := uuid.New()
	tok, jti, err := j.IssueToken(uid, "a@b.co", true, []string{"team-1"})
	if err != nil || jti == "" {
		t.Fatalf("issue: %v jti=%q", err, jti)
	}
	cl, err := j.ValidateToken(tok)
	if err != nil {
		t.Fatal(err)
	}
	if cl.UserID != uid.String() || !cl.IsAdmin {
		t.Fatalf("claims: %+v", cl)
	}
}

func TestParseForRefreshAfterExp(t *testing.T) {
	j := &JWT{
		secret:        []byte("test-secret-at-least-32-characters-long"),
		accessTTL:     -time.Hour, // already expired
		refreshMaxAge: 48 * time.Hour,
	}
	uid := uuid.New()
	tok, _, err := j.IssueToken(uid, "a@b.co", false, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := j.ValidateToken(tok); err == nil {
		t.Fatal("expected expired")
	}
	cl, err := j.ParseForRefresh(tok)
	if err != nil {
		t.Fatal(err)
	}
	if cl.UserID != uid.String() {
		t.Fatalf("uid %v", cl.UserID)
	}
}
