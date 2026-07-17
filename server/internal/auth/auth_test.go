package auth

import (
	"context"
	"errors"
	"testing"
)

func TestStaticAuthenticate(t *testing.T) {
	authenticator := NewStatic("correct-token-123", "node-one")
	nodeID, err := authenticator.Authenticate(context.Background(), "correct-token-123")
	if err != nil {
		t.Fatalf("Authenticate: %v", err)
	}
	if nodeID != "node-one" {
		t.Fatalf("nodeID = %q, want node-one", nodeID)
	}
}

func TestStaticRejectsWrongToken(t *testing.T) {
	authenticator := NewStatic("correct-token-123", "node-one")
	_, err := authenticator.Authenticate(context.Background(), "wrong-token-1234")
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("Authenticate error = %v, want ErrUnauthorized", err)
	}
}
