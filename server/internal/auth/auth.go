package auth

import (
	"context"
	"crypto/subtle"
	"errors"
)

var ErrUnauthorized = errors.New("unauthorized agent token")

type Authenticator interface {
	Authenticate(context.Context, string) (string, error)
}

// Static authenticates one development node. The interface is deliberately
// compatible with the later PostgreSQL-backed per-node token implementation.
type Static struct {
	token  []byte
	nodeID string
}

func NewStatic(token, nodeID string) *Static {
	return &Static{token: []byte(token), nodeID: nodeID}
}

func (a *Static) Authenticate(_ context.Context, token string) (string, error) {
	provided := []byte(token)
	if len(provided) != len(a.token) || subtle.ConstantTimeCompare(provided, a.token) != 1 {
		return "", ErrUnauthorized
	}
	return a.nodeID, nil
}
