package database

import (
	"context"
	"io"
	"log/slog"
	"os"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/z9wen/VPSentinel/server/internal/auth"
	"github.com/z9wen/VPSentinel/server/internal/management"
)

func TestPostgresManagementAndTokenLifecycle(t *testing.T) {
	databaseURL := os.Getenv("VPSENTINEL_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("VPSENTINEL_TEST_DATABASE_URL is not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	pool, err := Connect(ctx, databaseURL, slog.New(slog.NewJSONHandler(io.Discard, nil)))
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	defer pool.Close()
	if err := Migrate(ctx, pool, slog.Default()); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	store := NewPostgresStore(pool)
	const email = "management-test@example.com"
	const nodeID = "management-test-node"
	_, _ = pool.Exec(ctx, "DELETE FROM users WHERE email = $1", email)
	_, _ = pool.Exec(ctx, "DELETE FROM nodes WHERE id = $1", nodeID)
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cleanupCancel()
		_, _ = pool.Exec(cleanupCtx, "DELETE FROM users WHERE email = $1", email)
		_, _ = pool.Exec(cleanupCtx, "DELETE FROM nodes WHERE id = $1", nodeID)
	})

	if err := store.EnsureAdmin(ctx, email, "correct-password"); err != nil {
		t.Fatalf("EnsureAdmin: %v", err)
	}
	passwordHash, err := store.PasswordHash(ctx, email)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte("correct-password")) != nil {
		t.Fatalf("stored admin password is invalid: %v", err)
	}
	node, err := store.CreateNode(ctx, management.Node{ID: nodeID, Name: "Management test", Environment: "test"})
	if err != nil || node.Status != "pending" {
		t.Fatalf("CreateNode = %#v, %v", node, err)
	}
	const plainToken = "a-random-management-integration-token"
	token, err := store.CreateToken(ctx, nodeID, "integration", auth.HashToken(plainToken))
	if err != nil {
		t.Fatalf("CreateToken: %v", err)
	}
	authenticator := auth.NewPostgres(pool)
	if authenticatedNode, err := authenticator.Authenticate(ctx, plainToken); err != nil || authenticatedNode != nodeID {
		t.Fatalf("Authenticate = %q, %v", authenticatedNode, err)
	}
	if err := store.RevokeToken(ctx, nodeID, token.ID); err != nil {
		t.Fatalf("RevokeToken: %v", err)
	}
	if _, err := authenticator.Authenticate(ctx, plainToken); err == nil {
		t.Fatal("revoked token was accepted")
	}
}
