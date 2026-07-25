package auth

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Postgres struct{ pool *pgxpool.Pool }

func NewPostgres(pool *pgxpool.Pool) *Postgres { return &Postgres{pool: pool} }

func HashToken(token string) []byte {
	sum := sha256.Sum256([]byte(token))
	return sum[:]
}

func (a *Postgres) Authenticate(ctx context.Context, token string) (string, error) {
	var nodeID string
	err := a.pool.QueryRow(ctx, `
		UPDATE agent_tokens SET last_used_at = NOW()
		WHERE token_hash = $1 AND revoked_at IS NULL
		RETURNING node_id
	`, HashToken(token)).Scan(&nodeID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrUnauthorized
	}
	if err != nil {
		return "", err
	}
	return nodeID, nil
}

type SessionManager struct {
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

type sessionClaims struct {
	Email string `json:"email"`
	Exp   int64  `json:"exp"`
	Type  string `json:"type"`
	JTI   string `json:"jti"`
}

type TokenPair struct {
	AccessToken      string    `json:"access_token"`
	RefreshToken     string    `json:"refresh_token"`
	TokenType        string    `json:"token_type"`
	AccessExpiresAt  time.Time `json:"access_expires_at"`
	RefreshExpiresAt time.Time `json:"refresh_expires_at"`
}

func NewSessionManager(secret string, accessTTL, refreshTTL time.Duration) *SessionManager {
	return &SessionManager{secret: []byte(secret), accessTTL: accessTTL, refreshTTL: refreshTTL}
}

func (m *SessionManager) IssuePair(email string, now time.Time) (TokenPair, error) {
	jti, err := randomTokenID()
	if err != nil {
		return TokenPair{}, err
	}
	accessExpiresAt := now.Add(m.accessTTL)
	refreshExpiresAt := now.Add(m.refreshTTL)
	access, err := m.issue(sessionClaims{Email: email, Exp: accessExpiresAt.Unix(), Type: "access"})
	if err != nil {
		return TokenPair{}, err
	}
	refresh, err := m.issue(sessionClaims{Email: email, Exp: refreshExpiresAt.Unix(), Type: "refresh", JTI: jti})
	if err != nil {
		return TokenPair{}, err
	}
	return TokenPair{AccessToken: access, RefreshToken: refresh, TokenType: "Bearer", AccessExpiresAt: accessExpiresAt, RefreshExpiresAt: refreshExpiresAt}, nil
}

func (m *SessionManager) issue(claims sessionClaims) (string, error) {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	unsigned := header + "." + base64.RawURLEncoding.EncodeToString(payload)
	return unsigned + "." + m.signature(unsigned), nil
}

func (m *SessionManager) VerifyAccess(token string, now time.Time) (string, error) {
	claims, err := m.verify(token, now, "access")
	return claims.Email, err
}

func (m *SessionManager) VerifyRefresh(token string, now time.Time) (string, error) {
	claims, err := m.verify(token, now, "refresh")
	return claims.Email, err
}

func (m *SessionManager) verify(token string, now time.Time, expectedType string) (sessionClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return sessionClaims{}, errors.New("invalid session")
	}
	unsigned := parts[0] + "." + parts[1]
	want := m.signature(unsigned)
	if !hmac.Equal([]byte(parts[2]), []byte(want)) {
		return sessionClaims{}, errors.New("invalid session")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return sessionClaims{}, errors.New("invalid session")
	}
	var claims sessionClaims
	if err := json.Unmarshal(payload, &claims); err != nil || claims.Email == "" || claims.Type != expectedType || now.Unix() >= claims.Exp {
		return sessionClaims{}, errors.New("expired or invalid session")
	}
	return claims, nil
}

func randomTokenID() (string, error) {
	value := make([]byte, 24)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(value), nil
}

func (m *SessionManager) signature(value string) string {
	mac := hmac.New(sha256.New, m.secret)
	_, _ = mac.Write([]byte(value))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
