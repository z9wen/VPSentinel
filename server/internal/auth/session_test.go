package auth

import (
	"testing"
	"time"
)

func TestSessionManagerIssuesSeparateAccessAndRefreshTokens(t *testing.T) {
	now := time.Date(2026, time.July, 25, 10, 0, 0, 0, time.UTC)
	manager := NewSessionManager("0123456789abcdef0123456789abcdef", 15*time.Minute, 7*24*time.Hour)
	pair, err := manager.IssuePair("admin@example.com", now)
	if err != nil {
		t.Fatalf("IssuePair: %v", err)
	}
	if pair.AccessToken == pair.RefreshToken {
		t.Fatal("access and refresh tokens must differ")
	}
	if email, err := manager.VerifyAccess(pair.AccessToken, now); err != nil || email != "admin@example.com" {
		t.Fatalf("VerifyAccess = %q, %v", email, err)
	}
	if email, err := manager.VerifyRefresh(pair.RefreshToken, now); err != nil || email != "admin@example.com" {
		t.Fatalf("VerifyRefresh = %q, %v", email, err)
	}
	if _, err := manager.VerifyAccess(pair.RefreshToken, now); err == nil {
		t.Fatal("refresh token was accepted as an access token")
	}
	if _, err := manager.VerifyAccess(pair.AccessToken, now.Add(16*time.Minute)); err == nil {
		t.Fatal("expired access token was accepted")
	}
}

func TestSessionManagerRejectsTamperedToken(t *testing.T) {
	now := time.Now().UTC()
	manager := NewSessionManager("0123456789abcdef0123456789abcdef", time.Minute, time.Hour)
	pair, err := manager.IssuePair("admin@example.com", now)
	if err != nil {
		t.Fatalf("IssuePair: %v", err)
	}
	tampered := pair.AccessToken[:len(pair.AccessToken)-1] + "x"
	if _, err := manager.VerifyAccess(tampered, now); err == nil {
		t.Fatal("tampered token was accepted")
	}
}
