package auth

import (
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
