package auth

import (
	"errors"
	"testing"
	"time"
)

func TestGenerateToken(t *testing.T) {
	tests := []struct {
		name    string
		userID  string
		role    string
		ttl     time.Duration
		wantErr bool
	}{
		{
			name:   "TestGenerateToken_ValidClaims",
			userID: "user-123",
			role:   "admin",
			ttl:    time.Hour,
		},
		{
			name:   "TestGenerateToken_Expiry",
			userID: "user-exp",
			role:   "user",
			ttl:    time.Hour,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("JWT_SECRET", "jwt-test-secret")

			token, err := GenerateToken(tc.userID, tc.role, tc.ttl)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("generate token: %v", err)
			}

			claims, err := ValidateToken(token)
			if err != nil {
				t.Fatalf("validate generated token: %v", err)
			}

			if claims.UserID != tc.userID {
				t.Fatalf("unexpected user id: got %q want %q", claims.UserID, tc.userID)
			}
			if claims.Role != tc.role {
				t.Fatalf("unexpected role: got %q want %q", claims.Role, tc.role)
			}

			if tc.name == "TestGenerateToken_Expiry" {
				lowerBound := time.Now().UTC().Add(tc.ttl - 2*time.Second)
				upperBound := time.Now().UTC().Add(tc.ttl + 2*time.Second)
				if claims.ExpiresAt.Before(lowerBound) || claims.ExpiresAt.After(upperBound) {
					t.Fatalf("unexpected expiry window: got %s, expected between %s and %s", claims.ExpiresAt, lowerBound, upperBound)
				}
			}
		})
	}
}

func TestValidateToken(t *testing.T) {
	t.Setenv("JWT_SECRET", "jwt-test-secret")

	validToken, err := GenerateToken("valid-user", "user", time.Hour)
	if err != nil {
		t.Fatalf("generate valid token: %v", err)
	}
	expiredToken, err := GenerateToken("expired-user", "user", -1*time.Minute)
	if err != nil {
		t.Fatalf("generate expired token: %v", err)
	}

	tests := []struct {
		name      string
		token     string
		secret    string
		expectErr error
	}{
		{
			name:   "TestValidateToken_Valid",
			token:  validToken,
			secret: "jwt-test-secret",
		},
		{
			name:      "TestValidateToken_Expired",
			token:     expiredToken,
			secret:    "jwt-test-secret",
			expectErr: ErrTokenExpired,
		},
		{
			name:      "TestValidateToken_Tampered",
			token:     validToken + "tamper",
			secret:    "jwt-test-secret",
			expectErr: ErrTokenInvalid,
		},
		{
			name:      "TestValidateToken_WrongSecret",
			token:     validToken,
			secret:    "different-secret",
			expectErr: ErrTokenInvalid,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("JWT_SECRET", tc.secret)

			claims, err := ValidateToken(tc.token)
			if tc.expectErr != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tc.expectErr)
				}
				if !errors.Is(err, tc.expectErr) {
					t.Fatalf("unexpected error: got %v want %v", err, tc.expectErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("validate token: %v", err)
			}
			if claims.UserID != "valid-user" {
				t.Fatalf("unexpected user id: got %q want %q", claims.UserID, "valid-user")
			}
			if claims.Role != "user" {
				t.Fatalf("unexpected role: got %q want %q", claims.Role, "user")
			}
		})
	}
}

func TestGenerateToken_MissingJWTSecret(t *testing.T) {
	t.Setenv("JWT_SECRET", "")

	_, err := GenerateToken("user-1", "user", time.Hour)
	if err == nil {
		t.Fatal("expected error when JWT_SECRET is missing")
	}
	if !errors.Is(err, ErrTokenConfig) {
		t.Fatalf("unexpected error: got %v want %v", err, ErrTokenConfig)
	}
}
