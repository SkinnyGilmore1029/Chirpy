package auth

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestMakeAndValidateJWT(t *testing.T) {
	secret := "super-secret-key"
	userID := uuid.New()

	// --- Case 1: Valid token ---
	token, err := MakeJWT(userID, secret, time.Minute)
	if err != nil {
		t.Fatalf("unexpected error creating token: %v", err)
	}

	parsedID, err := ValidateJWT(token, secret)
	if err != nil {
		t.Fatalf("unexpected error validating token: %v", err)
	}
	if parsedID != userID {
		t.Errorf("expected userID %v, got %v", userID, parsedID)
	}

	// --- Case 2: Expired token ---
	expiredToken, err := MakeJWT(userID, secret, -time.Minute)
	if err != nil {
		t.Fatalf("unexpected error creating expired token: %v", err)
	}

	_, err = ValidateJWT(expiredToken, secret)
	if err == nil {
		t.Error("expected error for expired token, got nil")
	}

	// --- Case 3: Wrong secret ---
	token2, err := MakeJWT(userID, secret, time.Minute)
	if err != nil {
		t.Fatalf("unexpected error creating token: %v", err)
	}

	_, err = ValidateJWT(token2, "wrong-secret")
	if err == nil {
		t.Error("expected error for token signed with wrong secret, got nil")
	}
}

func TestGetBearerToken(t *testing.T) {
	tests := []struct {
		name      string
		headers   http.Header
		wantToken string
		expectErr bool
	}{
		{
			name: "valid bearer token",
			headers: http.Header{
				"Authorization": {"Bearer abc123"},
			},
			wantToken: "abc123",
			expectErr: false,
		},
		{
			name:      "missing header",
			headers:   http.Header{},
			wantToken: "",
			expectErr: true,
		},
		{
			name: "wrong format",
			headers: http.Header{
				"Authorization": {"Token abc123"},
			},
			wantToken: "",
			expectErr: true,
		},
		{
			name: "empty bearer token",
			headers: http.Header{
				"Authorization": {"Bearer "},
			},
			wantToken: "",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotToken, err := GetBearerToken(tt.headers)
			if tt.expectErr && err == nil {
				t.Errorf("expected error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if gotToken != tt.wantToken {
				t.Errorf("expected token %q, got %q", tt.wantToken, gotToken)
			}
		})
	}
}
