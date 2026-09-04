package auth

import (
	"testing"
	"time"
)

func TestHashAndVerifyPassword(t *testing.T) {
	hash, err := HashPassword("testpassword123")
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}
	if hash == "" {
		t.Fatal("hash should not be empty")
	}
	if err := VerifyPassword(hash, "testpassword123"); err != nil {
		t.Errorf("VerifyPassword should succeed for correct password: %v", err)
	}
	if err := VerifyPassword(hash, "wrongpassword"); err == nil {
		t.Error("VerifyPassword should fail for wrong password")
	}
}

func TestHashPasswordTooShort(t *testing.T) {
	_, err := HashPassword("short")
	if err == nil {
		t.Error("HashPassword should reject passwords < 8 chars")
	}
}

func TestIssueAndVerifyAccessToken(t *testing.T) {
	svc := New("access-secret", "refresh-secret", 15*time.Minute, 168*time.Hour)
	token, exp, err := svc.IssueAccessToken(1, "analyst")
	if err != nil {
		t.Fatalf("IssueAccessToken failed: %v", err)
	}
	if token == "" {
		t.Fatal("token should not be empty")
	}
	if !exp.After(time.Now()) {
		t.Error("expiry should be in the future")
	}

	claims, err := svc.VerifyAccessToken(token)
	if err != nil {
		t.Fatalf("VerifyAccessToken failed: %v", err)
	}
	if claims.UserID != 1 {
		t.Errorf("expected UserID 1, got %d", claims.UserID)
	}
	if claims.Role != "analyst" {
		t.Errorf("expected role 'analyst', got %s", claims.Role)
	}
	if claims.Type != "access" {
		t.Errorf("expected type 'access', got %s", claims.Type)
	}
}

func TestIssueAndVerifyRefreshToken(t *testing.T) {
	svc := New("access-secret", "refresh-secret", 15*time.Minute, 168*time.Hour)
	token, _, err := svc.IssueRefreshToken(1, "viewer")
	if err != nil {
		t.Fatalf("IssueRefreshToken failed: %v", err)
	}
	claims, err := svc.VerifyRefreshToken(token)
	if err != nil {
		t.Fatalf("VerifyRefreshToken failed: %v", err)
	}
	if claims.Type != "refresh" {
		t.Errorf("expected type 'refresh', got %s", claims.Type)
	}
}

func TestAccessTokenRejectedAsRefresh(t *testing.T) {
	svc := New("access-secret", "refresh-secret", 15*time.Minute, 168*time.Hour)
	token, _, _ := svc.IssueAccessToken(1, "admin")
	if _, err := svc.VerifyRefreshToken(token); err == nil {
		t.Error("access token should not be accepted as refresh token")
	}
}

func TestRefreshTokenRejectedAsAccess(t *testing.T) {
	svc := New("access-secret", "refresh-secret", 15*time.Minute, 168*time.Hour)
	token, _, _ := svc.IssueRefreshToken(1, "admin")
	if _, err := svc.VerifyAccessToken(token); err == nil {
		t.Error("refresh token should not be accepted as access token")
	}
}

func TestTokenWithWrongSecret(t *testing.T) {
	svc1 := New("secret1", "refresh1", 15*time.Minute, 168*time.Hour)
	svc2 := New("secret2", "refresh2", 15*time.Minute, 168*time.Hour)
	token, _, _ := svc1.IssueAccessToken(1, "admin")
	if _, err := svc2.VerifyAccessToken(token); err == nil {
		t.Error("token signed with different secret should be rejected")
	}
}
